package tree

type Tree[T comparable, E any] interface {
	Push(link T, entity E)
	Next(link T) Tree[T, E]
	Prev() Tree[T, E]
	Entity() E
	Link() T
}

func NewTree[T comparable, E any]() Tree[T, E] {
	return &node[T, E]{
		NextNodes: map[T]*node[T, E]{},
	}
}

type node[T comparable, E any] struct {
	LinkNode   T                 `json:"link_node"`
	EntityNode E                 `json:"entity_node"`
	PrevNode   *node[T, E]       `json:"prev_node"`
	NextNodes  map[T]*node[T, E] `json:"next_nodes"`
}

func (n *node[T, E]) Push(link T, entity E) {
	n.NextNodes[link] = &node[T, E]{
		LinkNode:   link,
		EntityNode: entity,
		PrevNode:   n,
		NextNodes:  map[T]*node[T, E]{},
	}
}

func (n *node[T, E]) Next(link T) Tree[T, E] {
	if n == nil {
		return *new(Tree[T, E])
	}
	return n.NextNodes[link]
}

func (n *node[T, E]) Prev() Tree[T, E] {
	if n == nil {
		return *new(Tree[T, E])
	}
	return n.PrevNode
}

func (n *node[T, E]) Entity() E {
	if n == nil {
		return *new(E)
	}
	return n.EntityNode
}

func (n *node[T, E]) Link() T {
	if n == nil {
		return *new(T)
	}
	return n.LinkNode
}
