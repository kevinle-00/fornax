package cmd

import (
	"context"
	"fmt"
	"os"
	"github.com/kevinle-00/fornax/internal/encode"
	"github.com/kevinle-00/fornax/internal/job"
	"github.com/kevinle-00/fornax/internal/queue"
	"github.com/kevinle-00/fornax/internal/validate"
	"github.com/kevinle-00/fornax/internal/worker"
	"github.com/spf13/cobra"
)

// fornax encode file1.webm file2.webm -f mp4 -o output/

var encodeCmd = &cobra.Command{
	Use:   "encode <input>... ",
	Short: "Encode media file format",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		outputDirectory, _ := cmd.Flags().GetString("output")

		if err := validate.IsValidOutputPath(outputDirectory); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		// TODO: Make encode manually cancellable by user

		encoder := encode.New()
		jobQueue := queue.NewJobQueue(10)
		for _, inputPath := range args {
			if err := validate.IsValidInputPath(inputPath); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}

			encodeInputs := job.EncodeInputs{
				InputPath:       inputPath,
				OutputDirectory: outputDirectory,
				Format:          format,
			}
			newJob := job.NewEncodeJob(encodeInputs, encoder)
			if err := jobQueue.Enqueue(newJob); err != nil {
				fmt.Println("Error:", err)
				os.Exit(1)
			}
		}
		jobQueue.Close()
		workerPool := worker.NewWorkerPool(jobQueue, 5)
		if err := workerPool.Start(context.Background()); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		workerPool.Stop()
	},
}

func init() {
	rootCmd.AddCommand(encodeCmd)
	encodeCmd.Flags().StringP("output", "o", ".", "Output directory")
	encodeCmd.Flags().StringP("format", "f", "", "Format")
	_ = encodeCmd.MarkFlagRequired("format")
}
