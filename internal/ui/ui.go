// Package ui
package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kevinle-00/fornax/internal/download"
	"github.com/kevinle-00/fornax/internal/encode"
	"github.com/kevinle-00/fornax/internal/job"
	"github.com/kevinle-00/fornax/internal/queue"
)

type Screen string

const (
	MenuScreen      Screen = "menu"
	InputScreen     Screen = "input"
	DashboardScreen Screen = "dashboard"
)

type inputStep struct {
	key         string
	placeholder string
}

type Model struct {
	choices       []string
	cursor        int
	selected      string
	screen        Screen
	input         textinput.Model
	inputStep     int
	commandInputs map[string]string
	queue         *queue.Queue
	downloader    download.Downloader
	encoder       encode.Encoder
	dashCursor    int
}

var stepDefinitions = map[string][]inputStep{
	"Download": {
		{key: "url", placeholder: "Enter URL..."},
		{key: "output", placeholder: "Enter output directory path..."},
		{key: "quality", placeholder: "Enter video quality..."},
	},

	"Encode": {
		{key: "input", placeholder: "Enter input file path..."},
		{key: "format", placeholder: "Enter media format..."},
		{key: "output", placeholder: "Enter output directory path..."},
	},

	"Process": {
		{key: "url", placeholder: "Enter URL..."},
		{key: "output", placeholder: "Enter output directory path..."},
		{key: "format", placeholder: "Enter media format..."},
		{key: "quality", placeholder: "Enter video quality..."},
	},
}

func NewModel(queue *queue.Queue, downloader download.Downloader, encoder encode.Encoder) Model {
	return Model{
		choices: []string{
			"Download", "Encode", "Process",
		}, screen: MenuScreen,
		queue:      queue,
		downloader: downloader,
		encoder:    encoder,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func updateMenuScreen(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.choices)-1 {
			m.cursor++
		}
	case "enter", "space", "l":
		selected := m.choices[m.cursor]
		m.selected = selected
		m.screen = InputScreen
		m.commandInputs = map[string]string{}
		m.inputStep = 0
		ti := textinput.New()
		placeholder := stepDefinitions[m.selected][m.inputStep].placeholder
		ti.Placeholder = placeholder
		ti.Focus()
		m.input = ti
	}
	return m, nil
}

func createJob(m Model) job.Job {
	var newJob job.Job
	switch m.selected {
	case "Download":
		inputs := job.DownloadInputs{
			URL:             m.commandInputs["url"],
			OutputDirectory: m.commandInputs["output"],
			Quality:         m.commandInputs["quality"],
		}
		newJob = job.NewDownloadJob(inputs, m.downloader)

	case "Encode":
		inputs := job.EncodeInputs{
			InputPath:       m.commandInputs["input"],
			Format:          m.commandInputs["format"],
			OutputDirectory: m.commandInputs["output"],
		}
		newJob = job.NewEncodeJob(inputs, m.encoder)

	case "Process":
		inputs := job.ProcessInputs{
			URL:             m.commandInputs["url"],
			OutputDirectory: m.commandInputs["output"],
			Format:          m.commandInputs["format"],
			Quality:         m.commandInputs["quality"],
		}
		newJob = job.NewProcessJob(inputs, m.downloader, m.encoder)
	}
	return newJob
}

func updateInputScreen(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "enter":
		currentStep := stepDefinitions[m.selected][m.inputStep]
		m.commandInputs[currentStep.key] = m.input.Value()
		m.inputStep++

		if m.inputStep == len(stepDefinitions[m.selected]) {
			newJob := createJob(m)
			if err := m.queue.Enqueue(newJob); err != nil {
				// TODO: add error UI
				return m, tea.Quit
			}
			m.screen = DashboardScreen
			return m, tickCmd()
		} else {
			m.input.SetValue("")
			newPlaceholder := stepDefinitions[m.selected][m.inputStep].placeholder
			m.input.Placeholder = newPlaceholder
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.screen {
		case MenuScreen:
			return updateMenuScreen(m, msg)
		case InputScreen:
			return updateInputScreen(m, msg)
		case DashboardScreen:
			return updateDashboardScreen(m, msg)
		}

	case tickMsg:
		return m, tickCmd()
	}
	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case MenuScreen:
		return m.viewMenu()

	case InputScreen:
		return m.viewInput()

	case DashboardScreen:
		return m.viewDashboard()
	}

	return ""
}

func (m Model) viewMenu() string {
	var s strings.Builder
	s.WriteString("\nWelcome to fornax! Press q to quit\n\n")

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		fmt.Fprintf(&s, "%s %s\n", cursor, choice)

	}
	return s.String()
}

func (m Model) viewInput() string {
	var s strings.Builder
	fmt.Fprintf(&s, "\n%s Input Menu\n\n", m.selected)

	steps := stepDefinitions[m.selected]

	for i, step := range steps {
		if i < m.inputStep {
			fmt.Fprintf(&s, "%s: %s\n", step.key, m.commandInputs[step.key])
		} else if i == m.inputStep {
			fmt.Fprintf(&s, "%s: %s\n", step.key, m.input.View())
		}
	}

	return s.String()
}

func updateDashboardScreen(m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	jobs := m.queue.Jobs()
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.dashCursor > 0 {
			m.dashCursor--
		}
	case "down", "j":
		if m.dashCursor < len(jobs)-1 {
			m.dashCursor++
		}
	case "m":
		m.cursor = 0
		m.screen = MenuScreen
		return m, nil
	case "r":
		if len(jobs) > 0 {
			selected := jobs[m.dashCursor]
			if selected.Status() == job.StatusFailed {
				newJob := selected.Requeue()
				if err := m.queue.Enqueue(newJob); err != nil {
					// TODO: add error UI
					return m, tea.Quit
				}
			}
		}
	}
	return m, nil
}

func (m Model) viewDashboard() string {
	var s strings.Builder
	s.WriteString("\nFornax Dashboard (r: requeue failed | m: menu | q: quit)\n\n")

	jobs := m.queue.Jobs()
	for i, j := range jobs {
		cursor := " "
		if i == m.dashCursor {
			cursor = ">"
		}

		var jobType string
		switch j.(type) {
		case *job.DownloadJob:
			jobType = "Download"
		case *job.EncodeJob:
			jobType = "Encode"
		case *job.ProcessJob:
			jobType = "Process"
		}
		jobContent := fmt.Sprintf("%s | Job: %s | Type: %s | Status: %s", cursor, j.ID()[:8], jobType, j.Status())
		// Use longest status ("processing") to keep layout stable across status changes
		maxWidth := len(jobContent) + len(string(job.StatusProcessing)) - len(string(j.Status())) + len(" |")
		jobLine := fmt.Sprintf("%-*s|", maxWidth-1, jobContent)
		fmt.Fprintf(&s, "%s\n", jobLine)

		if j.Status() == job.StatusProcessing || j.Status() == job.StatusDone {
			// Width accounts for "  " indent prefix
			bar := progress.New(progress.WithDefaultGradient(), progress.WithWidth(maxWidth-2))
			fmt.Fprintf(&s, "\n  %s\n", bar.ViewAs(j.Progress()))
		}

		// TODO: need to make error messages useful for user
		if j.Status() == job.StatusFailed {
			errLine := fmt.Sprintf("  Error: %v", j.Error())
			fmt.Fprintf(&s, "%s\n", errLine)
			if len(errLine) > maxWidth {
				maxWidth = len(errLine)
			}
		}

		fmt.Fprintf(&s, "%s\n", strings.Repeat("─", maxWidth))
	}
	return s.String()
}

type tickMsg time.Time

// Sends a tickMsg after a delay
func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
