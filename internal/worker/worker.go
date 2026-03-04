// Package worker
package worker

import (
	"context"
	"log"
	"sync"

	"github.com/kevinle-00/fornax/internal/queue"
)

type WorkerPool struct {
	queue *queue.JobQueue
	count int
	wg    sync.WaitGroup
}

type Pool interface {
	Start(ctx context.Context) error
	Stop()
}

func NewWorkerPool(queue *queue.JobQueue, count int) *WorkerPool {
	return &WorkerPool{
		queue: queue,
		count: count,
	}
}

func (w *WorkerPool) Start(ctx context.Context) error {
	for i := 0; i < w.count; i++ {
		w.wg.Go(func() {
			for {
				job, ok := w.queue.Dequeue(ctx)

				if !ok {
					return
				}

				if err := job.Execute(ctx); err != nil {
					log.Printf("job %s failed: %v", job.GetID(), err)
				}

			}
		})
	}
	return nil
}

func (w *WorkerPool) Stop() {
	w.wg.Wait()
}
