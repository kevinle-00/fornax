package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kevinle-00/fornax/internal/download"
	"github.com/kevinle-00/fornax/internal/encode"
	"github.com/kevinle-00/fornax/internal/queue"
	"github.com/kevinle-00/fornax/internal/ui"
	"github.com/kevinle-00/fornax/internal/worker"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	downloader := download.New()
	encoder := encode.New()
	queue := queue.New(10)
	workerPool := worker.NewWorkerPool(queue, 5)

	workerPool.Start(ctx)

	_, runErr := tea.NewProgram(ui.NewModel(queue, downloader, encoder), tea.WithContext(ctx)).Run()
	cancel()
	queue.Close()
	workerPool.Stop()

	if runErr != nil && !errors.Is(runErr, tea.ErrProgramKilled) {
		return fmt.Errorf("run TUI: %w", runErr)
	}
	return nil
}
