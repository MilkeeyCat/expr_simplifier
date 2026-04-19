package lexer_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MilkeeyCat/expr_simplifier/lexer"
)

func TestLexer(t *testing.T) {
	input := `
		foo
		69

		,
		+
		-
		*
		/

		(
		)
	`
	tokens := []lexer.Token{
		{lexer.TokenTypeIdent, "foo"},
		{lexer.TokenTypeInt, int64(69)},
		{lexer.TokenTypeComma, nil},
		{lexer.TokenTypePlus, nil},
		{lexer.TokenTypeMinus, nil},
		{lexer.TokenTypeAsterisk, nil},
		{lexer.TokenTypeSlash, nil},
		{lexer.TokenTypeLeftParen, nil},
		{lexer.TokenTypeRightParen, nil},
		{lexer.TokenTypeEOF, nil},
	}
	lexer := lexer.New(strings.NewReader(input))

	for _, expected := range tokens {
		actual, err := lexer.Next()

		assert.Nil(t, err)
		assert.Equal(t, expected, actual)
	}
}
