package cmd

import (
	"fmt"
	"strings"

	"github.com/kevinle-00/fornax/internal/encode"
	"github.com/kevinle-00/fornax/internal/job"
	"github.com/kevinle-00/fornax/internal/validate"
	"github.com/spf13/cobra"
)

func newEncodeCommand(encoder encode.Encoder) *cobra.Command {
	var outputDirectory string
	var format string

	encodeCmd := &cobra.Command{
		Use:   "encode <input>...",
		Short: "Convert local media files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			format = strings.TrimSpace(format)
			if err := validate.IsValidFormat(format); err != nil {
				return err
			}
			if err := validate.IsValidOutputPath(outputDirectory); err != nil {
				return err
			}
			for _, inputPath := range args {
				if err := validate.IsValidInputPath(inputPath); err != nil {
					return fmt.Errorf("invalid input %q: %w", inputPath, err)
				}
			}

			jobs := make([]namedJob, 0, len(args))
			for _, inputPath := range args {
				inputs := job.EncodeInputs{
					InputPath:       inputPath,
					OutputDirectory: outputDirectory,
					Format:          format,
				}
				jobs = append(jobs, namedJob{source: inputPath, job: job.NewEncodeJob(inputs, encoder)})
			}
			return runJobs(cmd.Context(), jobs, cmd.OutOrStdout())
		},
	}

	encodeCmd.Flags().StringVarP(&outputDirectory, "output", "o", ".", "Existing output directory")
	encodeCmd.Flags().StringVarP(&format, "format", "f", "", "Output format (required)")
	return encodeCmd
}
