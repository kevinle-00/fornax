package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kevinle-00/fornax/internal/download"
	"github.com/kevinle-00/fornax/internal/encode"
	"github.com/kevinle-00/fornax/internal/queue"
	"github.com/kevinle-00/fornax/internal/ui"
	"github.com/kevinle-00/fornax/internal/worker"
)

func main() {
	downloader := download.New()
	encoder := encode.New()
	queue := queue.NewJobQueue(10)
	workerPool := worker.NewWorkerPool(queue, 5)
	defer workerPool.Stop()
	defer queue.Close()

	workerPool.Start(context.Background())

	if _, err := tea.NewProgram(ui.NewModel(queue, downloader, encoder)).Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	/*
		if err := cmd.Execute(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	*/
}
