package egraph

import (
	"fmt"
	"slices"

	"github.com/MilkeeyCat/expr_simplifier/ast"
	"github.com/MilkeeyCat/expr_simplifier/dsu"
)

type eclassID = dsu.Key

type eclass struct {
	nodes   []enode
	parents []enode
}

func concatEnodes(a, b []enode) []enode {
	for _, nodeB := range b {
		if !slices.ContainsFunc(a, func(nodeA enode) bool {
			return nodeA.Key() == nodeB.Key()
		}) {
			a = append(a, nodeB)
		}
	}

	return a
}

type enode interface {
	String(graph *Egraph) string
	Key() enodeKey
	Children() []eclassID
	CanonicalizeChildren(children []eclassID)
}

// Key invariant of this e-graph implementation is that after value saturation
// all e-class IDs used in e-nodes are canonical.
type Egraph struct {
	root          eclassID
	classesDSU    *dsu.DisjointSet
	classes       map[eclassID]*eclass
	interner      *stringInterner
	nodeToClassID map[enodeKey]eclassID
	worklist      []enode
}

func New(expr ast.Expr) *Egraph {
	graph := &Egraph{
		classesDSU:    new(dsu.DisjointSet),
		classes:       make(map[eclassID]*eclass),
		interner:      newStringInterner(),
		nodeToClassID: make(map[enodeKey]eclassID),
	}

	graph.root = translateExpr(graph, expr)

	return graph
}

func (graph *Egraph) eclass(id eclassID) *eclass {
	if value, ok := graph.classes[graph.classesDSU.Find(id)]; ok {
		return value
	} else {
		panic(fmt.Sprintf("e-class with id %d not found", id))
	}
}

func (graph *Egraph) canonicalEclassID(id eclassID) eclassID {
	return graph.classesDSU.Find(id)
}

func (graph *Egraph) merge(a, b eclassID) {
	a = graph.canonicalEclassID(a)
	b = graph.canonicalEclassID(b)

	if a == b {
		return
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
}

func (graph *Egraph) rebuild() {
	for len(graph.worklist) > 0 {
		node := graph.worklist[len(graph.worklist)-1]
		graph.worklist = graph.worklist[:len(graph.worklist)-1]
		key := node.Key()
		children := node.Children()

		for i, child := range children {
			children[i] = graph.canonicalEclassID(child)
		}

		node.CanonicalizeChildren(children)

		classID, ok := graph.nodeToClassID[key]
		if !ok {
			panic("e-class ID not found")
		}

		delete(graph.nodeToClassID, key)

		key = node.Key()

		if existing, ok := graph.nodeToClassID[key]; ok {
			graph.merge(existing, classID)
		} else {
			graph.nodeToClassID[key] = classID
		}
	}
}

type Visitor interface {
	VisitEnode(graph *Egraph, node enode)
	VisitEclass(graph *Egraph, classID eclassID)
}

func (graph *Egraph) DFS(visitor Visitor) {
	root := graph.canonicalEclassID(graph.root)
	visited := map[eclassID]struct{}{root: {}}
	var stack []eclassID = []eclassID{root}

	for len(stack) > 0 {
		class := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		visitor.VisitEclass(graph, class)

		for _, node := range graph.eclass(class).nodes {
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

func translateExpr(graph *Egraph, expr ast.Expr) eclassID {
	var node enode

	switch expr := expr.(type) {
	case *ast.BinaryExpr:
		node = &binaryEnode{
			op:  expr.Op,
			lhs: translateExpr(graph, expr.Lhs),
			rhs: translateExpr(graph, expr.Rhs),
		}
	case *ast.UnaryExpr:
		node = &unaryEnode{
			op:    expr.Op,
			class: translateExpr(graph, expr.Expr),
		}
	case *ast.IntExpr:
		node = &intEnode{
			value: expr.Value,
		}
	case *ast.VariableExpr:
		node = &variableEnode{
			name: graph.interner.intern(expr.Name),
		}
	default:
		panic(fmt.Sprintf("unknown expr type %T", expr))
	}

	key := node.Key()

	if classID, ok := graph.nodeToClassID[key]; ok {
		return graph.canonicalEclassID(classID)
	}

	classID := graph.classesDSU.Add()
	graph.classes[classID] = &eclass{nodes: []enode{node}}
	graph.nodeToClassID[key] = classID

	for _, child := range node.Children() {
		childClass := graph.eclass(child)

		// is this good enough or does the e-node key have to be used for
		// the comparison?
		if !slices.Contains(childClass.parents, node) {
			childClass.parents = append(childClass.parents, node)
		}
	}

	return classID
}
