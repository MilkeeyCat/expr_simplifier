package lexer

type TokenType uint8

const (
	TokenTypeInvalid TokenType = iota
	TokenTypeEOF

	TokenTypeIdent
	TokenTypeInt

	TokenTypeComma
	TokenTypePlus
	TokenTypeMinus
	TokenTypeAsterisk
	TokenTypeSlash
	TokenTypeEqual
	TokenTypeGreaterThan

	TokenTypeLeftParen
	TokenTypeRightParen
)

type Token struct {
	Type  TokenType
	Value any
}
