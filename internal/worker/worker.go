// Package worker runs jobs concurrently.
package worker

import (
	"context"
	"sync"

	"github.com/kevinle-00/fornax/internal/queue"
)

type WorkerPool struct {
	queue *queue.Queue
	count int
	wg    sync.WaitGroup
}

func NewWorkerPool(q *queue.Queue, count int) *WorkerPool {
	return &WorkerPool{
		queue: q,
		count: count,
	}
}

func (w *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < w.count; i++ {
		w.wg.Go(func() {
			for {
				job, ok := w.queue.Dequeue(ctx)

				if !ok {
					return
				}

				// Jobs store their own errors for the UI and CLI.
				_ = job.Execute(ctx)
			}
		})
	}
}

func (w *WorkerPool) Stop() {
	w.wg.Wait()
}
