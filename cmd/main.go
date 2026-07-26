package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MiguelFVasquez/magiti-backup-tui/internal/tui"
)

func main() {
	p := tea.NewProgram(tui.NuevoModelo())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error al iniciar la TUI:", err)
		os.Exit(1)
	}
}
