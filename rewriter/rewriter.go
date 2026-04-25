package rewriter

import (
	"fmt"

	"github.com/MilkeeyCat/expr_simplifier/ast"
	"github.com/MilkeeyCat/expr_simplifier/egraph"
)

type Stats struct {
	Iteration uint
}

type Stopper interface {
	Stop(stats Stats) bool
}

type Rewriter struct {
	graph   *egraph.Egraph
	rules   []*ast.RewriteRule
	stopper Stopper
}

func New(
	graph *egraph.Egraph,
	rules []*ast.RewriteRule,
	stopper Stopper,
) *Rewriter {
	return &Rewriter{
		graph:   graph,
		rules:   rules,
		stopper: stopper,
	}
}

func (r *Rewriter) Run() {
	for i := uint(0); ; i += 1 {
		changed := false

		for _, rule := range r.rules {
			matches := matchPattern(r.graph, rule.Pattern)

			if len(matches) > 0 {
				changed = applyMatches(r.graph, matches, rule.Result) || changed
			}
		}

		r.graph.Rebuild()

		stats := Stats{
			Iteration: i,
		}

		if r.stopper.Stop(stats) || !changed {
			break
		}
	}
}

func applyMatches(graph *egraph.Egraph, matches []match, expr ast.Expr) bool {
	changed := false

	for _, match := range matches {
		for _, env := range match.envs {
			classID := buildEclass(graph, env, expr)

			changed = changed || graph.Merge(match.classID, classID)
		}
	}

	return changed
}

func buildEclass(graph *egraph.Egraph, env env, expr ast.Expr) egraph.EclassID {
	switch expr := expr.(type) {
	case *ast.BinaryExpr:
		return graph.Add(&egraph.BinaryEnode{
			Op:  expr.Op,
			Lhs: buildEclass(graph, env, expr.Lhs),
			Rhs: buildEclass(graph, env, expr.Rhs),
		})

	case *ast.UnaryExpr:
		return graph.Add(&egraph.UnaryEnode{
			Op:      expr.Op,
			ClassID: buildEclass(graph, env, expr.Expr),
		})

	case *ast.IntExpr:
		return graph.Add(&egraph.IntEnode{
			Value: expr.Value,
		})

	case *ast.VariableExpr:
		classID, ok := env[expr.Name]
		if !ok {
			panic(fmt.Sprintf("variable %s not found", expr.Name))
		}

		return classID

	default:
		panic(fmt.Sprintf("unknown expr type %T", expr))
	}
}
