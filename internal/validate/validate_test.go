package validate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kevinle-00/fornax/internal/validate"
)

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "valid URL", url: "https://example.com/video", wantErr: false},
		{name: "empty URL", url: "", wantErr: true},
		{name: "missing scheme", url: "example.com/video", wantErr: true},
		{name: "missing host", url: "https:///video", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.IsValidURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValidURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}

func TestIsValidInputPath(t *testing.T) {
	tempDir := t.TempDir()
	inputPath := filepath.Join(tempDir, "input.mp4")
	if err := os.WriteFile(inputPath, []byte("media"), 0o600); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{name: "existing file", path: inputPath, wantErr: false},
		{name: "empty path", path: "", wantErr: true},
		{name: "missing path", path: filepath.Join(tempDir, "missing.mp4"), wantErr: true},
		{name: "directory", path: tempDir, wantErr: true},
		{name: "invalid path", path: "\x00", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.IsValidInputPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValidInputPath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestIsValidOutputPath(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "output.mp4")
	if err := os.WriteFile(filePath, []byte("media"), 0o600); err != nil {
		t.Fatalf("failed to create output file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{name: "existing directory", path: tempDir, wantErr: false},
		{name: "empty path uses current directory", path: "", wantErr: false},
		{name: "missing directory", path: filepath.Join(tempDir, "missing"), wantErr: true},
		{name: "file", path: filePath, wantErr: true},
		{name: "invalid path", path: "\x00", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate.IsValidOutputPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValidOutputPath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}
