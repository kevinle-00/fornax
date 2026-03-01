// Package queue
package queue

import (
	"fmt"
	"sync"

	"github.com/kevinle-00/fornax/internal/job"
)

type Queue interface {
	Enqueue(job job.Job) error
	Dequeue() job.Job
	GetJobs() []job.Job
}

type JobQueue struct {
	jobs chan job.Job
	all  []job.Job
	mu   sync.Mutex
}

func New(capacity int) *JobQueue {
	return &JobQueue{
		jobs: make(chan job.Job, capacity),
		all:  []job.Job{},
	}
}

func (j *JobQueue) Enqueue(job job.Job) error {
	select {
	case j.jobs <- job:
		j.mu.Lock()
		j.all = append(j.all, job)
		j.mu.Unlock()
		return nil
	default:
		return fmt.Errorf("queue is full")
	}
}

func (j *JobQueue) Dequeue() job.Job {
	return <-j.jobs
}

func (j *JobQueue) GetJobs() []job.Job {
	j.mu.Lock()
	defer j.mu.Unlock()
	allCopy := make([]job.Job, len(j.all))
	copy(allCopy, j.all)
	return allCopy
}
