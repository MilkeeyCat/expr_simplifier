package egraph

import (
	"fmt"
	"hash/maphash"
	"slices"
	"strconv"
	"unsafe"

	"github.com/MilkeeyCat/expr_simplifier/ast"
)

const (
	enodeKeyKindBinary byte = iota
	enodeKeyKindUnary
	enodeKeyKindCall
	enodeKeyKindInt
	enodeKeyKindVariable
)

const wordSize = unsafe.Sizeof(uintptr(0))

func enodeIDToBytes(id EclassID) [wordSize]byte {
	return *(*[unsafe.Sizeof(EclassID(0))]byte)(unsafe.Pointer(&id))
}

type BinaryEnode struct {
	Op  ast.BinaryOp
	Lhs EclassID
	Rhs EclassID
}

func (node *BinaryEnode) String() string {
	return node.Op.String()
}

func (node *BinaryEnode) Children() []EclassID {
	return []EclassID{node.Lhs, node.Rhs}
}

func (node *BinaryEnode) FillHash(hash *maphash.Hash) {
	lhs := enodeIDToBytes(node.Lhs)
	rhs := enodeIDToBytes(node.Rhs)

	hash.WriteByte((enodeKeyKindBinary << 4) | byte(node.Op)&0x0f)
	hash.Write(lhs[:])
	hash.Write(rhs[:])
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

func (node *UnaryEnode) String() string {
	return node.Op.String()
}

func (node *UnaryEnode) Children() []EclassID {
	return []EclassID{node.ClassID}
}

func (node *UnaryEnode) FillHash(hash *maphash.Hash) {
	bytes := enodeIDToBytes(node.ClassID)

	hash.WriteByte((enodeKeyKindUnary << 4) | byte(node.Op)&0x0f)
	hash.Write(bytes[:])
}

func (node *UnaryEnode) canonicalizeChildren(children []EclassID) {
	if len(children) != 1 {
		panic(fmt.Sprintf("expected 1 child, got %d", len(children)))
	}

	node.ClassID = children[0]
}

type CallEnode struct {
	Name string
	Args []EclassID
}

func (node *CallEnode) String() string {
	return node.Name + "(" + strconv.Itoa(len(node.Args)) + ")"
}

func (node *CallEnode) Children() []EclassID {
	return slices.Clone(node.Args)
}

func (node *CallEnode) FillHash(hash *maphash.Hash) {
	hash.WriteByte((enodeKeyKindCall << 4))

	for _, classID := range node.Args {
		bytes := enodeIDToBytes(classID)

		hash.Write(bytes[:])
	}
}

func (node *CallEnode) canonicalizeChildren(children []EclassID) {
	if len(children) != len(node.Args) {
		panic(fmt.Sprintf("expected %d children, got %d", len(node.Args), len(children)))
	}

	node.Args = children
}

type IntEnode struct {
	Value int64
}

func (node *IntEnode) String() string {
	return strconv.FormatInt(node.Value, 10)
}

func (node *IntEnode) Children() []EclassID {
	return nil
}

func (node *IntEnode) FillHash(hash *maphash.Hash) {
	bytes := [4]byte{
		byte(node.Value & 0xff),
		byte((node.Value >> 8) & 0xff),
		byte((node.Value >> 16) & 0xff),
		byte((node.Value >> 24) & 0xff),
	}

	hash.WriteByte(enodeKeyKindInt << 4)
	hash.Write(bytes[:])
}

func (node *IntEnode) canonicalizeChildren(children []EclassID) {
	if len(children) != 0 {
		panic(fmt.Sprintf("expected 0 children, got %d", len(children)))
	}
}

type VariableEnode struct {
	Name string
}

func (node *VariableEnode) String() string {
	return node.Name
}

func (node *VariableEnode) Children() []EclassID {
	return nil
}

func (node *VariableEnode) FillHash(hash *maphash.Hash) {
	hash.WriteByte(enodeKeyKindVariable << 4)
	hash.WriteString(node.Name)
}

func (node *VariableEnode) canonicalizeChildren(children []EclassID) {
	if len(children) != 0 {
		panic(fmt.Sprintf("expected 0 children, got %d", len(children)))
	}
}
