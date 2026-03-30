// Package encode is a ffmpeg wrapper
package encode

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

type Encoder interface {
	Encode(ctx context.Context, inputPath, outputPath string, onProgress func(float64)) error
}

type FFmpeg struct{}

func New() *FFmpeg {
	return &FFmpeg{}
}

func (f *FFmpeg) Encode(ctx context.Context, inputPath, outputPath string, onProgress func(float64)) error {
	// Get total duration for progress calculation
	// -v quiet: suppress logs, -show_entries format=duration: only output duration
	// -of csv=p=0: output as plain number (e.g., "213.043084")
	probeCmd := exec.CommandContext(ctx, "ffprobe", "-v", "quiet", "-show_entries",
		"format=duration", "-of", "csv=p=0", inputPath)

	output, err := probeCmd.Output()
	if err != nil {
		return err
	}

	outputStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(outputStr, 64)
	if err != nil {
		return fmt.Errorf("failed to parse duration: %w", err)
	}

	// -progress pipe:1: output machine readable progress (key=value pairs) to stdout
	cmd := exec.CommandContext(ctx, "ffmpeg", "-i", inputPath, "-progress", "pipe:1", outputPath)
	cmd.Stderr = io.Discard

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		line := scanner.Text()
		before, after, found := strings.Cut(line, "=")

		if found && before == "out_time_us" {
			vidTime, _ := strconv.ParseFloat(after, 64)
			onProgress(vidTime / (duration * 1000000))
		}
	}
	return cmd.Wait()
}
