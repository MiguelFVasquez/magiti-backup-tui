package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/MiguelFVasquez/magiti-backup-tui/internal/config"
	"github.com/MiguelFVasquez/magiti-backup-tui/internal/daemon"
)

type (
	vista        int
	modoCarpetas int
)

const (
	vistaMenu vista = iota
	vistaEstado
	vistaLogs
	vistaHistorial
	vistaCarpetas
	vistaAyuda
)

const (
	modoListaCarpetas modoCarpetas = iota
	modoAgregarCarpeta
	modoConfirmarEliminar
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

	// Estado de la pantalla de historial
	historialCommits  []daemon.CommitInfo
	historialError    string
	historialCargando bool

	// Estado de la pantalla de carpetas
	carpetasModo      modoCarpetas
	carpetas          map[string]string
	carpetasOrden     []string // para tener un orden estable al listar
	cursorCarpetas    int
	inputCarpeta      textinput.Model
	carpetasError     string
	carpetasMensaje   string
	carpetaAConfirmar string
	// Improve en UX
	logsViewport      viewport.Model
	historialViewport viewport.Model
	viewportListo     bool

	inputCarpetaAviso string
}

func NuevoModelo() model {
	ti := textinput.New()
	ti.Placeholder = "nombre-de-la-carpeta"
	ti.CharLimit = 50
	ti.Width = 30
	ti.Validate = validarCaracterCarpeta

	return model{
		vista:             vistaMenu,
		inputCarpeta:      ti,
		logsViewport:      viewport.New(0, 0),
		historialViewport: viewport.New(0, 0),
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

type historialCargadoMsg struct {
	commits []daemon.CommitInfo
	err     error
}

type carpetasCargadasMsg struct {
	carpetas map[string]string
	err      error
}

type carpetaAccionCompletadaMsg struct {
	mensaje string
	err     error
}

type limpiarAvisoCarpetaMsg struct{}

const (
	maxLineasLog          = 30
	intervaloRefrescoLogs = 3 * time.Second
)

const maxCommitsHistorial = 15

// Mensajes para la view de logs
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

func leerHistorialCmd() tea.Cmd {
	return func() tea.Msg {
		repoDir, err := config.RepoDirDaemon()
		if err != nil {
			return historialCargadoMsg{err: err}
		}
		commits, err := daemon.LeerHistorial(repoDir, maxCommitsHistorial)
		if err != nil {
			return historialCargadoMsg{err: err}
		}
		return historialCargadoMsg{commits: commits}
	}
}

// Mensajes para la view de estado del servicio
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

// Mensajes para la view de configuración de la gestión de carpetas
func leerCarpetasCmd() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.LeerConfigDaemon()
		if err != nil {
			return carpetasCargadasMsg{err: err}
		}
		return carpetasCargadasMsg{carpetas: cfg.Carpetas}
	}
}

func agregarCarpetaCmd(nombre string) tea.Cmd {
	return func() tea.Msg {
		if err := config.AgregarCarpeta(nombre); err != nil {
			return carpetaAccionCompletadaMsg{err: err}
		}
		if err := daemon.Reiniciar(); err != nil {
			return carpetaAccionCompletadaMsg{err: fmt.Errorf("carpeta agregada, pero falló el reinicio del servicio: %w", err)}
		}
		return carpetaAccionCompletadaMsg{mensaje: "Carpeta '" + nombre + "' agregada y servicio reiniciado."}
	}
}

func eliminarCarpetaCmd(nombre string) tea.Cmd {
	return func() tea.Msg {
		if err := config.EliminarCarpeta(nombre); err != nil {
			return carpetaAccionCompletadaMsg{err: err}
		}
		if err := daemon.Reiniciar(); err != nil {
			return carpetaAccionCompletadaMsg{err: fmt.Errorf("carpeta eliminada, pero falló el reinicio del servicio: %w", err)}
		}
		return carpetaAccionCompletadaMsg{mensaje: "Carpeta '" + nombre + "' eliminada y servicio reiniciado."}
	}
}

func limpiarAvisoCarpetaCmd() tea.Cmd {
	return tea.Tick(1500*time.Millisecond, func(t time.Time) tea.Msg {
		return limpiarAvisoCarpetaMsg{}
	})
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
		case vistaLogs:
			return m.updateLogs(msg)
		case vistaHistorial:
			return m.updateHistorial(msg)
		case vistaCarpetas:
			return m.updateCarpetas(msg)
		case vistaAyuda:
			return m.updateAyuda(msg)
		}
	case limpiarAvisoCarpetaMsg:
		m.inputCarpetaAviso = ""
		return m, nil

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

		// Reservamos ~6 líneas para título + ayuda en cada pantalla.
		altoContenido := msg.Height - 6
		if altoContenido < 3 {
			altoContenido = 3
		}
		m.logsViewport.Width = msg.Width
		m.logsViewport.Height = altoContenido
		m.historialViewport.Width = msg.Width
		m.historialViewport.Height = altoContenido
		m.viewportListo = true
		return m, nil

	case logsCargadosMsg:
		if msg.err != nil {
			m.logsError = msg.err.Error()
		} else {
			m.logsLineas = msg.lineas
			m.logsError = ""
			contenido := ""
			for _, l := range m.logsLineas {
				contenido += estiloLineaLog(l) + "\n"
			}
			m.logsViewport.SetContent(contenido)
			m.logsViewport.GotoBottom()
		}
		return m, nil

	case tickLogsMsg:
		// Solo seguimos refrescando si el usuario sigue en la pantalla de logs.
		if m.vista != vistaLogs {
			return m, nil
		}
		return m, tea.Batch(leerLogsCmd(), tickLogsCmd())

	case historialCargadoMsg:
		m.historialCargando = false
		if msg.err != nil {
			m.historialError = msg.err.Error()
		} else {
			m.historialCommits = msg.commits
			m.historialError = ""
			contenido := ""
			for _, c := range m.historialCommits {
				contenido += estiloFechaCommit.Render(c.Fecha) + "  " + estiloMensajeCommit.Render(c.Mensaje) + "\n"
			}
			m.historialViewport.SetContent(contenido)
		}
		return m, nil

	case carpetasCargadasMsg:
		if msg.err != nil {
			m.carpetasError = msg.err.Error()
		} else {
			m.carpetas = msg.carpetas
			m.carpetasOrden = ordenarClaves(msg.carpetas)
			m.carpetasError = ""
			if m.cursorCarpetas >= len(m.carpetasOrden) {
				m.cursorCarpetas = 0
			}
		}
		return m, nil

	case carpetaAccionCompletadaMsg:
		m.carpetasModo = modoListaCarpetas
		if msg.err != nil {
			m.carpetasError = msg.err.Error()
			m.carpetasMensaje = ""
		} else {
			m.carpetasMensaje = msg.mensaje
			m.carpetasError = ""
		}
		return m, leerCarpetasCmd()

	}

	return m, nil
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?":
		m.vista = vistaAyuda
		return m, nil

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
		case 2:
			m.vista = vistaHistorial
			m.historialCargando = true
			return m, leerHistorialCmd()
		case 3:
			m.vista = vistaCarpetas
			m.carpetasModo = modoListaCarpetas
			return m, leerCarpetasCmd()
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
	var cmd tea.Cmd
	m.logsViewport, cmd = m.logsViewport.Update(msg)
	return m, cmd
}

func (m model) updateHistorial(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		m.vista = vistaMenu
		return m, nil
	case "r":
		m.historialCargando = true
		return m, leerHistorialCmd()
	}
	var cmd tea.Cmd
	m.historialViewport, cmd = m.historialViewport.Update(msg)
	return m, cmd
}

// esperarYLimpiarMsg limpia el mensaje de acción tras unos segundos, para
// que no quede pegado indefinidamente en pantalla.
type limpiarMensajeMsg struct{}

func esperarYLimpiarMsg() tea.Cmd {
	return tea.Tick(4*time.Second, func(t time.Time) tea.Msg {
		return limpiarMensajeMsg{}
	})
}

// ----------update de carpetas ------------------
func (m model) updateCarpetas(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.carpetasModo {

	case modoListaCarpetas:
		return m.updateCarpetasLista(msg)

	case modoAgregarCarpeta:
		return m.updateCarpetasAgregar(msg)

	case modoConfirmarEliminar:
		return m.updateCarpetasConfirmar(msg)
	}
	return m, nil
}

func (m model) updateCarpetasLista(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		m.vista = vistaMenu
		m.carpetasMensaje = ""
		m.carpetasError = ""
		return m, nil
	case "up", "k":
		if m.cursorCarpetas > 0 {
			m.cursorCarpetas--
		}
	case "down", "j":
		if m.cursorCarpetas < len(m.carpetasOrden)-1 {
			m.cursorCarpetas++
		}
	case "a":
		m.carpetasModo = modoAgregarCarpeta
		m.inputCarpeta.SetValue("")
		m.inputCarpeta.Focus()
		m.carpetasError = ""
		m.carpetasMensaje = ""
		return m, textinput.Blink
	case "d":
		if len(m.carpetasOrden) == 0 {
			return m, nil
		}
		m.carpetaAConfirmar = m.carpetasOrden[m.cursorCarpetas]
		m.carpetasModo = modoConfirmarEliminar
		return m, nil
	}
	return m, nil
}

func (m model) updateCarpetasAgregar(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.carpetasModo = modoListaCarpetas
		m.inputCarpeta.Blur()
		m.inputCarpetaAviso = ""
		return m, nil
	case "enter":
		nombre := strings.TrimSpace(m.inputCarpeta.Value())
		if err := config.ValidarNombreCarpeta(nombre); err != nil {
			m.carpetasError = err.Error()
			return m, nil
		}
		m.inputCarpeta.Blur()
		m.carpetasModo = modoListaCarpetas
		return m, agregarCarpetaCmd(nombre)
	}

	// Detecta si esta tecla sería rechazada por el validador del input,
	// para mostrar un aviso claro en vez de que simplemente "no pase nada".
	if msg.Type == tea.KeyRunes {
		candidato := m.inputCarpeta.Value() + string(msg.Runes)
		if err := validarCaracterCarpeta(candidato); err != nil {
			m.inputCarpetaAviso = "carácter no permitido"
			var cmd tea.Cmd
			m.inputCarpeta, cmd = m.inputCarpeta.Update(msg)
			return m, tea.Batch(cmd, limpiarAvisoCarpetaCmd())
		}
	}
	m.inputCarpetaAviso = ""

	var cmd tea.Cmd
	m.inputCarpeta, cmd = m.inputCarpeta.Update(msg)
	return m, cmd
}

func (m model) updateCarpetasConfirmar(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "y", "s":
		nombre := m.carpetaAConfirmar
		m.carpetasModo = modoListaCarpetas
		return m, eliminarCarpetaCmd(nombre)
	case "n", "esc":
		m.carpetasModo = modoListaCarpetas
		m.carpetaAConfirmar = ""
		return m, nil
	}
	return m, nil
}

// -----------------Update Help---------------------------
func (m model) updateAyuda(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q", "?":
		m.vista = vistaMenu
		return m, nil
	}
	return m, nil
}

// ---------- View ----------

func (m model) View() string {
	switch m.vista {
	case vistaEstado:
		return m.viewEstado()
	case vistaLogs:
		return m.viewLogs()
	case vistaHistorial:
		return m.viewHistorial()
	case vistaCarpetas:
		return m.viewCarpetas()
	case vistaAyuda:
		return m.viewAyuda()
	default:
		return m.viewMenu()
	}
}

// -------------View  del menu----------------------
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
	contenido += estiloAyuda.Render("↑/↓ navegar · Enter seleccionar · ? ayuda · q salir")
	return lipgloss.Place(
		m.ancho, m.alto,
		lipgloss.Center, lipgloss.Center,
		contenido,
	)
}

// --------------View del estado del servicio ---------------------------
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

// --------------View de los logs ------------------------
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
		s += m.logsViewport.View() + "\n"
	}

	s += estiloAyuda.Render("↑/↓ o j/k desplazar · Actualizando cada 3s · esc volver al menú")
	return s
}

// --------------View del historial de los commits ----------
func (m model) viewHistorial() string {
	s := estiloTitulo.Render("Historial de Respaldos") + "\n\n"

	if m.historialCargando {
		s += "Cargando historial...\n"
		return s
	}

	if m.historialError != "" {
		s += estiloError.Render("Error al leer el historial: "+m.historialError) + "\n"
		s += estiloAyuda.Render("\nesc volver al menú")
		return s
	}

	if len(m.historialCommits) == 0 {
		s += "Aún no hay respaldos registrados.\n"
	} else {
		s += m.historialViewport.View() + "\n"
	}

	s += estiloAyuda.Render("↑/↓ o j/k desplazar · r refrescar · esc volver al menú")
	return s
}

// ---------------------View y utils de las carpetas ----------------------
func ordenarClaves(m map[string]string) []string {
	claves := make([]string, 0, len(m))
	for k := range m {
		claves = append(claves, k)
	}
	sort.Strings(claves)
	return claves
}

func (m model) viewCarpetas() string {
	switch m.carpetasModo {
	case modoAgregarCarpeta:
		return m.viewCarpetasAgregar()
	case modoConfirmarEliminar:
		return m.viewCarpetasConfirmar()
	default:
		return m.viewCarpetasLista()
	}
}

func (m model) viewCarpetasLista() string {
	s := estiloTitulo.Render("Carpetas Vigiladas") + "\n\n"

	if m.carpetasError != "" {
		s += estiloError.Render("Error: "+m.carpetasError) + "\n\n"
	}
	if m.carpetasMensaje != "" {
		s += estiloActivo.Render(m.carpetasMensaje) + "\n\n"
	}

	if len(m.carpetasOrden) == 0 {
		s += "No hay carpetas vigiladas.\n"
	} else {
		for i, nombre := range m.carpetasOrden {
			destino := m.carpetas[nombre]
			linea := nombre + "  →  " + destino
			if i == m.cursorCarpetas {
				s += estiloSeleccionado.Render(linea) + "\n"
			} else {
				s += estiloNormal.Render(linea) + "\n"
			}
		}
	}

	s += estiloAyuda.Render("\na agregar · d eliminar seleccionada · esc volver al menú")
	return s
}

func (m model) viewCarpetasAgregar() string {
	s := estiloTitulo.Render("Agregar Carpeta Vigilada") + "\n\n"
	s += "Nombre de la subcarpeta a vigilar (dentro de watch_dir):\n\n"
	s += m.inputCarpeta.View() + "\n"

	if m.inputCarpetaAviso != "" {
		s += estiloError.Render(m.inputCarpetaAviso) + "\n"
	}
	s += "\n"

	if m.carpetasError != "" {
		s += estiloError.Render(m.carpetasError) + "\n\n"
	}

	s += estiloAyuda.Render("Solo minúsculas, números, '-' y '_' · Enter confirmar · esc cancelar")
	return s
}

func (m model) viewCarpetasConfirmar() string {
	s := estiloTitulo.Render("Confirmar Eliminación") + "\n\n"
	s += "¿Dejar de vigilar la carpeta '" + m.carpetaAConfirmar + "'?\n"
	s += estiloAyuda.Render("(no se borran los archivos físicos, solo se deja de respaldar automáticamente)\n\n")
	s += estiloAyuda.Render("y/s confirmar · n/esc cancelar")
	return s
}

// validarCaracterCarpeta se ejecuta en cada tecla dentro del input.
// Es más permisiva que config.ValidarNombreCarpeta porque debe aceptar
// estados intermedios de escritura; solo bloquea caracteres que NUNCA
// serán válidos (espacios, mayúsculas, barras, símbolos raros).
func validarCaracterCarpeta(valor string) error {
	if len(valor) > 50 {
		return fmt.Errorf("máximo 50 caracteres")
	}
	for _, r := range valor {
		esValido := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_'
		if !esValido {
			return fmt.Errorf("carácter no permitido: %q", r)
		}
	}
	return nil
}

// ----------View de ayuda---------------
func (m model) viewAyuda() string {
	s := estiloTitulo.Render("Ayuda - Atajos de Teclado") + "\n\n"

	secciones := []struct {
		titulo string
		atajos [][2]string
	}{
		{
			"Menú principal",
			[][2]string{
				{"↑/↓ o k/j", "navegar"},
				{"Enter", "seleccionar"},
				{"?", "ver esta ayuda"},
				{"q", "salir"},
			},
		},
		{
			"Estado del servicio",
			[][2]string{
				{"i", "iniciar servicio"},
				{"d", "detener servicio"},
				{"r", "reiniciar servicio"},
				{"esc", "volver al menú"},
			},
		},
		{
			"Logs / Historial",
			[][2]string{
				{"↑/↓ o j/k", "desplazar"},
				{"PgUp/PgDn", "desplazar por página"},
				{"r", "refrescar (solo historial)"},
				{"esc", "volver al menú"},
			},
		},
		{
			"Gestionar carpetas",
			[][2]string{
				{"a", "agregar carpeta"},
				{"d", "eliminar carpeta seleccionada"},
				{"y/n", "confirmar o cancelar eliminación"},
				{"esc", "volver / cancelar"},
			},
		},
	}

	for _, sec := range secciones {
		s += estiloEtiquetaMenu.Render(sec.titulo) + "\n"
		for _, atajo := range sec.atajos {
			s += "  " + estiloFechaCommit.Render(atajo[0]) + "  " + atajo[1] + "\n"
		}
		s += "\n"
	}

	s += estiloAyuda.Render("esc o ? volver al menú")
	return s
}
