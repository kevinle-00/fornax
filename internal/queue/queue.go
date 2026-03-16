// Package queue provides a bounded job queue for the worker pool.
package queue

import (
	"context"
	"errors"
	"sync"

	"github.com/kevinle-00/fornax/internal/job"
)

var ErrQueueFull = errors.New("queue is full")

type Queue struct {
	jobs chan job.Job
	all  []job.Job
	mu   sync.Mutex
}

func New(capacity int) *Queue {
	return &Queue{
		jobs: make(chan job.Job, capacity),
	}
}

func (q *Queue) Enqueue(job job.Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	select {
	case q.jobs <- job:
		q.all = append(q.all, job)
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *Queue) Dequeue(ctx context.Context) (job.Job, bool) {
	select {
	case job, ok := <-q.jobs:
		if !ok {
			return nil, false
		}
		return job, true
	case <-ctx.Done():
		return nil, false
	}
}

func (q *Queue) Jobs() []job.Job {
	q.mu.Lock()
	defer q.mu.Unlock()
	allCopy := make([]job.Job, len(q.all))
	copy(allCopy, q.all)
	return allCopy
}

func (q *Queue) Close() {
	close(q.jobs)
}
