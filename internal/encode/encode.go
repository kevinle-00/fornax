// Package encode is a ffmpeg wrapper
package encode

import (
	"context"
	"os"
	"os/exec"
)

type Encoder interface {
	Encode(ctx context.Context, inputPath, outputPath string) error
}

type FFmpeg struct{}

func New() *FFmpeg {
	return &FFmpeg{}
}

func (f *FFmpeg) Encode(ctx context.Context, inputPath, outputPath string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg", "-i", inputPath, outputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
