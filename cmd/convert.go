package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var convertCmd = &cobra.Command{
	Use:   "convert <input> <output>",
	Short: "Convert media file format",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		inputPath := args[0]
		outputPath := args[1]

		err := convert(inputPath, outputPath)
		if err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(convertCmd)
}

func convert(inputPath string, outputPath string) error {
	cmd := exec.Command("ffmpeg", "-i", inputPath, outputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
