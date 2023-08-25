package cache

import "sync"

type KeyCache[T any] interface {
	Exist(key T) bool
	Count() int
	Put(key T)
	Delete(key T)
	Clear()
}

func NewKeyCache[T comparable]() KeyCache[T] {
	return &keyCache[T]{
		m: map[T]struct{}{},
	}
}

type keyCache[T comparable] struct {
	mtx sync.RWMutex
	m   map[T]struct{}
}

func (c *keyCache[T]) Exist(key T) bool {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	_, ok := c.m[key]
	return ok
}

func (c *keyCache[T]) Count() int {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return len(c.m)
}

func (c *keyCache[T]) Put(key T) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.m[key] = struct{}{}
}

func (c *keyCache[T]) Delete(key T) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	delete(c.m, key)
}

func (c *keyCache[T]) Clear() {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.m = map[T]struct{}{}
}
