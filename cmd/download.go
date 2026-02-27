// Package cmd contains CLI commands
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/kevinle-00/fornax/internal/download"
	"github.com/kevinle-00/fornax/internal/validate"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download <url> [output]",
	Short: "Download from youtube URL",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		if err := validate.IsValidURL(url); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		output, _ := cmd.Flags().GetString("output")
		if err := validate.IsValidOutputPath(output); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		quality, _ := cmd.Flags().GetString("quality")
		downloader := download.New()
		// TODO: eventually make download cancellable by the user
		err := downloader.Download(context.Background(), url, output, quality)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringP("output", "o", "", "Output file")
	downloadCmd.Flags().StringP("quality", "q", "", "Video quality")
}
