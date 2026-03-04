// Package cmd contains CLI commands
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/kevinle-00/fornax/internal/download"
	"github.com/kevinle-00/fornax/internal/job"
	"github.com/kevinle-00/fornax/internal/queue"
	"github.com/kevinle-00/fornax/internal/validate"
	"github.com/kevinle-00/fornax/internal/worker"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download <url>... ",
	Short: "Download from youtube URL",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		output, _ := cmd.Flags().GetString("output")
		if err := validate.IsValidOutputPath(output); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		info, err := os.Stat(output)
		if err == nil && len(args) > 1 && !info.IsDir() {
			fmt.Println("Error: output must be a directory when downloading multiple URLs")
			os.Exit(1)
		}
		quality, _ := cmd.Flags().GetString("quality")
		downloader := download.New()
		// TODO: eventually make download cancellable by the user

		jobQueue := queue.NewJobQueue(10)
		workerPool := worker.NewWorkerPool(jobQueue, 5)
		for _, url := range args {
			if err := validate.IsValidURL(url); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
			downloadInputs := job.DownloadInputs{
				URL:        url,
				OutputPath: output,
				Quality:    quality,
			}
			newJob := job.NewDownloadJob(downloadInputs, downloader)
			if err := jobQueue.Enqueue(newJob); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		}
		// Close the queue so workers dont block waiting for more work
		jobQueue.Close()

		if err := workerPool.Start(context.Background()); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		workerPool.Stop()
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringP("output", "o", "", "Output file")
	downloadCmd.Flags().StringP("quality", "q", "", "Video quality")
}
