package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kevinle-00/fornax/internal/ui"
)

func main() {
	if _, err := tea.NewProgram(ui.NewModel()).Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	/*
		if err := cmd.Execute(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	*/
}
