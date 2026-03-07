// Package queue_test
package queue_test

import (
	"context"
	"testing"

	"github.com/kevinle-00/fornax/internal/job"
	"github.com/kevinle-00/fornax/internal/queue"
)

type mockJob struct {
	id     string
	status job.Status
}

func (m *mockJob) Execute(ctx context.Context) error {
	m.status = job.StatusDone
	return nil
}

func (m *mockJob) GetID() string {
	return m.id
}

func (m *mockJob) GetStatus() job.Status {
	return m.status
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
			queue := queue.NewJobQueue(tt.capacity)
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

			jobsSlice := queue.GetJobs()
			count := len(jobsSlice)
			if count != tt.expectedLen {
				t.Errorf("expected %d job, got %d", tt.jobCount, count)
			}
		})
	}
}
