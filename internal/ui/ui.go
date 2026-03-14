// Package ui
package ui

import (
	"fmt"
	"strings"
	"time"

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
	queue         *queue.JobQueue
	downloader    download.Downloader
	encoder       encode.Encoder
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

func NewModel(queue *queue.JobQueue, downloader download.Downloader, encoder encode.Encoder) Model {
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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.screen {
		case MenuScreen:
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
		case InputScreen:

			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "enter":
				currentStep := stepDefinitions[m.selected][m.inputStep]
				m.commandInputs[currentStep.key] = m.input.Value()
				m.inputStep++
				if m.inputStep == len(stepDefinitions[m.selected]) {
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

		case DashboardScreen:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
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

func (m Model) viewDashboard() string {
	var s strings.Builder
	s.WriteString("\nFornax Dashboard\n\n")

	jobs := m.queue.GetJobs()
	for _, job := range jobs {
		fmt.Fprintf(&s, "Job: %s\nStatus: %s\n", job.GetID(), job.GetStatus())
	}
	return s.String()
}

type tickMsg time.Time

// Sends a tickMsg after a delay
func tickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
