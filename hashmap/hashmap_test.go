package hashmap_test

import (
	"hash/maphash"
	"testing"

	"github.com/MilkeeyCat/expr_simplifier/hashmap"
	"github.com/stretchr/testify/assert"
)

type key struct {
	value string
}

func (k *key) FillHash(hash *maphash.Hash) {
	if _, err := hash.WriteString(k.value); err != nil {
		panic("failed to write bytes into hash")
	}
}

func cmp(lhs, rhs *key) bool {
	return lhs.value == rhs.value
}

func TestInsert(t *testing.T) {
	m := hashmap.New[*key, uint](10, cmp)

	old, ok := m.Insert(&key{"foo"}, 69)
	assert.Zero(t, old)
	assert.False(t, ok)

	old, ok = m.Insert(&key{"foo"}, 420)
	assert.Equal(t, uint(69), old)
	assert.True(t, ok)
}

func TestGet(t *testing.T) {
	m := hashmap.New[*key, uint](20, cmp)

	value, ok := m.Get(&key{"bar"})
	assert.Zero(t, value)
	assert.False(t, ok)

	m.Insert(&key{"bar"}, 29)

	value, ok = m.Get(&key{"bar"})
	assert.Equal(t, uint(29), value)
	assert.True(t, ok)
}

func TestRemove(t *testing.T) {
	m := hashmap.New[*key, uint](15, cmp)

	value, ok := m.Remove(&key{"baz"})
	assert.Zero(t, value)
	assert.False(t, ok)

	m.Insert(&key{"baz"}, 95)

	value, ok = m.Remove(&key{"baz"})
	assert.Equal(t, uint(95), value)
	assert.True(t, ok)

	value, ok = m.Remove(&key{"baz"})
	assert.Zero(t, value)
	assert.False(t, ok)
}
