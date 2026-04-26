package parser_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MilkeeyCat/expr_simplifier/lexer"
	"github.com/MilkeeyCat/expr_simplifier/parser"
)

func TestSimpleExprs(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{
			input:  "a + 5",
			output: "(a + 5)",
		},
		{
			input:  "10 - 5",
			output: "(10 - 5)",
		},
		{
			input:  "8 * b",
			output: "(8 * b)",
		},
		{
			input:  "a / b",
			output: "(a / b)",
		},
		{
			input:  "-3 - 9",
			output: "(-3 - 9)",
		},
		{
			input:  "2 + -41",
			output: "(2 + -41)",
		},
		{
			input:  "pow(79, 2)",
			output: "pow(79, 2)",
		},
	}

	for _, tt := range tests {
		parser, err := parser.New(lexer.New(strings.NewReader(tt.input)))
		assert.Nil(t, err)

		expr, err := parser.ParseExpr()
		assert.Nil(t, err)
		assert.Equal(t, tt.output, expr.String())
	}
}

func TestPrecedence(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{
			input:  "1 + 2 + 3",
			output: "((1 + 2) + 3)",
		},
		{
			input:  "1 + (2 + 3)",
			output: "(1 + (2 + 3))",
		},
		{
			input:  "1 + 2 * 3",
			output: "(1 + (2 * 3))",
		},
		{
			input:  "1 * 2 + 3 / 4",
			output: "((1 * 2) + (3 / 4))",
		},
		{
			input:  "sqrt(16) + 9 / 26",
			output: "(sqrt(16) + (9 / 26))",
		},
	}

	for _, tt := range tests {
		parser, err := parser.New(lexer.New(strings.NewReader(tt.input)))
		assert.Nil(t, err)

		expr, err := parser.ParseExpr()
		assert.Nil(t, err)
		assert.Equal(t, tt.output, expr.String())
	}
}

func TestRewriteRules(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{
		{
			input:  "a + 0 => 0",
			output: "(a + 0) => 0",
		},
		{
			input:  "a * 1 => a",
			output: "(a * 1) => a",
		},
		{
			input:  "a - a => 0",
			output: "(a - a) => 0",
		},
		{
			input:  "pow(sqrt(a), 2) => a",
			output: "pow(sqrt(a), 2) => a",
		},
	}

	for _, tt := range tests {
		parser, err := parser.New(lexer.New(strings.NewReader(tt.input)))
		assert.Nil(t, err)

		rewriteRule, err := parser.ParseRewriteRule()
		assert.Nil(t, err)
		assert.Equal(t, tt.output, rewriteRule.String())
	}
}
