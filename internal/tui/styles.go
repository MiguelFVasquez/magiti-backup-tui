// Package tui para el manejo de la interfaz
package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	estiloTituloASCII = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7DA6C9")).
				Bold(true).
				MarginBottom(1)

	estiloEtiquetaMenu = lipgloss.NewStyle().
				Background(lipgloss.Color("#7DA6C9")).
				Foreground(lipgloss.Color("#1A1A1A")).
				Bold(true).
				Padding(0, 1).
				MarginBottom(1)

	estiloTitulo = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7DA6C9")).
			MarginBottom(1)

	estiloSeleccionado = lipgloss.NewStyle().
				Background(lipgloss.Color("#3A3A5A")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Padding(0, 1)

	estiloNormal = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C4C4C4")).
			Padding(0, 1)

	estiloActivo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF87")).
			Bold(true)

	estiloInactivo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F5F")).
			Bold(true)

	estiloInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)

	estiloAyuda = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	estiloError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F5F"))
)

var (
	estiloLogInfo  = lipgloss.NewStyle().Foreground(lipgloss.Color("#C4C4C4"))
	estiloLogWarn  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD866"))
	estiloLogError = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F5F"))
)

func estiloLineaLog(linea string) string {
	switch {
	case strings.Contains(linea, "[ERROR]"):
		return estiloLogError.Render(linea)
	case strings.Contains(linea, "[WARN]"):
		return estiloLogWarn.Render(linea)
	default:
		return estiloLogInfo.Render(linea)
	}
}

const tituloASCII = `
 __  __    _    ____ ___ _____ ___
|  \/  |  / \  / ___|_ _|_   _|_ _|
| |\/| | / _ \| |  _ | |  | |  | |
| |  | |/ ___ \ |_| || |  | |  | |
|_|  |_/_/   \_\____|___| |_| |___|
`
