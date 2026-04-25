package hashmap

import (
	"hash/maphash"
	"iter"
)

type key interface {
	FillHash(hash *maphash.Hash)
}

// it could be a part of `key` interface but it's ugly to use any :\
type cmp[T key] func(lhs, rhs T) bool

type node[K key, V any] struct {
	key   K
	value V
	next  *node[K, V]
}

// not thread-safe btw
type HashMap[K key, V any] struct {
	hash    maphash.Hash
	buckets []*node[K, V]
	cmp     cmp[K]
}

func New[K key, V any](size uint, cmp cmp[K]) *HashMap[K, V] {
	return &HashMap[K, V]{
		buckets: make([]*node[K, V], size),
		cmp:     cmp,
	}
}

func (hm *HashMap[K, V]) index(key K) uint64 {
	hm.hash.Reset()
	key.FillHash(&hm.hash)

	return hm.hash.Sum64() % uint64(len(hm.buckets))
}

func (hm *HashMap[K, V]) Insert(key K, value V) (V, bool) {
	idx := hm.index(key)
	head := hm.buckets[idx]

	for n := range iterNodes(head) {
		if hm.cmp(n.key, key) {
			old := n.value
			n.value = value

			return old, true
		}
	}

	hm.buckets[idx] = &node[K, V]{
		key:   key,
		value: value,
		next:  head,
	}

	return *new(V), false
}

func (hm *HashMap[K, V]) Get(key K) (V, bool) {
	idx := hm.index(key)
	head := hm.buckets[idx]

	for n := range iterNodes(head) {
		if hm.cmp(n.key, key) {
			return n.value, true
		}
	}

	return *new(V), false
}

func (hm *HashMap[K, V]) Remove(key K) (V, bool) {
	idx := hm.index(key)
	head := hm.buckets[idx]

	if head == nil {
		return *new(V), false
	}

	if hm.cmp(head.key, key) {
		value := head.value
		hm.buckets[idx] = head.next

		return value, true
	}

	for ; head.next != nil; head = head.next {
		if hm.cmp(head.next.key, key) {
			value := head.next.value
			head.next = head.next.next

			return value, true
		}
	}

	return *new(V), false
}

func iterNodes[K key, V any](n *node[K, V]) iter.Seq[*node[K, V]] {
	return func(yield func(*node[K, V]) bool) {
		for ; n != nil; n = n.next {
			if !yield(n) {
				return
			}
		}
	}
}
