// Package download is a yt-dlp wrapper
package download

import (
	"context"
	"os"
	"os/exec"
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
		cmdArgs = append(cmdArgs, "-o", outputPath)
	}
	if quality != "" {
		cmdArgs = append(cmdArgs, "-f", quality)
	}
	cmdArgs = append(cmdArgs, url)
	cmd := exec.CommandContext(ctx, "yt-dlp", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
