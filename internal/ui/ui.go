// Package ui
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
}

var stepDefinitions = map[string][]inputStep{
	"Download": {
		{key: "url", placeholder: "Enter URL..."},
		{key: "output", placeholder: "Enter output directory path..."},
		{key: "quality", placeholder: "Enter video quality..."},
	},

	"Encode": {
		{key: "input", placeholder: "Enter input file path..."},
		{key: "output", placeholder: "Enter output directory path..."},
	},

	"Process": {
		{key: "url", placeholder: "Enter URL..."},
		{key: "output", placeholder: "Enter output directory path..."},
		{key: "format", placeholder: "Enter media format..."},
		{key: "quality", placeholder: "Enter video quality..."},
	},
}

func NewModel() Model {
	return Model{choices: []string{
		"Download", "Encode", "Process",
	}, screen: MenuScreen}
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
					m.screen = MenuScreen
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
	}
	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case MenuScreen:
		return m.viewMenu()

	case InputScreen:
		return m.viewInput()
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
