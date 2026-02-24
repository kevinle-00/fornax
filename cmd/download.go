// Package cmd contains CLI commands
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download <url> [output]",
	Short: "Download from youtube URL",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url := args[0]
		output, _ := cmd.Flags().GetString("output")
		quality, _ := cmd.Flags().GetString("quality")
		err := download(url, output, quality)
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

func download(url, output, quality string) error {
	cmdArgs := []string{}
	if output != "" {
		cmdArgs = append(cmdArgs, "-o", output)
	}
	if quality != "" {
		cmdArgs = append(cmdArgs, "-f", quality)
	}
	cmdArgs = append(cmdArgs, url)
	cmd := exec.Command("yt-dlp", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
