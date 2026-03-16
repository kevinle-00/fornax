package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/kevinle-00/fornax/internal/download"
	"github.com/kevinle-00/fornax/internal/encode"
	"github.com/kevinle-00/fornax/internal/job"
	"github.com/kevinle-00/fornax/internal/queue"
	"github.com/kevinle-00/fornax/internal/validate"
	"github.com/kevinle-00/fornax/internal/worker"
	"github.com/spf13/cobra"
)

// fornax process <url>... -f mp4 -o output/

var processCmd = &cobra.Command{
	Use:   "process <url>... -f <format>",
	Short: "Download youtube video and then convert to desired media file format",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		outputDirectory, _ := cmd.Flags().GetString("output")
		format, _ := cmd.Flags().GetString("format")

		if err := validate.IsValidOutputPath(outputDirectory); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		quality, _ := cmd.Flags().GetString("quality")

		encoder := encode.New()
		downloader := download.New()
		jobQueue := queue.New(10)
		for _, url := range args {
			if err := validate.IsValidURL(url); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			processInputs := job.ProcessInputs{
				URL:             url,
				OutputDirectory: outputDirectory,
				Format:          format,
				Quality:         quality,
			}
			newJob := job.NewProcessJob(processInputs, downloader, encoder)
			if err := jobQueue.Enqueue(newJob); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		}
		jobQueue.Close()
		workerPool := worker.NewWorkerPool(jobQueue, 5)
		workerPool.Start(context.Background())

		workerPool.Stop()
	},
}

func init() {
	rootCmd.AddCommand(processCmd)
	processCmd.Flags().StringP("quality", "q", "", "Video quality")
	processCmd.Flags().StringP("output", "o", ".", "Output directory")
	processCmd.Flags().StringP("format", "f", "", "Format")
	_ = processCmd.MarkFlagRequired("format")
}
