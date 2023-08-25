package cache

import "sync"

type MemCache[K comparable, T any] interface {
	Delete(key K)
	Exist(key K) bool
	Get(key K) T
	Put(key K, value T)
	GetPut(key K, value T) T
}

func NewMemCache[K comparable, T any]() MemCache[K, T] {
	return &memCache[K, T]{
		m: map[K]T{},
	}
}

type memCache[K comparable, T any] struct {
	mtx sync.RWMutex
	m   map[K]T
}

func (c *memCache[K, T]) Delete(key K) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	delete(c.m, key)
}

func (c *memCache[K, T]) Exist(key K) bool {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	_, ok := c.m[key]
	return ok
}

func (c *memCache[K, T]) Get(key K) T {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	value, _ := c.m[key]
	return value
}

func (c *memCache[K, T]) Put(key K, value T) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.m[key] = value
}

func (c *memCache[K, T]) GetPut(key K, value T) T {
	c.mtx.RLock()
	if value, ok := c.m[key]; ok {
		c.mtx.RUnlock()
		return value
	}
	c.mtx.RUnlock()

	c.mtx.Lock()
	c.m[key] = value
	c.mtx.Unlock()

	return value
}
