package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strings"

	"github.com/MilkeeyCat/expr_simplifier/ast"
	"github.com/MilkeeyCat/expr_simplifier/egraph"
	"github.com/MilkeeyCat/expr_simplifier/lexer"
	"github.com/MilkeeyCat/expr_simplifier/parser"
	"github.com/MilkeeyCat/expr_simplifier/rewriter"
)

func main() {
	if err := main_(); err != nil {
		fmt.Println(err)

		os.Exit(1)
	}
}

func main_() error {
	fs := flag.NewFlagSet("expr_simplifier", flag.ExitOnError)
	visualize := fs.Bool("viz", false, "Produce a visualization of resulting e-graph(default false)")
	rulesPath := fs.String("rules", "", "Path to rewrite rules")
	out := fs.String("out", "", "The path where to put visualization file(default current directory)")

	fs.SetOutput(os.Stdout)

	if err := fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	if *rulesPath == "" {
		return errors.New("provide the -rules flag")
	}

	args := fs.Args()

	if len(args) != 1 {
		return errors.New("the last argument is expected to be a maffs expression")
	}

	parser, err := parser.New(lexer.New(strings.NewReader(args[0])))
	if err != nil {
		return fmt.Errorf("failed to lex the input expression: %w", err)
	}

	expr, err := parser.ParseExpr()
	if err != nil {
		return fmt.Errorf("failed to parse the input expression: %w", err)
	}

	rules, err := os.ReadFile(*rulesPath)
	if err != nil {
		return err
	}

	rewriteRules, err := parseRewriteRules(strings.Lines(string(rules)))
	if err != nil {
		return err
	}

	graph := egraph.New(expr)
	rewriter := rewriter.New(graph, rewriteRules, nopStopper{})

	rewriter.Run()

	if *visualize {
		file, err := os.Create(filepath.Join(*out, "visualization.svg"))
		if err != nil {
			return err
		}

		if err := egraph.Visualize(context.Background(), graph, file); err != nil {
			return err
		}
	}

	fmt.Println("Original expression: " + expr.String())

	expr = graph.Extract(cost)

	fmt.Println("Simplified expression: " + expr.String())

	return nil
}

func parseRewriteRules(rules iter.Seq[string]) ([]*ast.RewriteRule, error) {
	var rewriteRules []*ast.RewriteRule

	for rule := range rules {
		parser, err := parser.New(lexer.New(strings.NewReader(rule)))
		if err != nil {
			return nil, fmt.Errorf("failed to lex rewrite rule: %w", err)
		}

		rr, err := parser.ParseRewriteRule()
		if err != nil {
			return nil, fmt.Errorf("failed to parse rewrite rule: %w", err)
		}

		rewriteRules = append(rewriteRules, rr)
	}

	return rewriteRules, nil
}

// runs the saturation run until there're no rewrites that can be applied
type nopStopper struct{}

func (ns nopStopper) Stop(stats rewriter.Stats) bool {
	return false
}

func cost(node egraph.Enode, costs map[egraph.EclassID]egraph.Cost) egraph.Cost {
	switch node := node.(type) {
	case *egraph.BinaryEnode:
		if costs[node.Lhs] == egraph.MaxCost || costs[node.Rhs] == egraph.MaxCost {
			return egraph.MaxCost
		}

		return 1 + costs[node.Lhs] + costs[node.Rhs]
	case *egraph.UnaryEnode:
		if costs[node.ClassID] == egraph.MaxCost {
			return egraph.MaxCost
		}

		return 1 + costs[node.ClassID]
	case *egraph.IntEnode:
		return 1
	case *egraph.VariableEnode:
		return 1
	default:
		panic(fmt.Sprintf("unknown e-node type %T", node))
	}
}
