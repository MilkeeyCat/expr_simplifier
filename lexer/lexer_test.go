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
		=
		>

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
		{lexer.TokenTypeEqual, nil},
		{lexer.TokenTypeGreaterThan, nil},
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

func TestUnknownChars(t *testing.T) {
	results := []struct {
		token lexer.Token
		err   error
	}{
		{
			token: lexer.Token{lexer.TokenTypeIdent, "a"},
		},
		{
			token: lexer.Token{lexer.TokenTypeInvalid, nil},
			err:   lexer.ErrUnknownCharacter{Char: '&'},
		},
		{
			token: lexer.Token{lexer.TokenTypeInt, int64(5)},
		},
		{
			token: lexer.Token{lexer.TokenTypeEOF, nil},
		},
	}
	lexer := lexer.New(strings.NewReader("a & 5"))

	for _, result := range results {
		token, err := lexer.Next()

		assert.Equal(t, result.token, token)
		assert.Equal(t, result.err, err)
	}
}
