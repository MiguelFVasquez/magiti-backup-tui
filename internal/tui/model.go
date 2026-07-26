package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/MiguelFVasquez/magiti-backup-tui/internal/config"
	"github.com/MiguelFVasquez/magiti-backup-tui/internal/daemon"
)

type vista int

const (
	vistaMenu vista = iota
	vistaEstado
	vistaLogs
)

var opcionesMenu = []string{
	"Estado del servicio",
	"Ver logs en vivo",
	"Historial de commits",
	"Gestionar carpetas vigiladas",
}

type model struct {
	vista          vista
	cursorMenu     int
	estadoServicio daemon.Estado
	mensajeAccion  string
	errorAccion    string
	cargando       bool
	ancho          int
	alto           int

	// Estado de la pantalla de logs
	logsLineas []string
	logsError  string
}

func NuevoModelo() model {
	return model{
		vista: vistaMenu,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

// ---------- Mensajes internos ----------

type (
	estadoActualizadoMsg daemon.Estado
	accionCompletadaMsg  struct {
		mensaje string
		err     error
	}
)

type logsCargadosMsg struct {
	lineas []string
	err    error
}

type tickLogsMsg struct{}

const (
	maxLineasLog          = 30
	intervaloRefrescoLogs = 3 * time.Second
)

func leerLogsCmd() tea.Cmd {
	return func() tea.Msg {
		ruta, err := config.RutaLogDaemon()
		if err != nil {
			return logsCargadosMsg{err: err}
		}
		lineas, err := daemon.LeerUltimasLineas(ruta, maxLineasLog)
		if err != nil {
			return logsCargadosMsg{err: err}
		}
		return logsCargadosMsg{lineas: lineas}
	}
}

func tickLogsCmd() tea.Cmd {
	return tea.Tick(intervaloRefrescoLogs, func(t time.Time) tea.Msg {
		return tickLogsMsg{}
	})
}

func consultarEstadoCmd() tea.Cmd {
	return func() tea.Msg {
		return estadoActualizadoMsg(daemon.ConsultarEstado())
	}
}

func ejecutarAccionCmd(accion func() error, nombreAccion string) tea.Cmd {
	return func() tea.Msg {
		err := accion()
		if err != nil {
			return accionCompletadaMsg{err: err}
		}
		return accionCompletadaMsg{mensaje: nombreAccion + " ejecutado correctamente."}
	}
}

// ---------- Update ----------

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch m.vista {
		case vistaMenu:
			return m.updateMenu(msg)
		case vistaEstado:
			return m.updateEstado(msg)
		}

	case estadoActualizadoMsg:
		m.estadoServicio = daemon.Estado(msg)
		m.cargando = false
		return m, nil

	case accionCompletadaMsg:
		m.cargando = false
		if msg.err != nil {
			m.errorAccion = msg.err.Error()
			m.mensajeAccion = ""
		} else {
			m.mensajeAccion = msg.mensaje
			m.errorAccion = ""
		}
		// Refresca el estado automáticamente tras cualquier acción.
		return m, tea.Batch(consultarEstadoCmd(), esperarYLimpiarMsg())

	case limpiarMensajeMsg:
		m.mensajeAccion = ""
		m.errorAccion = ""
		return m, nil

	case tea.WindowSizeMsg:
		m.ancho = msg.Width
		m.alto = msg.Height
		return m, nil

	case logsCargadosMsg:
		if msg.err != nil {
			m.logsError = msg.err.Error()
		} else {
			m.logsLineas = msg.lineas
			m.logsError = ""
		}

	case tickLogsMsg:
		// Solo seguimos refrescando si el usuario sigue en la pantalla de logs.
		if m.vista != vistaLogs {
			return m, nil
		}
		return m, tea.Batch(leerLogsCmd(), tickLogsCmd())

		return m, nil

	}

	return m, nil
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursorMenu > 0 {
			m.cursorMenu--
		}
	case "down", "j":
		if m.cursorMenu < len(opcionesMenu)-1 {
			m.cursorMenu++
		}
	case "enter":
		switch m.cursorMenu {
		case 0:
			m.vista = vistaEstado
			m.cargando = true
			return m, consultarEstadoCmd()
		case 1:
			m.vista = vistaLogs
			return m, tea.Batch(leerLogsCmd(), tickLogsCmd())
		default:
			// Las demás pantallas se implementan en los siguientes pasos.
		}
	}
	return m, nil
}

func (m model) updateEstado(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		m.vista = vistaMenu
		m.mensajeAccion = ""
		m.errorAccion = ""
		return m, nil
	case "i":
		m.cargando = true
		return m, ejecutarAccionCmd(daemon.Iniciar, "Inicio")
	case "d":
		m.cargando = true
		return m, ejecutarAccionCmd(daemon.Detener, "Detención")
	case "r":
		m.cargando = true
		return m, ejecutarAccionCmd(daemon.Reiniciar, "Reinicio")
	}
	return m, nil
}

func (m model) updateLogs(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		m.vista = vistaMenu
		return m, nil
	}
	return m, nil
}

// esperarYLimpiarMsg limpia el mensaje de acción tras unos segundos, para
// que no quede pegado indefinidamente en pantalla.
type limpiarMensajeMsg struct{}

func esperarYLimpiarMsg() tea.Cmd {
	return tea.Tick(4*time.Second, func(t time.Time) tea.Msg {
		return limpiarMensajeMsg{}
	})
}

// ---------- View ----------

func (m model) View() string {
	switch m.vista {
	case vistaEstado:
		return m.viewEstado()
	case vistaLogs:
		return m.viewLogs()
	default:
		return m.viewMenu()
	}
}

func (m model) viewMenu() string {
	contenido := estiloTituloASCII.Render(tituloASCII) + "\n"
	contenido += estiloEtiquetaMenu.Render("MENÚ PRINCIPAL") + "\n\n"

	for i, opcion := range opcionesMenu {
		if i == m.cursorMenu {
			contenido += estiloSeleccionado.Render(opcion) + "\n"
		} else {
			contenido += estiloNormal.Render(opcion) + "\n"
		}
	}

	estado := "servicio no consultado aún"
	if m.estadoServicio.Activo {
		estado = "servicio activo"
	}
	contenido += "\n" + estiloInfo.Render("Estado: "+estado) + "\n"
	contenido += estiloAyuda.Render("↑/↓ navegar · Enter seleccionar · q salir")

	return lipgloss.Place(
		m.ancho, m.alto,
		lipgloss.Center, lipgloss.Center,
		contenido,
	)
}

func (m model) viewEstado() string {
	s := estiloTituloASCII.Render("Estado del Servicio") + "\n\n"

	if m.cargando {
		s += "Consultando...\n"
		return s
	}

	if m.estadoServicio.Activo {
		s += "Estado: " + estiloActivo.Render("● ACTIVO") + "\n\n"
	} else {
		s += "Estado: " + estiloInactivo.Render("● INACTIVO") + "\n\n"
	}

	if m.mensajeAccion != "" {
		s += estiloActivo.Render(m.mensajeAccion) + "\n\n"
	}
	if m.errorAccion != "" {
		s += estiloError.Render("Error: "+m.errorAccion) + "\n\n"
	}

	s += estiloAyuda.Render("i iniciar · d detener · r reiniciar · esc volver al menú")
	return s
}

func (m model) viewLogs() string {
	s := estiloTitulo.Render("Logs del Daemon") + "\n\n"

	if m.logsError != "" {
		s += estiloError.Render("Error al leer el log: "+m.logsError) + "\n"
		s += estiloAyuda.Render("\nesc volver al menú")
		return s
	}

	if len(m.logsLineas) == 0 {
		s += "Sin registros aún.\n"
	} else {
		for _, linea := range m.logsLineas {
			s += estiloLineaLog(linea) + "\n"
		}
	}

	s += estiloAyuda.Render("\nActualizando cada 3s · esc volver al menú")
	return s
}
