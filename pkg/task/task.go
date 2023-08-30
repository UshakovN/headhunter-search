package task

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
)

type Queue interface {
	Push(task func() error)
	ContinuouslyHandle(ctx context.Context)
}

func NewQueue(workers int) Queue {
	return &queue{
		workers: workers,
		tasks:   make(chan func() error),
	}
}

type queue struct {
	workers int
	tasks   chan func() error
}

func (q *queue) Push(task func() error) {
	q.tasks <- task
}

func (q *queue) ContinuouslyHandle(ctx context.Context) {
	wg := sync.WaitGroup{}

	for i := 0; i < q.workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			select {
			case <-ctx.Done():
				log.Infof("task handling stopped. context cancelled")
				return

			case task, ok := <-q.tasks:
				if ok {
					if err := task(); err != nil {
						log.Errorf("task handling error: %v", err)
					}
				}
			}
		}()
	}
	wg.Wait()
}
