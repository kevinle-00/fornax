package cmd

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevinle-00/fornax/internal/download"
	"github.com/kevinle-00/fornax/internal/encode"
	"github.com/kevinle-00/fornax/internal/queue"
	"github.com/kevinle-00/fornax/internal/ui"
	"github.com/kevinle-00/fornax/internal/worker"
	"github.com/spf13/cobra"
)

const defaultWorkers = 5

func Execute(ctx context.Context) error {
	return newRootCommand(download.New(), encode.New()).ExecuteContext(ctx)
}

func newRootCommand(downloader download.Downloader, encoder encode.Encoder) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "fornax",
		Short:         "Download and convert media",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTUI(cmd.Context(), downloader, encoder)
		},
	}

	rootCmd.AddCommand(
		newDownloadCommand(downloader),
		newEncodeCommand(encoder),
		newProcessCommand(downloader, encoder),
	)
	return rootCmd
}

func runTUI(parentCtx context.Context, downloader download.Downloader, encoder encode.Encoder) error {
	ctx, cancel := context.WithCancel(parentCtx)
	jobQueue := queue.New(10)
	workerPool := worker.NewWorkerPool(jobQueue, defaultWorkers)
	workerPool.Start(ctx)

	_, runErr := tea.NewProgram(ui.NewModel(jobQueue, downloader, encoder), tea.WithContext(ctx)).Run()
	cancel()
	jobQueue.Close()
	workerPool.Stop()

	if runErr != nil && !errors.Is(runErr, tea.ErrProgramKilled) {
		return fmt.Errorf("run TUI: %w", runErr)
	}
	return nil
}
