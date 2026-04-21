package ast

import (
	"fmt"
	"strconv"
)

type Expr interface {
	fmt.Stringer
}

type BinaryOp uint8

func (op BinaryOp) String() string {
	switch op {
	case BinaryOpAdd:
		return "+"
	case BinaryOpSub:
		return "-"
	case BinaryOpMul:
		return "*"
	case BinaryOpDiv:
		return "/"
	default:
		panic("unknown binary operator")
	}
}

const (
	BinaryOpAdd BinaryOp = iota
	BinaryOpSub
	BinaryOpMul
	BinaryOpDiv
)

type BinaryExpr struct {
	Op  BinaryOp
	Lhs Expr
	Rhs Expr
}

func (expr *BinaryExpr) String() string {
	return "(" + expr.Lhs.String() + " " + expr.Op.String() + " " + expr.Rhs.String() + ")"
}

type UnaryOp uint8

func (op UnaryOp) String() string {
	switch op {
	case UnaryOpNeg:
		return "-"
	default:
		panic("unknown unary operator")
	}
}

const (
	UnaryOpNeg UnaryOp = iota
)

type UnaryExpr struct {
	Op   UnaryOp
	Expr Expr
}

func (expr *UnaryExpr) String() string {
	return expr.Op.String() + expr.Expr.String()
}

type IntExpr struct {
	Value int64
}

func (expr *IntExpr) String() string {
	return strconv.FormatInt(expr.Value, 10)
}

type VariableExpr struct {
	Name string
}

func (expr *VariableExpr) String() string {
	return expr.Name
}
