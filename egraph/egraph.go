package egraph

import (
	"fmt"

	"github.com/MilkeeyCat/expr_simplifier/ast"
)

type eclass struct {
	nodes []enode
}

type enode interface {
	String(graph *Egraph) string
	Key() enodeKey
	Children() []*eclass
}

type Egraph struct {
	root     *eclass
	interner *stringInterner
	classes  map[enodeKey]*eclass
}

func New(expr ast.Expr) *Egraph {
	graph := &Egraph{
		interner: newStringInterner(),
		classes:  make(map[enodeKey]*eclass),
	}

	graph.root = translateExpr(graph, expr)

	return graph
}

type Visitor interface {
	VisitEnode(node enode)
	VisitEclass(class *eclass)
}

func (graph *Egraph) DFS(visitor Visitor) {
	visited := make(map[*eclass]struct{})
	var stack []*eclass = []*eclass{graph.root}

	for len(stack) > 0 {
		class := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		visitor.VisitEclass(class)

		for _, node := range class.nodes {
			visitor.VisitEnode(node)

			for _, child := range node.Children() {
				if _, ok := visited[child]; !ok {
					stack = append(stack, child)
					visited[child] = struct{}{}
				}
			}
		}
	}
}

func translateExpr(graph *Egraph, expr ast.Expr) *eclass {
	var node enode

	switch expr := expr.(type) {
	case *ast.ExprBinary:
		node = &binaryEnode{
			op:  expr.Op,
			lhs: translateExpr(graph, expr.Lhs),
			rhs: translateExpr(graph, expr.Rhs),
		}
	case *ast.ExprUnary:
		node = &unaryEnode{
			op:    expr.Op,
			class: translateExpr(graph, expr.Expr),
		}
	case *ast.ExprInt:
		node = &intEnode{
			value: expr.Value,
		}
	case *ast.ExprVariable:
		node = &variableEnode{
			name: graph.interner.intern(expr.Name),
		}
	default:
		panic(fmt.Sprintf("unknown expr type %T", expr))
	}

	key := node.Key()

	if class, ok := graph.classes[key]; ok {
		return class
	}

	class := &eclass{
		nodes: []enode{node},
	}
	graph.classes[key] = class

	return class
}
