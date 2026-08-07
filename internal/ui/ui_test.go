package ui

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevinle-00/fornax/internal/job"
	"github.com/kevinle-00/fornax/internal/queue"
)

type stubDownloader struct {
	err error
}

func (s *stubDownloader) Download(ctx context.Context, url, outputPath, quality string, onProgress func(float64)) error {
	return s.err
}

type stubEncoder struct{}

func (s *stubEncoder) Encode(ctx context.Context, inputPath, outputPath string, onProgress func(float64)) error {
	return nil
}

func newInputModel(t *testing.T, selected string, step int, value string, capacity int) Model {
	t.Helper()

	q := queue.New(capacity)
	t.Cleanup(q.Close)

	input := textinput.New()
	input.SetValue(value)

	model := NewModel(q, &stubDownloader{}, &stubEncoder{})
	model.screen = InputScreen
	model.selected = selected
	model.inputStep = step
	model.input = input
	model.commandInputs = map[string]string{}
	return model
}

func TestUpdateInputScreenRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		selected string
		step     int
		value    string
	}{
		{name: "invalid URL", selected: "Download", step: 0, value: "not-a-url"},
		{name: "missing input file", selected: "Encode", step: 0, value: filepath.Join(t.TempDir(), "missing.mp4")},
		{name: "missing output directory", selected: "Process", step: 1, value: filepath.Join(t.TempDir(), "missing")},
		{name: "empty format", selected: "Encode", step: 1, value: "   "},
		{name: "unsafe format", selected: "Encode", step: 1, value: "../mp4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := newInputModel(t, tt.selected, tt.step, tt.value, 1)

			updated, cmd := updateInputScreen(model, tea.KeyMsg{Type: tea.KeyEnter})

			if cmd != nil {
				t.Error("expected invalid input to return no command")
			}
			if updated.screen != InputScreen {
				t.Errorf("expected to remain on input screen, got %s", updated.screen)
			}
			if updated.inputStep != tt.step {
				t.Errorf("expected to remain on step %d, got %d", tt.step, updated.inputStep)
			}
			if updated.errorMessage == "" {
				t.Error("expected validation error message")
			}
			if !strings.Contains(updated.viewInput(), "Error: "+updated.errorMessage) {
				t.Error("expected input view to show validation error")
			}
			if len(updated.queue.Jobs()) != 0 {
				t.Error("expected invalid input not to queue a job")
			}
		})
	}
}

func TestUpdateInputScreenAdvancesAfterValidInput(t *testing.T) {
	model := newInputModel(t, "Download", 0, "https://example.com/video", 1)

	updated, cmd := updateInputScreen(model, tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Error("expected intermediate input to return no command")
	}
	if updated.inputStep != 1 {
		t.Errorf("expected input step 1, got %d", updated.inputStep)
	}
	if got := updated.commandInputs["url"]; got != "https://example.com/video" {
		t.Errorf("expected URL to be stored, got %q", got)
	}
	if updated.input.Value() != "" {
		t.Errorf("expected input to be cleared, got %q", updated.input.Value())
	}
	if updated.errorMessage != "" {
		t.Errorf("expected no error message, got %q", updated.errorMessage)
	}
}

func TestUpdateInputScreenQueuesCompletedForm(t *testing.T) {
	model := newInputModel(t, "Download", 2, "", 1)
	model.commandInputs = map[string]string{
		"url":    "https://example.com/video",
		"output": t.TempDir(),
	}

	updated, cmd := updateInputScreen(model, tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Error("expected completed form to start dashboard polling")
	}
	if updated.screen != DashboardScreen {
		t.Errorf("expected dashboard screen, got %s", updated.screen)
	}
	if len(updated.queue.Jobs()) != 1 {
		t.Errorf("expected one queued job, got %d", len(updated.queue.Jobs()))
	}
	if updated.errorMessage != "" {
		t.Errorf("expected no error message, got %q", updated.errorMessage)
	}
}

func TestUpdateInputScreenShowsQueueError(t *testing.T) {
	model := newInputModel(t, "Download", 2, "best", 1)
	existing := job.NewDownloadJob(job.DownloadInputs{}, &stubDownloader{})
	if err := model.queue.Enqueue(existing); err != nil {
		t.Fatalf("failed to fill queue: %v", err)
	}
	model.commandInputs = map[string]string{
		"url":    "https://example.com/video",
		"output": t.TempDir(),
	}

	updated, cmd := updateInputScreen(model, tea.KeyMsg{Type: tea.KeyEnter})

	if cmd != nil {
		t.Error("expected queue error to return no command")
	}
	if updated.screen != InputScreen {
		t.Errorf("expected to remain on input screen, got %s", updated.screen)
	}
	if !strings.Contains(updated.errorMessage, queue.ErrQueueFull.Error()) {
		t.Errorf("expected queue full error, got %q", updated.errorMessage)
	}
	if len(updated.queue.Jobs()) != 1 {
		t.Errorf("expected queue to still contain one job, got %d", len(updated.queue.Jobs()))
	}
}

func TestUpdateInputScreenEscapeReturnsToMenu(t *testing.T) {
	model := newInputModel(t, "Download", 1, "output", 1)
	model.errorMessage = "invalid input"

	updated, cmd := updateInputScreen(model, tea.KeyMsg{Type: tea.KeyEsc})

	if cmd != nil {
		t.Error("expected escape to return no command")
	}
	if updated.screen != MenuScreen {
		t.Errorf("expected menu screen, got %s", updated.screen)
	}
	if updated.errorMessage != "" {
		t.Errorf("expected error message to be cleared, got %q", updated.errorMessage)
	}
}

func TestUpdateDashboardScreenShowsRequeueError(t *testing.T) {
	q := queue.New(1)
	t.Cleanup(q.Close)

	failedJob := job.NewDownloadJob(job.DownloadInputs{}, &stubDownloader{err: errors.New("download failed")})
	if err := failedJob.Execute(context.Background()); err == nil {
		t.Fatal("expected test job to fail")
	}
	if err := q.Enqueue(failedJob); err != nil {
		t.Fatalf("failed to fill queue: %v", err)
	}

	model := NewModel(q, &stubDownloader{}, &stubEncoder{})
	model.screen = DashboardScreen
	updated, cmd := updateDashboardScreen(model, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})

	if cmd != nil {
		t.Error("expected requeue error to return no command")
	}
	if updated.screen != DashboardScreen {
		t.Errorf("expected to remain on dashboard, got %s", updated.screen)
	}
	if !strings.Contains(updated.errorMessage, queue.ErrQueueFull.Error()) {
		t.Errorf("expected queue full error, got %q", updated.errorMessage)
	}
	if !strings.Contains(updated.viewDashboard(), "Error: "+updated.errorMessage) {
		t.Error("expected dashboard to show requeue error")
	}
}
