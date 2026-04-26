package egraph

import (
	"fmt"
	"hash/maphash"
	"math"
	"slices"

	"github.com/MilkeeyCat/expr_simplifier/ast"
	"github.com/MilkeeyCat/expr_simplifier/dsu"
	"github.com/MilkeeyCat/expr_simplifier/hashmap"
)

const hashMapSize = 10

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
			return cmpEnodes(nodeA, nodeB)
		}) {
			a = append(a, nodeB)
		}
	}

	return a
}

func cmpEnodes(a, b Enode) bool {
	switch a := a.(type) {
	case *BinaryEnode:
		if b, ok := b.(*BinaryEnode); ok {
			return a.Op == b.Op && a.Lhs == b.Lhs && a.Rhs == b.Rhs
		}
	case *UnaryEnode:
		if b, ok := b.(*UnaryEnode); ok {
			return a.Op == b.Op && a.ClassID == b.ClassID
		}
	case *CallEnode:
		if b, ok := b.(*CallEnode); ok {
			return a.Name == b.Name && slices.Compare(a.Args, b.Args) == 0
		}
	case *IntEnode:
		if b, ok := b.(*IntEnode); ok {
			return a.Value == b.Value
		}
	case *VariableEnode:
		if b, ok := b.(*VariableEnode); ok {
			return a.Name == b.Name
		}
	default:
		panic(fmt.Sprintf("unknown e-node type %T", a))
	}

	return false
}

type Enode interface {
	fmt.Stringer

	Children() []EclassID
	FillHash(hash *maphash.Hash)
	canonicalizeChildren(children []EclassID)
}

// Key invariant of this e-graph implementation is that after value saturation
// all e-class IDs used in e-nodes are canonical.
type Egraph struct {
	root          EclassID
	classesDSU    *dsu.DisjointSet
	classes       map[EclassID]*Eclass
	nodeToClassID *hashmap.HashMap[Enode, EclassID]
	worklist      []Enode
}

func New(expr ast.Expr) *Egraph {
	graph := &Egraph{
		classesDSU:    new(dsu.DisjointSet),
		classes:       make(map[EclassID]*Eclass),
		nodeToClassID: hashmap.New[Enode, EclassID](hashMapSize, cmpEnodes),
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
		classID, ok := graph.nodeToClassID.Get(node)
		if !ok {
			panic("e-class ID not found")
		}
		children := node.Children()

		for i, child := range children {
			children[i] = graph.canonicalEclassID(child)
		}

		node.canonicalizeChildren(children)

		// deduplicate e-class's nodes. After canonicalizing the e-node's
		// children, it may become identical to node(-s?) already present in the
		// e-class.
		{
			class := graph.Eclass(classID)
			nodes := make([]Enode, 0, len(class.nodes))
			visited := hashmap.New[Enode, struct{}](hashMapSize, cmpEnodes)

			for _, node := range class.nodes {
				if _, ok := visited.Get(node); !ok {
					nodes = append(nodes, node)
					visited.Insert(node, struct{}{})
				}
			}

			class.nodes = nodes
		}

		graph.nodeToClassID.Remove(node)

		if existing, ok := graph.nodeToClassID.Get(node); ok {
			graph.Merge(existing, classID)
		} else {
			graph.nodeToClassID.Insert(node, classID)
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
	if classID, ok := graph.nodeToClassID.Get(node); ok {
		return graph.canonicalEclassID(classID)
	}

	classID := graph.classesDSU.Add()
	graph.classes[classID] = &Eclass{nodes: []Enode{node}}
	graph.nodeToClassID.Insert(node, classID)

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
				classID, ok := graph.nodeToClassID.Get(node)
				if !ok {
					panic("e-class not found")
				}
				classID = graph.canonicalEclassID(classID)

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
	case *ast.CallExpr:
		args := make([]EclassID, len(expr.Args))

		for i, expr := range expr.Args {
			args[i] = translateExpr(graph, expr)
		}

		node = &CallEnode{
			Name: expr.Name,
			Args: args,
		}
	case *ast.IntExpr:
		node = &IntEnode{
			Value: expr.Value,
		}
	case *ast.VariableExpr:
		node = &VariableEnode{
			Name: expr.Name,
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
	case *CallEnode:
		args := make([]ast.Expr, len(node.Args))

		for i, classID := range node.Args {
			args[i] = buildExpr(graph, nodes, classID)
		}

		return &ast.CallExpr{
			Name: node.Name,
			Args: args,
		}
	case *IntEnode:
		return &ast.IntExpr{
			Value: node.Value,
		}
	case *VariableEnode:
		return &ast.VariableExpr{
			Name: node.Name,
		}
	default:
		panic(fmt.Sprintf("unknown e-node type %T", node))
	}
}
