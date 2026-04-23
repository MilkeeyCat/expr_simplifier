package egraph

import (
	"fmt"
	"strconv"
	"unsafe"

	"github.com/MilkeeyCat/expr_simplifier/ast"
)

type enodeKeyKind byte

const (
	enodeKeyKindBinary enodeKeyKind = iota
	enodeKeyKindUnary
	enodeKeyKindInt
	enodeKeyKindVariable
)

const wordSize = unsafe.Sizeof(uintptr(0))

type enodeKey struct {
	kind enodeKeyKind
	data [wordSize * 2]byte
}

func enodeIDToBytes(id EclassID) [wordSize]byte {
	return *(*[unsafe.Sizeof(EclassID(0))]byte)(unsafe.Pointer(&id))
}

type BinaryEnode struct {
	Op  ast.BinaryOp
	Lhs EclassID
	Rhs EclassID
}

func (node *BinaryEnode) String(graph *Egraph) string {
	return node.Op.String()
}

func (node *BinaryEnode) Key() enodeKey {
	var data [wordSize * 2]byte
	lhs := enodeIDToBytes(node.Lhs)
	rhs := enodeIDToBytes(node.Rhs)

	copy(data[:], lhs[:])
	copy(data[wordSize:], rhs[:])

	return enodeKey{
		kind: enodeKeyKindBinary,
		data: data,
	}
}

func (node *BinaryEnode) Children() []EclassID {
	return []EclassID{node.Lhs, node.Rhs}
}

func (node *BinaryEnode) canonicalizeChildren(children []EclassID) {
	if len(children) != 2 {
		panic(fmt.Sprintf("expected 2 children, got %d", len(children)))
	}

	node.Lhs = children[0]
	node.Rhs = children[1]
}

type UnaryEnode struct {
	Op      ast.UnaryOp
	ClassID EclassID
}

func (node *UnaryEnode) String(graph *Egraph) string {
	return node.Op.String()
}

func (node *UnaryEnode) Key() enodeKey {
	var data [wordSize * 2]byte
	expr := enodeIDToBytes(node.ClassID)

	copy(data[:], expr[:])

	return enodeKey{
		kind: enodeKeyKindUnary,
		data: data,
	}
}

func (node *UnaryEnode) Children() []EclassID {
	return []EclassID{node.ClassID}
}

func (node *UnaryEnode) canonicalizeChildren(children []EclassID) {
	if len(children) != 1 {
		panic(fmt.Sprintf("expected 1 child, got %d", len(children)))
	}

	node.ClassID = children[0]
}

type IntEnode struct {
	Value int64
}

func (node *IntEnode) String(graph *Egraph) string {
	return strconv.FormatInt(node.Value, 10)
}

func (node *IntEnode) Key() enodeKey {
	var data [wordSize * 2]byte
	bytes := *(*[8]byte)(unsafe.Pointer(&node.Value))

	copy(data[:], bytes[:])

	return enodeKey{
		kind: enodeKeyKindInt,
		data: data,
	}
}

func (node *IntEnode) Children() []EclassID {
	return nil
}

func (node *IntEnode) canonicalizeChildren(children []EclassID) {
	if len(children) != 0 {
		panic(fmt.Sprintf("expected 0 children, got %d", len(children)))
	}
}

type VariableEnode struct {
	Name stringIdx
}

func (node *VariableEnode) Key() enodeKey {
	var data [wordSize * 2]byte
	bytes := (*[4]byte)(unsafe.Pointer(&node.Name))

	copy(data[:], bytes[:])

	return enodeKey{
		kind: enodeKeyKindVariable,
		data: data,
	}
}

func (node *VariableEnode) String(graph *Egraph) string {
	return graph.interner.get(node.Name)
}

func (node *VariableEnode) Children() []EclassID {
	return nil
}

func (node *VariableEnode) canonicalizeChildren(children []EclassID) {
	if len(children) != 0 {
		panic(fmt.Sprintf("expected 0 children, got %d", len(children)))
	}
}
