package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	method := os.Args[1]
	switch method {
	case "download":
		if len(os.Args) < 3 {
			fmt.Println("Usage: fornax download <url> [output]")
			os.Exit(1)
		}
		url := os.Args[2]
		args := os.Args[3:] // returns [] if len < 3
		err := download(url, args...)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "convert":
		if len(os.Args) < 4 {
			fmt.Println("Usage: fornax convert <input> <output>")
			os.Exit(1)
		}
		inputPath := os.Args[2]
		outputPath := os.Args[3]
		err := convert(inputPath, outputPath)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: fornax <command> [args]")
	fmt.Println("Commands:")
	fmt.Println("  download <url> [output]  Download from URL")
	fmt.Println("  convert <input> <output> Convert media file")
}

func download(url string, opts ...string) error {
	cmdArgs := []string{url}

	// Option flags, only handles output path for now
	if len(opts) > 0 && opts[0] != "" {
		cmdArgs = []string{"-o", opts[0], url}
	}

	cmd := exec.Command("yt-dlp", cmdArgs...)
	// Piping the yt-dlp's standard output/errors to our terminal
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func convert(inputPath string, outputPath string) error {
	cmd := exec.Command("ffmpeg", "-i", inputPath, outputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
