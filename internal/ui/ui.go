// Package ui
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Screen string

const (
	MenuScreen      Screen = "menu"
	InputScreen     Screen = "input"
	DashboardScreen Screen = "dashboard"
)

type Model struct {
	choices  []string
	cursor   int
	selected string
	screen   Screen
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
			}
		case InputScreen:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
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
	var s string

	switch m.selected {
	case "Download":
		s = "\nDownload Input Menu\n\n"
		s += "URL: "

	case "Encode":
		s = "\nEncode Input Menu\n\n"
		s += "File: "

	case "Process":
		s = "\nProcess Input Menu\n\n"
		s += "URL: "
	}

	return s
}
