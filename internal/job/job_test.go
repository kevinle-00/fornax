package job_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kevinle-00/fornax/internal/job"
)

// mockDownloader implements download.Downloader.
type mockDownloader struct {
	err error
}

func (m *mockDownloader) Download(ctx context.Context, url, outputPath, quality string) error {
	return m.err
}

// mockEncoder implements encode.Encoder.
type mockEncoder struct {
	err error
}

func (m *mockEncoder) Encode(ctx context.Context, inputPath, outputPath string) error {
	return m.err
}

func TestBaseJob_GetID(t *testing.T) {
	dl := job.NewDownloadJob(job.DownloadInputs{
		URL:             "https://example.com/video",
		OutputDirectory: "/tmp",
		Quality:         "best",
	}, &mockDownloader{})

	id := dl.GetID()
	if id == "" {
		t.Error("expected non-empty ID, got empty string")
	}
}

func TestBaseJob_SetStatusGetStatus(t *testing.T) {
	dl := job.NewDownloadJob(job.DownloadInputs{}, &mockDownloader{})

	if got := dl.GetStatus(); got != job.StatusPending {
		t.Errorf("expected status %q, got %q", job.StatusPending, got)
	}

	dl.SetStatus(job.StatusProcessing)
	if got := dl.GetStatus(); got != job.StatusProcessing {
		t.Errorf("expected status %q, got %q", job.StatusProcessing, got)
	}
}

func TestBaseJob_SetErrorGetError(t *testing.T) {
	dl := job.NewDownloadJob(job.DownloadInputs{}, &mockDownloader{})

	if got := dl.GetError(); got != nil {
		t.Errorf("expected nil error, got %v", got)
	}

	want := errors.New("something went wrong")
	dl.SetError(want)
	if got := dl.GetError(); got != want {
		t.Errorf("expected error %v, got %v", want, got)
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

			if got := dj.GetStatus(); got != job.StatusPending {
				t.Fatalf("expected initial status %q, got %q", job.StatusPending, got)
			}

			err := dj.Execute(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("expected error=%v, got %v", tt.wantErr, err)
			}

			if got := dj.GetStatus(); got != tt.expectedStatus {
				t.Errorf("expected status %q, got %q", tt.expectedStatus, got)
			}

			if tt.wantErr {
				if got := dj.GetError(); got == nil {
					t.Error("expected non-nil error from GetError, got nil")
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
				InputPath:  "/tmp/input.mp4",
				OutputPath: "/tmp/output.mp3",
			}, &mockEncoder{err: tt.encErr})

			if got := ej.GetStatus(); got != job.StatusPending {
				t.Fatalf("expected initial status %q, got %q", job.StatusPending, got)
			}

			err := ej.Execute(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("expected error=%v, got %v", tt.wantErr, err)
			}

			if got := ej.GetStatus(); got != tt.expectedStatus {
				t.Errorf("expected status %q, got %q", tt.expectedStatus, got)
			}

			if tt.wantErr {
				if got := ej.GetError(); got == nil {
					t.Error("expected non-nil error from GetError, got nil")
				}
			}
		})
	}
}
