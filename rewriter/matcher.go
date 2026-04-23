package rewriter

import (
	"maps"

	"github.com/MilkeeyCat/expr_simplifier/ast"
	"github.com/MilkeeyCat/expr_simplifier/egraph"
)

type env map[string]egraph.EclassID

type match struct {
	envs    []env
	classID egraph.EclassID
}

type matcher struct {
	pattern ast.Expr
	matches []match
}

func (m *matcher) VisitEclass(graph *egraph.Egraph, classID egraph.EclassID) {
	envs := m.match(graph, classID, m.pattern, make(env))

	if len(envs) > 0 {
		m.matches = append(m.matches, match{
			envs:    envs,
			classID: classID,
		})
	}
}

func (m *matcher) VisitEnode(graph *egraph.Egraph, node egraph.Enode) {}

func (m *matcher) match(
	graph *egraph.Egraph,
	classID egraph.EclassID,
	pat ast.Expr,
	bindingsEnv env,
) []env {
	class := graph.Eclass(classID)

	switch expr := pat.(type) {
	case *ast.BinaryExpr:
		var envs []env

		for _, node := range class.Nodes() {
			if node, ok := node.(*egraph.BinaryEnode); ok {
				if node.Op != expr.Op {
					continue
				}

				left := m.match(graph, node.Lhs, expr.Lhs, bindingsEnv)

				for _, env := range left {
					envs = append(envs, m.match(graph, node.Rhs, expr.Rhs, env)...)
				}
			}
		}

		return envs

	case *ast.UnaryExpr:
		var envs []env

		for _, node := range class.Nodes() {
			if node, ok := node.(*egraph.UnaryEnode); ok {
				if node.Op != expr.Op {
					continue
				}

				envs = append(envs, m.match(graph, node.ClassID, expr.Expr, bindingsEnv)...)
			}
		}

		return envs

	case *ast.IntExpr:
		for _, node := range class.Nodes() {
			if node, ok := node.(*egraph.IntEnode); ok {
				if node.Value == expr.Value {
					return []env{bindingsEnv}
				}
			}
		}

	case *ast.VariableExpr:
		if old, ok := bindingsEnv[expr.Name]; ok {
			if classID == old {
				return []env{bindingsEnv}
			}

			return nil
		}

		bindingsEnv = maps.Clone(bindingsEnv)
		bindingsEnv[expr.Name] = classID

		return []env{bindingsEnv}
	}

	return nil
}

func matchPattern(graph *egraph.Egraph, pat ast.Expr) []match {
	matcher := &matcher{
		pattern: pat,
	}

	graph.DFS(matcher)

	return matcher.matches
}
