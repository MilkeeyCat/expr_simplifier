package dsu

type Key int

type node struct {
	parent Key
	rank   uint
}

type DisjointSet struct {
	nodes []node
}

func (ds *DisjointSet) Add() Key {
	key := Key(len(ds.nodes))
	ds.nodes = append(ds.nodes, node{parent: key})

	return key
}

func (ds *DisjointSet) Find(key Key) Key {
	node := &ds.nodes[key]

	if node.parent == key {
		return key
	}

	node.parent = ds.Find(node.parent)

	return node.parent
}

func (ds *DisjointSet) Union(keyA, keyB Key) {
	rootA := &ds.nodes[ds.Find(keyA)]
	rootB := &ds.nodes[ds.Find(keyB)]

	if rootA == rootB {
		return
	}

	if rootA.rank < rootB.rank {
		rootA, rootB = rootB, rootA
	}

	rootB.parent = rootA.parent

	if rootA.rank == rootB.rank {
		rootA.rank++
	}
}
