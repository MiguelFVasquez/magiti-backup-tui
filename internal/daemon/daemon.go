package daemon

import (
	"os/exec"
	"strings"
)

const servicio = "magiti-backup.service"

// Estado representa la situación actual del servicio.
type Estado struct {
	Activo   bool
	Detalle  string
	ErrorMsg string
}

// ConsultarEstado ejecuta `systemctl is-active` y `systemctl status` para
// obtener tanto un booleano simple como el detalle textual completo.
func ConsultarEstado() Estado {
	activoCmd := exec.Command("sudo", "-n", "systemctl", "is-active", servicio)
	salidaActivo, _ := activoCmd.Output()
	activo := strings.TrimSpace(string(salidaActivo)) == "active"

	detalleCmd := exec.Command("sudo", "-n", "systemctl", "status", servicio, "--no-pager", "-l")
	salidaDetalle, err := detalleCmd.CombinedOutput()

	e := Estado{
		Activo:  activo,
		Detalle: string(salidaDetalle),
	}
	if err != nil && !activo {
		// systemctl status devuelve código de error si el servicio está
		// inactivo; no lo tratamos como fallo real, solo lo reflejamos.
		e.ErrorMsg = ""
	}
	return e
}

// Iniciar arranca el servicio.
func Iniciar() error {
	return ejecutarAccion("start")
}

// Detener detiene el servicio.
func Detener() error {
	return ejecutarAccion("stop")
}

// Reiniciar reinicia el servicio (usado luego de cambios de configuración).
func Reiniciar() error {
	return ejecutarAccion("restart")
}

func ejecutarAccion(accion string) error {
	cmd := exec.Command("sudo", "-n", "systemctl", accion, servicio)
	salida, err := cmd.CombinedOutput()
	if err != nil {
		return &ErrorAccion{Accion: accion, Salida: string(salida), Original: err}
	}
	return nil
}

// ErrorAccion envuelve un fallo al ejecutar una acción de systemctl,
// incluyendo la salida cruda para poder mostrarla en la TUI.
type ErrorAccion struct {
	Accion   string
	Salida   string
	Original error
}

func (e *ErrorAccion) Error() string {
	return "fallo al ejecutar '" + e.Accion + "': " + e.Original.Error() + " | " + e.Salida
}
