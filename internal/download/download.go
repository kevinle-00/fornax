// Package download runs yt-dlp downloads.
package download

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type Downloader interface {
	Download(ctx context.Context, url, outputPath, quality string, onProgress func(float64)) error
}

type YtDLP struct{}

func New() *YtDLP {
	return &YtDLP{}
}

func (y *YtDLP) Download(ctx context.Context, url, outputPath, quality string, onProgress func(float64)) error {
	cmdArgs := []string{}
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
	cmdArgs = append(cmdArgs, "--newline", "--no-color")
	cmdArgs = append(cmdArgs, url)
	cmd := exec.CommandContext(ctx, "yt-dlp", cmdArgs...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("read yt-dlp output: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start yt-dlp: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if progress, ok := parseProgress(scanner.Text()); ok && onProgress != nil {
			onProgress(progress)
		}
	}
	scanErr := scanner.Err()
	waitErr := cmd.Wait()
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("yt-dlp canceled: %w", err)
	}
	if scanErr != nil {
		return fmt.Errorf("read yt-dlp progress: %w", scanErr)
	}
	if waitErr != nil {
		return toolError("yt-dlp", waitErr, stderr.String())
	}

	return nil
}

func parseProgress(line string) (float64, bool) {
	if !strings.HasPrefix(line, "[download]") {
		return 0, false
	}

	for _, field := range strings.Fields(line) {
		percentText, found := strings.CutSuffix(field, "%")
		if !found {
			continue
		}
		percent, err := strconv.ParseFloat(percentText, 64)
		if err != nil || math.IsNaN(percent) || math.IsInf(percent, 0) {
			continue
		}
		percent = max(0, min(percent, 100))
		return percent / 100, true
	}

	return 0, false
}

func toolError(name string, err error, stderr string) error {
	details := strings.Join(strings.Fields(stderr), " ")
	if details == "" {
		return fmt.Errorf("%s failed: %w", name, err)
	}
	return fmt.Errorf("%s failed: %w: %s", name, err, details)
}
