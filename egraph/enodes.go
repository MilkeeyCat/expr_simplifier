package egraph

import (
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

func enodeIDToBytes(id eclassID) [wordSize]byte {
	return *(*[unsafe.Sizeof(eclassID(0))]byte)(unsafe.Pointer(&id))
}

type binaryEnode struct {
	op  ast.BinaryOp
	lhs eclassID
	rhs eclassID
}

func (node *binaryEnode) String(graph *Egraph) string {
	return node.op.String()
}

func (node *binaryEnode) Key() enodeKey {
	var data [wordSize * 2]byte
	lhs := enodeIDToBytes(node.lhs)
	rhs := enodeIDToBytes(node.rhs)

	copy(data[:], lhs[:])
	copy(data[wordSize:], rhs[:])

	return enodeKey{
		kind: enodeKeyKindBinary,
		data: data,
	}
}

func (node *binaryEnode) Children() []eclassID {
	return []eclassID{node.lhs, node.rhs}
}

type unaryEnode struct {
	op    ast.UnaryOp
	class eclassID
}

func (node *unaryEnode) String(graph *Egraph) string {
	return node.op.String()
}

func (node *unaryEnode) Key() enodeKey {
	var data [wordSize * 2]byte
	expr := enodeIDToBytes(node.class)

	copy(data[:], expr[:])

	return enodeKey{
		kind: enodeKeyKindUnary,
		data: data,
	}
}

func (node *unaryEnode) Children() []eclassID {
	return []eclassID{node.class}
}

type intEnode struct {
	value int64
}

func (node *intEnode) String(graph *Egraph) string {
	return strconv.FormatInt(node.value, 10)
}

func (node *intEnode) Key() enodeKey {
	var data [wordSize * 2]byte
	bytes := *(*[8]byte)(unsafe.Pointer(&node.value))

	copy(data[:], bytes[:])

	return enodeKey{
		kind: enodeKeyKindInt,
		data: data,
	}
}

func (node *intEnode) Children() []eclassID {
	return nil
}

type variableEnode struct {
	name stringIdx
}

func (node *variableEnode) Key() enodeKey {
	var data [wordSize * 2]byte
	bytes := (*[4]byte)(unsafe.Pointer(&node.name))

	copy(data[:], bytes[:])

	return enodeKey{
		kind: enodeKeyKindVariable,
		data: data,
	}
}

func (node *variableEnode) String(graph *Egraph) string {
	return graph.interner.get(node.name)
}

func (node *variableEnode) Children() []eclassID {
	return nil
}
