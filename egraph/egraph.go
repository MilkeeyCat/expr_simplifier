package egraph

import (
	"fmt"
	"math"
	"slices"

	"github.com/MilkeeyCat/expr_simplifier/ast"
	"github.com/MilkeeyCat/expr_simplifier/dsu"
)

type EclassID = dsu.Key

type Eclass struct {
	nodes   []Enode
	parents []Enode
}

func (e *Eclass) Nodes() []Enode {
	return e.nodes
}

func concatEnodes(a, b []Enode) []Enode {
	for _, nodeB := range b {
		if !slices.ContainsFunc(a, func(nodeA Enode) bool {
			return nodeA.Key() == nodeB.Key()
		}) {
			a = append(a, nodeB)
		}
	}

	return a
}

type Enode interface {
	String(graph *Egraph) string
	Key() enodeKey
	Children() []EclassID
	canonicalizeChildren(children []EclassID)
}

// Key invariant of this e-graph implementation is that after value saturation
// all e-class IDs used in e-nodes are canonical.
type Egraph struct {
	root          EclassID
	classesDSU    *dsu.DisjointSet
	classes       map[EclassID]*Eclass
	interner      *stringInterner
	nodeToClassID map[enodeKey]EclassID
	worklist      []Enode
}

func New(expr ast.Expr) *Egraph {
	graph := &Egraph{
		classesDSU:    new(dsu.DisjointSet),
		classes:       make(map[EclassID]*Eclass),
		interner:      newStringInterner(),
		nodeToClassID: make(map[enodeKey]EclassID),
	}

	graph.root = translateExpr(graph, expr)

	return graph
}

func (graph *Egraph) Eclass(id EclassID) *Eclass {
	if value, ok := graph.classes[graph.classesDSU.Find(id)]; ok {
		return value
	} else {
		panic(fmt.Sprintf("e-class with id %d not found", id))
	}
}

func (graph *Egraph) canonicalEclassID(id EclassID) EclassID {
	return graph.classesDSU.Find(id)
}

func (graph *Egraph) Merge(a, b EclassID) bool {
	a = graph.canonicalEclassID(a)
	b = graph.canonicalEclassID(b)

	if a == b {
		return false
	}

	graph.classesDSU.Union(a, b)

	root := graph.classesDSU.Find(a)
	rootClass, ok := graph.classes[root]
	if !ok {
		panic("e-class not found")
	}

	other := a
	if root == a {
		other = b
	}

	otherClass, ok := graph.classes[other]
	if !ok {
		panic("e-class not found")
	}

	rootClass.nodes = concatEnodes(rootClass.nodes, otherClass.nodes)
	rootClass.parents = concatEnodes(rootClass.parents, otherClass.parents)
	graph.worklist = concatEnodes(graph.worklist, otherClass.parents)
	delete(graph.classes, other)

	return true
}

func (graph *Egraph) Rebuild() {
	for len(graph.worklist) > 0 {
		node := graph.worklist[len(graph.worklist)-1]
		graph.worklist = graph.worklist[:len(graph.worklist)-1]
		key := node.Key()
		children := node.Children()

		for i, child := range children {
			children[i] = graph.canonicalEclassID(child)
		}

		node.canonicalizeChildren(children)

		classID, ok := graph.nodeToClassID[key]
		if !ok {
			panic("e-class ID not found")
		}

		// deduplicate e-class's nodes. After canonicalizing the e-node's
		// children, it may become identical to node(-s?) already present in the
		// e-class.
		//
		{
			class := graph.Eclass(classID)
			nodes := make([]Enode, 0, len(class.nodes))
			keys := make(map[enodeKey]struct{}, len(class.nodes))

			for _, node := range class.nodes {
				key := node.Key()

				if _, ok := keys[key]; !ok {
					nodes = append(nodes, node)
					keys[key] = struct{}{}
				}
			}

			class.nodes = nodes
		}

		delete(graph.nodeToClassID, key)

		key = node.Key()

		if existing, ok := graph.nodeToClassID[key]; ok {
			graph.Merge(existing, classID)
		} else {
			graph.nodeToClassID[key] = classID
		}
	}
}

type Visitor interface {
	VisitEnode(graph *Egraph, node Enode)
	VisitEclass(graph *Egraph, classID EclassID)
}

func (graph *Egraph) DFS(visitor Visitor) {
	root := graph.canonicalEclassID(graph.root)
	visited := map[EclassID]struct{}{root: {}}
	stack := []EclassID{root}

	for len(stack) > 0 {
		class := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		visitor.VisitEclass(graph, class)

		for _, node := range graph.Eclass(class).nodes {
			visitor.VisitEnode(graph, node)

			for _, child := range node.Children() {
				child = graph.canonicalEclassID(child)

				if _, ok := visited[child]; !ok {
					stack = append(stack, child)
					visited[child] = struct{}{}
				}
			}
		}
	}
}

func (graph *Egraph) Add(node Enode) EclassID {
	key := node.Key()

	if classID, ok := graph.nodeToClassID[key]; ok {
		return graph.canonicalEclassID(classID)
	}

	classID := graph.classesDSU.Add()
	graph.classes[classID] = &Eclass{nodes: []Enode{node}}
	graph.nodeToClassID[key] = classID

	for _, child := range node.Children() {
		childClass := graph.Eclass(child)

		// is this good enough or does the e-node key have to be used for
		// the comparison?
		if !slices.Contains(childClass.parents, node) {
			childClass.parents = append(childClass.parents, node)
		}
	}

	return classID
}

type Cost = uint
type CostFunc func(node Enode, costs map[EclassID]Cost) Cost

const MaxCost Cost = math.MaxUint

func (graph *Egraph) Extract(costFunc CostFunc) ast.Expr {
	costs := make(map[EclassID]Cost)
	best := make(map[EclassID]Enode)
	var worklist []EclassID

	for id, class := range graph.classes {
		isLeaf := true

		for _, node := range class.Nodes() {
			if len(node.Children()) > 0 {
				isLeaf = false
			}
		}

		if isLeaf {
			worklist = append(worklist, id)
		}

		costs[id] = MaxCost
	}

	for len(worklist) > 0 {
		classID := worklist[len(worklist)-1]
		class := graph.classes[classID]
		worklist = worklist[:len(worklist)-1]
		bestCost := MaxCost
		var bestNode Enode

		for _, node := range class.Nodes() {
			cost := costFunc(node, costs)

			if cost < bestCost {
				bestNode = node
				bestCost = cost
			}
		}

		if bestCost < costs[classID] {
			costs[classID] = bestCost
			best[classID] = bestNode

			var classIDs []EclassID

			for _, node := range class.parents {
				classID := graph.canonicalEclassID(graph.nodeToClassID[node.Key()])

				if !slices.Contains(classIDs, classID) {
					classIDs = append(classIDs, classID)
				}
			}

			worklist = append(worklist, classIDs...)
		}
	}

	return buildExpr(graph, best, graph.canonicalEclassID(graph.root))
}

func translateExpr(graph *Egraph, expr ast.Expr) EclassID {
	var node Enode

	switch expr := expr.(type) {
	case *ast.BinaryExpr:
		node = &BinaryEnode{
			Op:  expr.Op,
			Lhs: translateExpr(graph, expr.Lhs),
			Rhs: translateExpr(graph, expr.Rhs),
		}
	case *ast.UnaryExpr:
		node = &UnaryEnode{
			Op:      expr.Op,
			ClassID: translateExpr(graph, expr.Expr),
		}
	case *ast.IntExpr:
		node = &IntEnode{
			Value: expr.Value,
		}
	case *ast.VariableExpr:
		node = &VariableEnode{
			Name: graph.interner.intern(expr.Name),
		}
	default:
		panic(fmt.Sprintf("unknown expr type %T", expr))
	}

	return graph.Add(node)
}

func buildExpr(graph *Egraph, nodes map[EclassID]Enode, classID EclassID) ast.Expr {
	switch node := nodes[classID].(type) {
	case *BinaryEnode:
		return &ast.BinaryExpr{
			Op:  node.Op,
			Lhs: buildExpr(graph, nodes, node.Lhs),
			Rhs: buildExpr(graph, nodes, node.Rhs),
		}
	case *UnaryEnode:
		return &ast.UnaryExpr{
			Op:   node.Op,
			Expr: buildExpr(graph, nodes, node.ClassID),
		}
	case *IntEnode:
		return &ast.IntExpr{
			Value: node.Value,
		}
	case *VariableEnode:
		return &ast.VariableExpr{
			Name: graph.interner.get(node.Name),
		}
	default:
		panic(fmt.Sprintf("unknown e-node type %T", node))
	}
}
