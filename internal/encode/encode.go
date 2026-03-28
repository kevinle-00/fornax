// Package encode is a ffmpeg wrapper
package encode

import (
	"context"
	"io"
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
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}
