package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/kevinle-00/fornax/internal/download"
	"github.com/kevinle-00/fornax/internal/encode"
)

type mockDownloader struct {
	mu         sync.Mutex
	calls      []string
	errors     map[string]error
	createFile bool
}

func (m *mockDownloader) Download(ctx context.Context, url, outputPath, quality string, onProgress func(float64)) error {
	m.mu.Lock()
	m.calls = append(m.calls, url)
	err := m.errors[url]
	m.mu.Unlock()
	if err != nil {
		return err
	}
	if m.createFile {
		if err := os.WriteFile(filepath.Join(filepath.Dir(outputPath), "video.webm"), []byte("media"), 0o600); err != nil {
			return err
		}
	}
	return nil
}

func (m *mockDownloader) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

type mockEncoder struct {
	mu    sync.Mutex
	calls []string
	err   error
}

func (m *mockEncoder) Encode(ctx context.Context, inputPath, outputPath string, onProgress func(float64)) error {
	m.mu.Lock()
	m.calls = append(m.calls, inputPath)
	m.mu.Unlock()
	return m.err
}

func (m *mockEncoder) callCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.calls)
}

func executeCommand(t *testing.T, downloader download.Downloader, encoder encode.Encoder, args ...string) (string, error) {
	t.Helper()
	root := newRootCommand(downloader, encoder)
	var output bytes.Buffer
	root.SetOut(&output)
	root.SetErr(&output)
	root.SetArgs(args)
	err := root.ExecuteContext(context.Background())
	return output.String(), err
}

func TestDownloadCommandSupportsMoreThanQueueCapacity(t *testing.T) {
	downloader := &mockDownloader{}
	args := []string{"download", "--output", t.TempDir()}
	for i := range 12 {
		args = append(args, fmt.Sprintf("https://example.com/video/%d", i))
	}

	output, err := executeCommand(t, downloader, &mockEncoder{}, args...)

	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}
	if downloader.callCount() != 12 {
		t.Errorf("expected 12 downloads, got %d", downloader.callCount())
	}
	if count := strings.Count(output, "DONE https://example.com/video/"); count != 12 {
		t.Errorf("expected 12 success lines, got %d\n%s", count, output)
	}
}

func TestDownloadCommandValidatesAllURLsBeforeStarting(t *testing.T) {
	downloader := &mockDownloader{}

	_, err := executeCommand(t, downloader, &mockEncoder{}, "download",
		"https://example.com/video", "not-a-url", "--output", t.TempDir())

	if err == nil {
		t.Fatal("expected command to fail validation")
	}
	if downloader.callCount() != 0 {
		t.Errorf("expected no downloads to start, got %d", downloader.callCount())
	}
}

func TestDownloadCommandReportsFailedJobs(t *testing.T) {
	failedURL := "https://example.com/failed"
	downloader := &mockDownloader{errors: map[string]error{failedURL: errors.New("download exploded")}}

	output, err := executeCommand(t, downloader, &mockEncoder{}, "download",
		"https://example.com/ok", failedURL, "--output", t.TempDir())

	if err == nil || !strings.Contains(err.Error(), "1 of 2 jobs failed") {
		t.Fatalf("expected aggregate failure, got %v", err)
	}
	if !strings.Contains(output, "DONE https://example.com/ok") {
		t.Errorf("expected successful result, got\n%s", output)
	}
	if !strings.Contains(output, "FAILED "+failedURL+": download exploded") {
		t.Errorf("expected failed result, got\n%s", output)
	}
}

func TestEncodeCommandProcessesLocalFiles(t *testing.T) {
	inputDir := t.TempDir()
	inputs := []string{
		filepath.Join(inputDir, "one.webm"),
		filepath.Join(inputDir, "two.webm"),
	}
	for _, input := range inputs {
		if err := os.WriteFile(input, []byte("media"), 0o600); err != nil {
			t.Fatalf("failed to create input: %v", err)
		}
	}
	encoder := &mockEncoder{}

	output, err := executeCommand(t, &mockDownloader{}, encoder, "encode",
		inputs[0], inputs[1], "--format", "mp4", "--output", t.TempDir())

	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}
	if encoder.callCount() != 2 {
		t.Errorf("expected 2 encodes, got %d", encoder.callCount())
	}
	for _, input := range inputs {
		if !strings.Contains(output, "DONE "+input) {
			t.Errorf("expected result for %s, got\n%s", input, output)
		}
	}
}

func TestEncodeCommandRequiresFormat(t *testing.T) {
	input := filepath.Join(t.TempDir(), "input.webm")
	if err := os.WriteFile(input, []byte("media"), 0o600); err != nil {
		t.Fatalf("failed to create input: %v", err)
	}
	encoder := &mockEncoder{}

	_, err := executeCommand(t, &mockDownloader{}, encoder, "encode", input)

	if err == nil || !strings.Contains(err.Error(), "format is required") {
		t.Fatalf("expected required format error, got %v", err)
	}
	if encoder.callCount() != 0 {
		t.Errorf("expected no encodes to start, got %d", encoder.callCount())
	}
}

func TestProcessCommandDownloadsAndEncodes(t *testing.T) {
	downloader := &mockDownloader{createFile: true}
	encoder := &mockEncoder{}
	url := "https://example.com/video"

	output, err := executeCommand(t, downloader, encoder, "process", url,
		"--format", "mp4", "--output", t.TempDir())

	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}
	if downloader.callCount() != 1 || encoder.callCount() != 1 {
		t.Errorf("expected one download and encode, got %d and %d", downloader.callCount(), encoder.callCount())
	}
	if !strings.Contains(output, "DONE "+url) {
		t.Errorf("expected success result, got\n%s", output)
	}
}
