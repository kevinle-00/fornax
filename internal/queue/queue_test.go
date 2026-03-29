// Package queue_test
package queue_test

import (
	"context"
	"testing"
	"time"

	"github.com/kevinle-00/fornax/internal/job"
	"github.com/kevinle-00/fornax/internal/queue"
)

// mockJob implements job.Job.

type mockJob struct {
	id     string
	status job.Status
}

func (m *mockJob) Execute(ctx context.Context) error {
	m.status = job.StatusDone
	return nil
}

func (m *mockJob) ID() string {
	return m.id
}

func (m *mockJob) Status() job.Status {
	return m.status
}

func (m *mockJob) Error() error {
	return nil
}

func (m *mockJob) Progress() float64 {
	return 0
}

func (m *mockJob) Requeue() job.Job {
	return &mockJob{id: m.id, status: job.StatusPending}
}

func TestEnqueue(t *testing.T) {
	tests := []struct {
		name        string
		capacity    int
		jobCount    int
		wantErr     bool
		expectedLen int
	}{
		{
			name:        "Enqueue with enough capacity",
			capacity:    10,
			jobCount:    5,
			wantErr:     false,
			expectedLen: 5,
		},
		{
			name:        "Enqueue into a full queue",
			capacity:    10,
			jobCount:    11,
			wantErr:     true,
			expectedLen: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := queue.New(tt.capacity)
			job := &mockJob{id: "1"}

			for i := 0; i < tt.jobCount; i++ {
				err := queue.Enqueue(job)

				if i < tt.capacity && err != nil {
					t.Errorf("enqueue at index %d failed: %v", i, err)
					return
				} else if i >= tt.capacity && err == nil {
					t.Errorf("enqueue at index %d should have failed (capacity %d), but succeeded", i, tt.capacity)
				}

			}

			jobsSlice := queue.Jobs()
			count := len(jobsSlice)
			if count != tt.expectedLen {
				t.Errorf("expected %d job, got %d", tt.jobCount, count)
			}
		})
	}
}

func TestDequeue(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		jobCount int
		wantOK   bool
	}{
		{
			name:     "Dequeue with jobs in queue",
			capacity: 10,
			jobCount: 5,
			wantOK:   true,
		},
		{
			name:     "Dequeue with no jobs in queue",
			capacity: 10,
			jobCount: 0,
			wantOK:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := queue.New(tt.capacity)
			job := &mockJob{id: "1"}

			for i := 0; i < tt.jobCount; i++ {
				err := queue.Enqueue(job)
				if err != nil {
					t.Errorf("Failed to enqueue job")
					return
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			_, ok := queue.Dequeue(ctx)
			if ok != tt.wantOK {
				t.Errorf("Dequeue expected %t, got %t", tt.wantOK, ok)
			}
		})
	}
}

func TestClose(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		jobCount int
	}{
		{
			name:     "Dequeue returns false after Close",
			capacity: 10,
			jobCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := queue.New(tt.capacity)
			job := &mockJob{id: "1"}

			for i := 0; i < tt.jobCount; i++ {
				err := q.Enqueue(job)
				if err != nil {
					t.Errorf("Failed to enqueue job")
					return
				}
			}

			q.Close()

			for {
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				_, ok := q.Dequeue(ctx)
				cancel()
				if !ok {
					break
				}
			}

			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			_, ok := q.Dequeue(ctx)
			if ok {
				t.Errorf("Dequeue after Close expected false, got true")
			}
		})
	}
}

func TestJobsReturnsCopy(t *testing.T) {
	tests := []struct {
		name        string
		capacity    int
		jobCount    int
		expectedLen int
	}{
		{
			name:        "Appending to returned slice does not affect internal state",
			capacity:    10,
			jobCount:    3,
			expectedLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := queue.New(tt.capacity)
			job := &mockJob{id: "1"}

			for i := 0; i < tt.jobCount; i++ {
				err := q.Enqueue(job)
				if err != nil {
					t.Errorf("Failed to enqueue job")
					return
				}
			}

			jobs := q.Jobs()
			_ = append(jobs, &mockJob{id: "extra"})

			jobsAgain := q.Jobs()
			if len(jobsAgain) != tt.expectedLen {
				t.Errorf("expected %d jobs, got %d", tt.expectedLen, len(jobsAgain))
			}
		})
	}
}
