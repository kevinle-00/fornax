package download

import (
	"context"
	"errors"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestParseProgress(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		wantProgress float64
		wantOK       bool
	}{
		{name: "progress", line: "[download] 42.5% of 10MiB", wantProgress: 0.425, wantOK: true},
		{name: "complete", line: "[download] 100.0% of 10MiB", wantProgress: 1, wantOK: true},
		{name: "clamps high value", line: "[download] 120% of 10MiB", wantProgress: 1, wantOK: true},
		{name: "clamps low value", line: "[download] -5% of 10MiB", wantProgress: 0, wantOK: true},
		{name: "ignores malformed percentage", line: "[download] unknown% of 10MiB", wantOK: false},
		{name: "ignores NaN", line: "[download] NaN% of 10MiB", wantOK: false},
		{name: "ignores other output", line: "[info] 50% ready", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseProgress(tt.line)
			if ok != tt.wantOK {
				t.Fatalf("parseProgress(%q) ok = %v, want %v", tt.line, ok, tt.wantOK)
			}
			if ok && math.Abs(got-tt.wantProgress) > 0.0001 {
				t.Errorf("parseProgress(%q) = %v, want %v", tt.line, got, tt.wantProgress)
			}
		})
	}
}

func TestDownloadReportsProgress(t *testing.T) {
	requireShell(t)
	dir := t.TempDir()
	writeTool(t, dir, "yt-dlp", "printf '[download] invalid%% of 10MiB\\n'\nprintf '[download] 25.0%% of 10MiB\\n'")
	t.Setenv("PATH", dir)

	var progress []float64
	err := New().Download(context.Background(), "https://example.com/video", "", "", func(value float64) {
		progress = append(progress, value)
	})

	if err != nil {
		t.Fatalf("expected download to succeed, got %v", err)
	}
	if len(progress) != 1 || progress[0] != 0.25 {
		t.Errorf("expected progress [0.25], got %v", progress)
	}
}

func TestDownloadIncludesToolError(t *testing.T) {
	requireShell(t)
	dir := t.TempDir()
	writeTool(t, dir, "yt-dlp", "printf 'download exploded\\n' >&2\nexit 1")
	t.Setenv("PATH", dir)

	err := New().Download(context.Background(), "https://example.com/video", "", "", nil)

	if err == nil {
		t.Fatal("expected download to fail")
	}
	if !strings.Contains(err.Error(), "yt-dlp failed") || !strings.Contains(err.Error(), "download exploded") {
		t.Errorf("expected useful yt-dlp error, got %q", err)
	}
}

func TestDownloadRespectsCancellation(t *testing.T) {
	requireShell(t)
	dir := t.TempDir()
	writeTool(t, dir, "yt-dlp", "exec /bin/sleep 30")
	t.Setenv("PATH", dir)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	started := time.Now()
	err := New().Download(ctx, "https://example.com/video", "", "", nil)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected deadline exceeded, got %v", err)
	}
	if elapsed := time.Since(started); elapsed > 3*time.Second {
		t.Errorf("expected quick cancellation, took %v", elapsed)
	}
}

func requireShell(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("test requires a POSIX shell")
	}
}

func writeTool(t *testing.T, dir, name, body string) {
	t.Helper()
	path := filepath.Join(dir, name)
	content := "#!/bin/sh\n" + body + "\n"
	if err := os.WriteFile(path, []byte(content), 0o700); err != nil {
		t.Fatalf("failed to create %s: %v", name, err)
	}
}
