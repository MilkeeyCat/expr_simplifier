package parser

type precedence uint8

const (
	PrecedenceLowest precedence = iota
	PrecedenceSum
	PrecedenceProduct
	PrecedencePrefix
	PrecedenceCall
)
