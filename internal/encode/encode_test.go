package encode

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
		duration     float64
		wantProgress float64
		wantOK       bool
	}{
		{name: "progress", line: "out_time_us=5000000", duration: 10, wantProgress: 0.5, wantOK: true},
		{name: "complete", line: "out_time_us=10000000", duration: 10, wantProgress: 1, wantOK: true},
		{name: "clamps high value", line: "out_time_us=12000000", duration: 10, wantProgress: 1, wantOK: true},
		{name: "clamps low value", line: "out_time_us=-1000", duration: 10, wantProgress: 0, wantOK: true},
		{name: "ignores malformed time", line: "out_time_us=unknown", duration: 10, wantOK: false},
		{name: "ignores NaN", line: "out_time_us=NaN", duration: 10, wantOK: false},
		{name: "ignores other keys", line: "progress=continue", duration: 10, wantOK: false},
		{name: "ignores invalid duration", line: "out_time_us=1000", duration: 0, wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseProgress(tt.line, tt.duration)
			if ok != tt.wantOK {
				t.Fatalf("parseProgress(%q) ok = %v, want %v", tt.line, ok, tt.wantOK)
			}
			if ok && math.Abs(got-tt.wantProgress) > 0.0001 {
				t.Errorf("parseProgress(%q) = %v, want %v", tt.line, got, tt.wantProgress)
			}
		})
	}
}

func TestEncodeReportsProgress(t *testing.T) {
	requireShell(t)
	dir := t.TempDir()
	writeTool(t, dir, "ffprobe", "printf '10.0\\n'")
	writeTool(t, dir, "ffmpeg", "printf 'out_time_us=invalid\\n'\nprintf 'out_time_us=5000000\\n'")
	t.Setenv("PATH", dir)

	var progress []float64
	err := New().Encode(context.Background(), "input.webm", "output.mp4", func(value float64) {
		progress = append(progress, value)
	})

	if err != nil {
		t.Fatalf("expected encode to succeed, got %v", err)
	}
	if len(progress) != 1 || progress[0] != 0.5 {
		t.Errorf("expected progress [0.5], got %v", progress)
	}
}

func TestEncodeIncludesProbeError(t *testing.T) {
	requireShell(t)
	dir := t.TempDir()
	writeTool(t, dir, "ffprobe", "printf 'probe exploded\\n' >&2\nexit 1")
	t.Setenv("PATH", dir)

	err := New().Encode(context.Background(), "input.webm", "output.mp4", nil)

	if err == nil {
		t.Fatal("expected probe to fail")
	}
	if !strings.Contains(err.Error(), "ffprobe failed") || !strings.Contains(err.Error(), "probe exploded") {
		t.Errorf("expected useful ffprobe error, got %q", err)
	}
}

func TestEncodeIncludesFFmpegError(t *testing.T) {
	requireShell(t)
	dir := t.TempDir()
	writeTool(t, dir, "ffprobe", "printf '10.0\\n'")
	writeTool(t, dir, "ffmpeg", "printf 'encode exploded\\n' >&2\nexit 1")
	t.Setenv("PATH", dir)

	err := New().Encode(context.Background(), "input.webm", "output.mp4", nil)

	if err == nil {
		t.Fatal("expected encode to fail")
	}
	if !strings.Contains(err.Error(), "ffmpeg failed") || !strings.Contains(err.Error(), "encode exploded") {
		t.Errorf("expected useful ffmpeg error, got %q", err)
	}
}

func TestEncodeRespectsCancellation(t *testing.T) {
	requireShell(t)
	dir := t.TempDir()
	writeTool(t, dir, "ffprobe", "printf '10.0\\n'")
	writeTool(t, dir, "ffmpeg", "exec /bin/sleep 30")
	t.Setenv("PATH", dir)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	started := time.Now()
	err := New().Encode(ctx, "input.webm", "output.mp4", nil)

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
