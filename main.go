package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fornax <url>")
		os.Exit(1)
	}

	url := os.Args[1]
	err := download(url)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func download(url string) error {
	cmd := exec.Command("yt-dlp", url)
	// Piping the yt-dlp's standard output/errors to our terminal
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
