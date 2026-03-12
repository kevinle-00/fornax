// Package download is a yt-dlp wrapper
package download

import (
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Interface in Go defines behaviour, a contract of methods that need to be implemented

type Downloader interface {
	Download(ctx context.Context, url, outputPath, quality string) error
}

// Structs in Go are data containers for fields, no behaviour

type YtDlp struct{} //

// Here this is the function to instantiate the ytdlp class/object? It returns the address of the yt-dlp object?

func New() *YtDlp {
	return &YtDlp{} // {} <-- struct literal syntax, creates an instance of the struct
}

// y here is a method receiver, it indicates what struct this method belongs to.
// eg.
// downloader := New() <-- creates a *YtDlp
// downloader.Download() <-- downloader is passed into Download, equivalent to this / self in other languages

func (y *YtDlp) Download(ctx context.Context, url, outputPath, quality string) error {
	cmdArgs := []string{} // Slice syntax, slices in go are dynamic arrays
	if outputPath != "" {
		info, err := os.Stat(outputPath)
		if err == nil && info.IsDir() {
			outputPath = filepath.Join(outputPath, "%(title)s.%(ext)s")
		}
		cmdArgs = append(cmdArgs, "-o", outputPath)
	}
	if quality != "" {
		cmdArgs = append(cmdArgs, "-f", quality)
	}
	cmdArgs = append(cmdArgs, url)
	cmd := exec.CommandContext(ctx, "yt-dlp", cmdArgs...)
	// TODO: Capture into a buffer instead later to render progress in TUI
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}
