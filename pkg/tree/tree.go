package tree

type Tree[T comparable, E any] interface {
	Push(link T, entity E)
	Next(link T) Tree[T, E]
	Prev() Tree[T, E]
	Entity() E
}

func NewTree[T comparable, E any]() Tree[T, E] {
	return &node[T, E]{
		next: map[T]*node[T, E]{},
	}
}

type node[T comparable, E any] struct {
	entity E
	prev   *node[T, E]
	next   map[T]*node[T, E]
}

func (n *node[T, E]) Push(link T, entity E) {
	n.next[link] = &node[T, E]{
		entity: entity,
		prev:   n,
		next:   map[T]*node[T, E]{},
	}
}

func (n *node[T, E]) Next(link T) Tree[T, E] {
	return n.next[link]
}

func (n *node[T, E]) Prev() Tree[T, E] {
	return n.prev
}

func (n *node[T, E]) Entity() E {
	if n == nil {
		return *new(E)
	}
	return n.entity
}
