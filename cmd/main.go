package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	mensaje string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	return m.mensaje + "\n\nPresiona 'q' para salir.\n"
}

func main() {
	inicial := model{mensaje: "MAGITI Backup TUI - Prueba de entorno OK"}

	p := tea.NewProgram(inicial)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error al iniciar la TUI:", err)
		os.Exit(1)
	}
}
