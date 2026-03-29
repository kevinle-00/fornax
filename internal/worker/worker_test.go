package worker_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/kevinle-00/fornax/internal/job"
	"github.com/kevinle-00/fornax/internal/queue"
	"github.com/kevinle-00/fornax/internal/worker"
)

type mockJob struct {
	id       string
	status   job.Status
	execErr  error
	executed bool
	mu       sync.Mutex
}

func newMockJob(id string, execErr error) *mockJob {
	return &mockJob{
		id:      id,
		status:  job.StatusPending,
		execErr: execErr,
	}
}

func (m *mockJob) Execute(ctx context.Context) error {
	m.mu.Lock()
	m.executed = true
	if m.execErr != nil {
		m.status = job.StatusFailed
	} else {
		m.status = job.StatusDone
	}
	m.mu.Unlock()
	return m.execErr
}

func (m *mockJob) ID() string {
	return m.id
}

func (m *mockJob) Status() job.Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status
}

func (m *mockJob) Error() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.execErr
}

func (m *mockJob) Progress() float64 {
	return 0
}

func (m *mockJob) Requeue() job.Job {
	return newMockJob(m.id, m.execErr)
}

func (m *mockJob) wasExecuted() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.executed
}

func TestWorkerPoolProcessesJobs(t *testing.T) {
	q := queue.New(10)

	jobs := make([]*mockJob, 5)
	for i := range jobs {
		jobs[i] = newMockJob(string(rune('A'+i)), nil)
		if err := q.Enqueue(jobs[i]); err != nil {
			t.Fatalf("failed to enqueue job: %v", err)
		}
	}

	pool := worker.NewWorkerPool(q, 3)
	pool.Start(context.Background())

	q.Close()
	pool.Stop()

	for _, j := range jobs {
		if !j.wasExecuted() {
			t.Errorf("job %s was not executed", j.ID())
		}
		if j.Status() != job.StatusDone {
			t.Errorf("job %s has status %s, want %s", j.ID(), j.Status(), job.StatusDone)
		}
	}
}

func TestWorkerPoolHandlesFailedJobs(t *testing.T) {
	q := queue.New(10)

	goodJob1 := newMockJob("good-1", nil)
	failJob := newMockJob("fail-1", errors.New("something went wrong"))
	goodJob2 := newMockJob("good-2", nil)

	for _, j := range []job.Job{goodJob1, failJob, goodJob2} {
		if err := q.Enqueue(j); err != nil {
			t.Fatalf("failed to enqueue job: %v", err)
		}
	}

	pool := worker.NewWorkerPool(q, 2)
	pool.Start(context.Background())

	q.Close()
	pool.Stop()

	if !goodJob1.wasExecuted() {
		t.Error("good-1 was not executed")
	}
	if !failJob.wasExecuted() {
		t.Error("fail-1 was not executed")
	}
	if !goodJob2.wasExecuted() {
		t.Error("good-2 was not executed")
	}

	if goodJob1.Status() != job.StatusDone {
		t.Errorf("good-1 has status %s, want %s", goodJob1.Status(), job.StatusDone)
	}
	if failJob.Status() != job.StatusFailed {
		t.Errorf("fail-1 has status %s, want %s", failJob.Status(), job.StatusFailed)
	}
	if goodJob2.Status() != job.StatusDone {
		t.Errorf("good-2 has status %s, want %s", goodJob2.Status(), job.StatusDone)
	}
}

func TestWorkerPoolRespectsContext(t *testing.T) {
	q := queue.New(10)

	ctx, cancel := context.WithCancel(context.Background())

	pool := worker.NewWorkerPool(q, 2)
	pool.Start(ctx)

	cancel()

	done := make(chan struct{})
	go func() {
		pool.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Workers stopped as expected.
	case <-time.After(2 * time.Second):
		t.Fatal("workers did not stop after context cancellation")
	}
}
