package utils

import "sync"

type Map[T comparable] struct {
  mtx sync.RWMutex
  m   map[T]struct{}
}

func NewMap[T comparable]() *Map[T] {
  return &Map[T]{
    m: map[T]struct{}{},
  }
}

func (c *Map[T]) Exist(key T) bool {
  c.mtx.RLock()
  defer c.mtx.RUnlock()
  _, ok := c.m[key]
  return ok
}

func (c *Map[T]) Count() int {
  c.mtx.RLock()
  defer c.mtx.RUnlock()
  return len(c.m)
}

func (c *Map[T]) Put(key T) {
  c.mtx.Lock()
  defer c.mtx.Unlock()
  c.m[key] = struct{}{}
}

func (c *Map[T]) Delete(key T) {
  c.mtx.Lock()
  defer c.mtx.Unlock()
  delete(c.m, key)
}

func (c *Map[T]) Clear() {
  c.mtx.Lock()
  defer c.mtx.Unlock()
  c.m = map[T]struct{}{}
}
