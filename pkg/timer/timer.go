package timer

import (
	"sync"
	"time"
)

type Timer[T comparable] interface {
	Set(key T, timeout time.Duration)
	Wait(key T)
}

type timer[T comparable] struct {
	mtx sync.RWMutex
	m   map[T]*wait
}

type wait struct {
	t time.Time
	d time.Duration
}

func NewTimer[T comparable]() Timer[T] {
	return &timer[T]{
		m: map[T]*wait{},
	}
}

func (t *timer[T]) Set(key T, timeout time.Duration) {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	t.m[key] = &wait{
		d: timeout,
		t: time.Now().UTC(),
	}
}

func (t *timer[T]) Wait(key T) {
	t.mtx.RLock()
	defer t.mtx.RUnlock()

	if wait, ok := t.m[key]; ok {
		now := time.Now().UTC()
		if next := wait.t.Add(wait.d); now.Before(next) {
			sub := next.Sub(now)
			time.Sleep(sub)
		}
	}
}
