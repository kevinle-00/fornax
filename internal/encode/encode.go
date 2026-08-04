// Package encode is a ffmpeg wrapper
package encode

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"math"
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
	probeCmd := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-show_entries",
		"format=duration", "-of", "csv=p=0", inputPath)
	var probeStderr bytes.Buffer
	probeCmd.Stderr = &probeStderr

	output, err := probeCmd.Output()
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("ffprobe canceled: %w", ctxErr)
		}
		return toolError("ffprobe", err, probeStderr.String())
	}

	outputStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(outputStr, 64)
	if err != nil {
		return fmt.Errorf("failed to parse duration: %w", err)
	}
	if duration <= 0 || math.IsNaN(duration) || math.IsInf(duration, 0) {
		return fmt.Errorf("invalid media duration: %v", duration)
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", "-nostdin", "-n", "-v", "error", "-i", inputPath,
		"-progress", "pipe:1", "-nostats", outputPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("read ffmpeg output: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start ffmpeg: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if progress, ok := parseProgress(scanner.Text(), duration); ok && onProgress != nil {
			onProgress(progress)
		}
	}
	scanErr := scanner.Err()
	waitErr := cmd.Wait()
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("ffmpeg canceled: %w", err)
	}
	if scanErr != nil {
		return fmt.Errorf("read ffmpeg progress: %w", scanErr)
	}
	if waitErr != nil {
		return toolError("ffmpeg", waitErr, stderr.String())
	}

	return nil
}

func parseProgress(line string, duration float64) (float64, bool) {
	key, value, found := strings.Cut(line, "=")
	if !found || key != "out_time_us" || duration <= 0 {
		return 0, false
	}

	microseconds, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(microseconds) || math.IsInf(microseconds, 0) {
		return 0, false
	}
	progress := microseconds / (duration * 1_000_000)
	return max(0, min(progress, 1)), true
}

func toolError(name string, err error, stderr string) error {
	details := strings.Join(strings.Fields(stderr), " ")
	if details == "" {
		return fmt.Errorf("%s failed: %w", name, err)
	}
	return fmt.Errorf("%s failed: %w: %s", name, err, details)
}
