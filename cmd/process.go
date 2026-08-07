package cmd

import (
	"fmt"
	"strings"

	"github.com/kevinle-00/fornax/internal/download"
	"github.com/kevinle-00/fornax/internal/encode"
	"github.com/kevinle-00/fornax/internal/job"
	"github.com/kevinle-00/fornax/internal/validate"
	"github.com/spf13/cobra"
)

func newProcessCommand(downloader download.Downloader, encoder encode.Encoder) *cobra.Command {
	var outputDirectory string
	var format string
	var quality string

	processCmd := &cobra.Command{
		Use:   "process <url>...",
		Short: "Download and convert media",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format = strings.TrimSpace(format)
			if err := validate.IsValidFormat(format); err != nil {
				return err
			}
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
				inputs := job.ProcessInputs{
					URL:             url,
					OutputDirectory: outputDirectory,
					Format:          format,
					Quality:         quality,
				}
				jobs = append(jobs, namedJob{source: url, job: job.NewProcessJob(inputs, downloader, encoder)})
			}
			return runJobs(cmd.Context(), jobs, cmd.OutOrStdout())
		},
	}

	processCmd.Flags().StringVarP(&quality, "quality", "q", "", "yt-dlp format selector")
	processCmd.Flags().StringVarP(&outputDirectory, "output", "o", ".", "Existing output directory")
	processCmd.Flags().StringVarP(&format, "format", "f", "", "Output format (required)")
	return processCmd
}
