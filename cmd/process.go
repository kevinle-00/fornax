package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kevinle-00/fornax/internal/download"
	"github.com/kevinle-00/fornax/internal/encode"
	"github.com/spf13/cobra"
)

// fornax process <url> <output>

var processCmd = &cobra.Command{
	Use:   "process <url> <output>",
	Short: "Download youtube video and then convert to desired media file format",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		outputPath := args[1]
		quality, _ := cmd.Flags().GetString("quality")

		err := process(url, outputPath, quality)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(processCmd)
	processCmd.Flags().StringP("quality", "q", "", "Video quality")
}

func process(url, outputPath, quality string) error {
	tempPath := "/tmp/fornax-temp.%(ext)s"
	downloader := download.New()
	err := downloader.Download(context.Background(), url, tempPath, quality)
	if err != nil {
		return err
	}

	files, globErr := filepath.Glob("/tmp/fornax-temp.*")
	if globErr != nil {
		return globErr
	}
	if len(files) == 0 {
		return fmt.Errorf("could not find downloaded file in /tmp")
	}

	tempFile := files[0]
	defer os.Remove(tempFile)
	encoder := encode.New()

	// TODO: make sure cancellable
	err = encoder.Encode(context.Background(), tempFile, outputPath)
	if err != nil {
		return err
	}
	return nil
}
