package cmd

import (
	"fmt"

	"github.com/kevinle-00/fornax/internal/download"
	"github.com/kevinle-00/fornax/internal/job"
	"github.com/kevinle-00/fornax/internal/validate"
	"github.com/spf13/cobra"
)

func newDownloadCommand(downloader download.Downloader) *cobra.Command {
	var outputDirectory string
	var quality string

	downloadCmd := &cobra.Command{
		Use:   "download <url>...",
		Short: "Download media from URLs",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validate.IsValidOutputPath(outputDirectory); err != nil {
				return err
			}
			for _, url := range args {
				if err := validate.IsValidURL(url); err != nil {
					return fmt.Errorf("invalid source %q: %w", url, err)
				}
			}

			jobs := make([]namedJob, 0, len(args))
			for _, url := range args {
				inputs := job.DownloadInputs{
					URL:             url,
					OutputDirectory: outputDirectory,
					Quality:         quality,
				}
				jobs = append(jobs, namedJob{source: url, job: job.NewDownloadJob(inputs, downloader)})
			}
			return runJobs(cmd.Context(), jobs, cmd.OutOrStdout())
		},
	}

	downloadCmd.Flags().StringVarP(&outputDirectory, "output", "o", ".", "Existing output directory")
	downloadCmd.Flags().StringVarP(&quality, "quality", "q", "", "yt-dlp format selector")
	return downloadCmd
}
