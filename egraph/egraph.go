package egraph

import (
	"fmt"

	"github.com/MilkeeyCat/expr_simplifier/ast"
	"github.com/MilkeeyCat/expr_simplifier/dsu"
)

type eclassID = dsu.Key

type eclass struct {
	nodes []enode
}

type enode interface {
	String(graph *Egraph) string
	Key() enodeKey
	Children() []eclassID
}

type Egraph struct {
	root          eclassID
	classesDSU    *dsu.DisjointSet
	classes       map[eclassID]*eclass
	interner      *stringInterner
	nodeToClassID map[enodeKey]eclassID
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
	return graph.classes[graph.classesDSU.Find(id)]
}

func (graph *Egraph) canonicalEclassID(id eclassID) eclassID {
	return graph.classesDSU.Find(id)
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
		return classID
	}

	classID := graph.classesDSU.Add()
	graph.classes[classID] = &eclass{nodes: []enode{node}}
	graph.nodeToClassID[key] = classID

	return classID
}
