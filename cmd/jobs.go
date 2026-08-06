package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/kevinle-00/fornax/internal/job"
	"github.com/kevinle-00/fornax/internal/queue"
	"github.com/kevinle-00/fornax/internal/worker"
)

type namedJob struct {
	source string
	job    job.Job
}

func runJobs(ctx context.Context, jobs []namedJob, output io.Writer) error {
	jobQueue := queue.New(len(jobs))
	for _, item := range jobs {
		if err := jobQueue.Enqueue(item.job); err != nil {
			jobQueue.Close()
			return fmt.Errorf("queue %q: %w", item.source, err)
		}
	}
	jobQueue.Close()

	workerCount := min(defaultWorkers, len(jobs))
	workerPool := worker.NewWorkerPool(jobQueue, workerCount)
	workerPool.Start(ctx)
	workerPool.Stop()

	failures := 0
	for _, item := range jobs {
		var writeErr error
		switch item.job.Status() {
		case job.StatusDone:
			_, writeErr = fmt.Fprintf(output, "DONE %s\n", item.source)
		case job.StatusFailed:
			failures++
			_, writeErr = fmt.Fprintf(output, "FAILED %s: %v\n", item.source, item.job.Error())
		default:
			failures++
			_, writeErr = fmt.Fprintf(output, "CANCELED %s\n", item.source)
		}
		if writeErr != nil {
			return fmt.Errorf("write job result: %w", writeErr)
		}
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	if failures > 0 {
		return fmt.Errorf("%d of %d jobs failed", failures, len(jobs))
	}
	return nil
}
