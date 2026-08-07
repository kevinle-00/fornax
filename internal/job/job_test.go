package job_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kevinle-00/fornax/internal/job"
)

type mockDownloader struct {
	err           error
	progress      []float64
	createFiles   []string
	tempDir       string
	afterProgress func()
}

func (m *mockDownloader) Download(ctx context.Context, url, outputPath, quality string, onProgress func(float64)) error {
	m.tempDir = filepath.Dir(outputPath)
	for _, name := range m.createFiles {
		if err := os.WriteFile(filepath.Join(m.tempDir, name), []byte("media"), 0o600); err != nil {
			return err
		}
	}
	for _, value := range m.progress {
		onProgress(value)
		if m.afterProgress != nil {
			m.afterProgress()
		}
	}
	return m.err
}

type mockEncoder struct {
	err           error
	progress      []float64
	inputPath     string
	outputPath    string
	afterProgress func()
}

func (m *mockEncoder) Encode(ctx context.Context, inputPath, outputPath string, onProgress func(float64)) error {
	m.inputPath = inputPath
	m.outputPath = outputPath
	for _, value := range m.progress {
		onProgress(value)
		if m.afterProgress != nil {
			m.afterProgress()
		}
	}
	return m.err
}

func TestBaseJob_ID(t *testing.T) {
	dl := job.NewDownloadJob(job.DownloadInputs{
		URL:             "https://example.com/video",
		OutputDirectory: "/tmp",
		Quality:         "best",
	}, &mockDownloader{})

	id := dl.ID()
	if id == "" {
		t.Error("expected non-empty ID, got empty string")
	}
}

func TestDownloadJob_Execute(t *testing.T) {
	tests := []struct {
		name           string
		dlErr          error
		wantErr        bool
		expectedStatus job.Status
	}{
		{
			name:           "success transitions to done",
			dlErr:          nil,
			wantErr:        false,
			expectedStatus: job.StatusDone,
		},
		{
			name:           "failure transitions to failed",
			dlErr:          errors.New("download failed"),
			wantErr:        true,
			expectedStatus: job.StatusFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dj := job.NewDownloadJob(job.DownloadInputs{
				URL:             "https://example.com/video",
				OutputDirectory: "/tmp",
				Quality:         "best",
			}, &mockDownloader{err: tt.dlErr})

			if got := dj.Status(); got != job.StatusPending {
				t.Fatalf("expected initial status %q, got %q", job.StatusPending, got)
			}

			err := dj.Execute(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("expected error=%v, got %v", tt.wantErr, err)
			}

			if got := dj.Status(); got != tt.expectedStatus {
				t.Errorf("expected status %q, got %q", tt.expectedStatus, got)
			}
			if !tt.wantErr && dj.Progress() != 1 {
				t.Errorf("expected completed progress 1, got %v", dj.Progress())
			}

			if tt.wantErr {
				if got := dj.Error(); got == nil {
					t.Error("expected non-nil error from Error, got nil")
				}
			}
		})
	}
}

func TestEncodeJob_Execute(t *testing.T) {
	tests := []struct {
		name           string
		encErr         error
		wantErr        bool
		expectedStatus job.Status
	}{
		{
			name:           "success transitions to done",
			encErr:         nil,
			wantErr:        false,
			expectedStatus: job.StatusDone,
		},
		{
			name:           "failure transitions to failed",
			encErr:         errors.New("encode failed"),
			wantErr:        true,
			expectedStatus: job.StatusFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ej := job.NewEncodeJob(job.EncodeInputs{
				InputPath:       "/tmp/input.mp4",
				OutputDirectory: "/tmp",
				Format:          "mp3",
			}, &mockEncoder{err: tt.encErr})

			if got := ej.Status(); got != job.StatusPending {
				t.Fatalf("expected initial status %q, got %q", job.StatusPending, got)
			}

			err := ej.Execute(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("expected error=%v, got %v", tt.wantErr, err)
			}

			if got := ej.Status(); got != tt.expectedStatus {
				t.Errorf("expected status %q, got %q", tt.expectedStatus, got)
			}
			if !tt.wantErr && ej.Progress() != 1 {
				t.Errorf("expected completed progress 1, got %v", ej.Progress())
			}

			if tt.wantErr {
				if got := ej.Error(); got == nil {
					t.Error("expected non-nil error from Error, got nil")
				}
			}
		})
	}
}

func TestProcessJob_Execute(t *testing.T) {
	tests := []struct {
		name            string
		downloadErr     error
		encodeErr       error
		downloadedFiles []string
		wantErr         bool
		expectedStatus  job.Status
	}{
		{
			name:            "success transitions to done",
			downloadedFiles: []string{"video.webm"},
			expectedStatus:  job.StatusDone,
		},
		{
			name:            "download failure transitions to failed",
			downloadErr:     errors.New("download failed"),
			downloadedFiles: []string{"video.part"},
			wantErr:         true,
			expectedStatus:  job.StatusFailed,
		},
		{
			name:           "missing download transitions to failed",
			wantErr:        true,
			expectedStatus: job.StatusFailed,
		},
		{
			name:            "encode failure transitions to failed",
			encodeErr:       errors.New("encode failed"),
			downloadedFiles: []string{"video.webm"},
			wantErr:         true,
			expectedStatus:  job.StatusFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputDir := t.TempDir()
			downloader := &mockDownloader{
				err:         tt.downloadErr,
				createFiles: tt.downloadedFiles,
			}
			encoder := &mockEncoder{err: tt.encodeErr}
			processJob := job.NewProcessJob(job.ProcessInputs{
				URL:             "https://example.com/video",
				OutputDirectory: outputDir,
				Format:          "mp4",
				Quality:         "best",
			}, downloader, encoder)

			err := processJob.Execute(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("expected error=%v, got %v", tt.wantErr, err)
			}
			if got := processJob.Status(); got != tt.expectedStatus {
				t.Errorf("expected status %q, got %q", tt.expectedStatus, got)
			}
			if !tt.wantErr && processJob.Progress() != 1 {
				t.Errorf("expected completed progress 1, got %v", processJob.Progress())
			}
			if !tt.wantErr {
				if got := filepath.Base(encoder.inputPath); got != "video.webm" {
					t.Errorf("expected encoder input video.webm, got %s", got)
				}
				wantOutput := filepath.Join(outputDir, "video.mp4")
				if encoder.outputPath != wantOutput {
					t.Errorf("expected encoder output %s, got %s", wantOutput, encoder.outputPath)
				}
			}
			if tt.wantErr && processJob.Error() == nil {
				t.Error("expected non-nil error from Error, got nil")
			}
			if _, err := os.Stat(downloader.tempDir); !os.IsNotExist(err) {
				t.Errorf("expected temporary directory to be removed, got error %v", err)
			}
		})
	}
}

func TestProcessJob_ProgressAcrossPhases(t *testing.T) {
	var processJob *job.ProcessJob
	var observed []float64
	recordProgress := func() {
		observed = append(observed, processJob.Progress())
	}

	downloader := &mockDownloader{
		progress:      []float64{0, 1},
		createFiles:   []string{"video.webm"},
		afterProgress: recordProgress,
	}
	encoder := &mockEncoder{
		progress:      []float64{0, 1},
		afterProgress: recordProgress,
	}
	processJob = job.NewProcessJob(job.ProcessInputs{
		URL:             "https://example.com/video",
		OutputDirectory: t.TempDir(),
		Format:          "mp4",
	}, downloader, encoder)

	if err := processJob.Execute(context.Background()); err != nil {
		t.Fatalf("expected process job to succeed, got %v", err)
	}

	want := []float64{0, 0.5, 0.5, 1}
	if len(observed) != len(want) {
		t.Fatalf("expected %d progress updates, got %d: %v", len(want), len(observed), observed)
	}
	for i := range want {
		if observed[i] != want[i] {
			t.Errorf("progress update %d = %v, want %v", i, observed[i], want[i])
		}
	}
}
