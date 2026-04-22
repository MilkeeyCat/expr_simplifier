package parser

import (
	"errors"

	"github.com/MilkeeyCat/expr_simplifier/ast"
	"github.com/MilkeeyCat/expr_simplifier/lexer"
)

var precedences = map[lexer.TokenType]precedence{
	lexer.TokenTypePlus:      PrecedenceSum,
	lexer.TokenTypeMinus:     PrecedenceSum,
	lexer.TokenTypeAsterisk:  PrecedenceProduct,
	lexer.TokenTypeSlash:     PrecedenceProduct,
	lexer.TokenTypeLeftParen: PrecedenceCall,
}

type Parser struct {
	lexer          *lexer.Lexer
	curToken       lexer.Token
	peekToken      lexer.Token
	prefixParseFns map[lexer.TokenType]func() (ast.Expr, error)
	infixParseFns  map[lexer.TokenType]func(lhs ast.Expr) (ast.Expr, error)
}

func New(l *lexer.Lexer) (*Parser, error) {
	parser := &Parser{
		lexer: l,
	}

	parser.prefixParseFns = map[lexer.TokenType]func() (ast.Expr, error){
		lexer.TokenTypeIdent:     parser.parseIdent,
		lexer.TokenTypeInt:       parser.parseInt,
		lexer.TokenTypeMinus:     parser.parseUnaryExpr,
		lexer.TokenTypeLeftParen: parser.parseGroup,
	}

	parser.infixParseFns = map[lexer.TokenType]func(lhs ast.Expr) (ast.Expr, error){
		lexer.TokenTypePlus:     parser.parseBinaryExpr,
		lexer.TokenTypeMinus:    parser.parseBinaryExpr,
		lexer.TokenTypeAsterisk: parser.parseBinaryExpr,
		lexer.TokenTypeSlash:    parser.parseBinaryExpr,
	}

	if err := parser.nextToken(); err != nil {
		return nil, err
	}

	if err := parser.nextToken(); err != nil {
		return nil, err
	}

	return parser, nil
}

func (p *Parser) expect(tokenType lexer.TokenType) error {
	if p.curToken.Type != tokenType {
		return errors.New("unexpected token")
	}

	return p.nextToken()
}

func (p *Parser) nextToken() error {
	p.curToken = p.peekToken

	token, err := p.lexer.Next()
	if err != nil {
		return err
	}

	p.peekToken = token

	return nil
}

func (p *Parser) curPrecedence() precedence {
	if prec, ok := precedences[p.curToken.Type]; ok {
		return prec
	}

	return PrecedenceLowest
}

func (p *Parser) ParseExpr() (ast.Expr, error) {
	return p.parseExpr(PrecedenceLowest)
}

func (p *Parser) parseExpr(precedence precedence) (ast.Expr, error) {
	prefix, ok := p.prefixParseFns[p.curToken.Type]
	if !ok {
		return nil, errors.New("failed to parse prefix operator")
	}

	left, err := prefix()
	if err != nil {
		return nil, err
	}

	for precedence < p.curPrecedence() {
		infix, ok := p.infixParseFns[p.curToken.Type]
		if !ok {
			return nil, errors.New("failed to parse infix operator")
		}

		left, err = infix(left)
		if err != nil {
			return nil, err
		}
	}

	return left, nil
}

func (p *Parser) parseIdent() (ast.Expr, error) {
	expr := &ast.VariableExpr{
		Name: p.curToken.Value.(string),
	}

	if err := p.expect(lexer.TokenTypeIdent); err != nil {
		return nil, err
	}

	return expr, nil
}

func (p *Parser) parseInt() (ast.Expr, error) {
	expr := &ast.IntExpr{
		Value: p.curToken.Value.(int64),
	}

	if err := p.expect(lexer.TokenTypeInt); err != nil {
		return nil, err
	}

	return expr, nil
}

func (p *Parser) parseUnaryExpr() (ast.Expr, error) {
	var op ast.UnaryOp

	switch p.curToken.Type {
	case lexer.TokenTypeMinus:
		op = ast.UnaryOpNeg
	}

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	expr, err := p.parseExpr(PrecedencePrefix)
	if err != nil {
		return nil, err
	}

	return &ast.UnaryExpr{
		Op:   op,
		Expr: expr,
	}, nil

}

func (p *Parser) parseGroup() (ast.Expr, error) {
	if err := p.expect(lexer.TokenTypeLeftParen); err != nil {
		return nil, err
	}

	expr, err := p.ParseExpr()
	if err != nil {
		return nil, err
	}

	if err := p.expect(lexer.TokenTypeRightParen); err != nil {
		return nil, err
	}

	return expr, nil
}

func (p *Parser) parseBinaryExpr(lhs ast.Expr) (ast.Expr, error) {
	var op ast.BinaryOp

	switch p.curToken.Type {
	case lexer.TokenTypePlus:
		op = ast.BinaryOpAdd
	case lexer.TokenTypeMinus:
		op = ast.BinaryOpSub
	case lexer.TokenTypeAsterisk:
		op = ast.BinaryOpMul
	case lexer.TokenTypeSlash:
		op = ast.BinaryOpDiv
	}

	precedence := p.curPrecedence()

	if err := p.nextToken(); err != nil {
		return nil, err
	}

	rhs, err := p.parseExpr(precedence)
	if err != nil {
		return nil, err
	}

	return &ast.BinaryExpr{
		Op:  op,
		Lhs: lhs,
		Rhs: rhs,
	}, nil
}

func (p *Parser) ParseRewriteRule() (*ast.RewriteRule, error) {
	pat, err := p.ParseExpr()
	if err != nil {
		return nil, err
	}

	if err := p.expect(lexer.TokenTypeEqual); err != nil {
		return nil, err
	}

	if err := p.expect(lexer.TokenTypeGreaterThan); err != nil {
		return nil, err
	}

	result, err := p.ParseExpr()
	if err != nil {
		return nil, err
	}

	return &ast.RewriteRule{
		Pattern: pat,
		Result:  result,
	}, nil
}
