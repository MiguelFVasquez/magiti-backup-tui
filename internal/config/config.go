// Package config para las configuraciones del proyecto
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ConfigDaemon refleja la estructura de scripts/config.json del proyecto
// magiti-config-backup (el daemon de respaldo).
type ConfigDaemon struct {
	WatchDir string            `json:"watch_dir"`
	RepoDir  string            `json:"repo_dir"`
	Carpetas map[string]string `json:"carpetas"`
}

// RutaConfigDaemon determina dónde buscar el config.json del daemon.
// Prioriza la variable de entorno MAGITI_DAEMON_CONFIG si está definida,
// y si no, asume la ubicación por defecto usada durante el desarrollo.
func RutaConfigDaemon() string {
	if ruta := os.Getenv("MAGITI_DAEMON_CONFIG"); ruta != "" {
		return ruta
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "proyectos", "magiti-config-backup", "scripts", "config.json")
}

// LeerConfigDaemon carga y parsea el config.json del daemon.
func LeerConfigDaemon() (*ConfigDaemon, error) {
	ruta := RutaConfigDaemon()
	datos, err := os.ReadFile(ruta)
	if err != nil {
		return nil, err
	}

	var cfg ConfigDaemon
	if err := json.Unmarshal(datos, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// RutaLogDaemon devuelve la ruta absoluta del archivo de log del daemon,
// derivada de repo_dir.
func RutaLogDaemon() (string, error) {
	cfg, err := LeerConfigDaemon()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfg.RepoDir, "scripts", "backup-daemon.log"), nil
}

func RepoDirDaemon() (string, error) {
	cfg, err := LeerConfigDaemon()
	if err != nil {
		return "", err
	}
	return cfg.RepoDir, nil
}

// GuardarConfigDaemon escribe la configuración completa de vuelta al
// config.json del daemon, preservando el formato indentado.
func GuardarConfigDaemon(cfg *ConfigDaemon) error {
	ruta := RutaConfigDaemon()
	datos, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ruta, datos, 0o644)
}

// AgregarCarpeta agrega una nueva carpeta vigilada (mapeo 1:1: la carpeta
// dentro de watch_dir y la carpeta destino dentro del repo tienen el mismo
// nombre). También crea físicamente la carpeta dentro de watch_dir si no
// existe, para que inotifywait tenga algo que vigilar.
func AgregarCarpeta(nombre string) error {
	if err := ValidarNombreCarpeta(nombre); err != nil {
		return err
	}

	cfg, err := LeerConfigDaemon()
	if err != nil {
		return err
	}

	if cfg.Carpetas == nil {
		cfg.Carpetas = map[string]string{}
	}
	if _, existe := cfg.Carpetas[nombre]; existe {
		return fmt.Errorf("la carpeta '%s' ya está siendo vigilada", nombre)
	}
	cfg.Carpetas[nombre] = nombre

	rutaFisica := filepath.Join(cfg.WatchDir, nombre)
	if err := os.MkdirAll(rutaFisica, 0o755); err != nil {
		return fmt.Errorf("no se pudo crear el directorio %s: %w", rutaFisica, err)
	}

	rutaDestino := filepath.Join(cfg.RepoDir, nombre)
	if err := os.MkdirAll(rutaDestino, 0o755); err != nil {
		return fmt.Errorf("no se pudo crear el directorio destino %s: %w", rutaDestino, err)
	}

	return GuardarConfigDaemon(cfg)
}

// EliminarCarpeta quita una carpeta del mapeo vigilado. No borra los
// archivos físicos (ni en watch_dir ni en el repo), solo deja de vigilarla.
func EliminarCarpeta(nombre string) error {
	cfg, err := LeerConfigDaemon()
	if err != nil {
		return err
	}

	if _, existe := cfg.Carpetas[nombre]; !existe {
		return fmt.Errorf("la carpeta '%s' no está siendo vigilada", nombre)
	}
	delete(cfg.Carpetas, nombre)

	return GuardarConfigDaemon(cfg)
}

var patronNombreCarpetaValido = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,49}$`)

// ValidarNombreCarpeta verifica que el nombre propuesto sea seguro para
// usarse como nombre de directorio: solo minúsculas, dígitos, guion y
// guion bajo, sin espacios, sin barras (evita crear subcarpetas o rutas
// fuera de watch_dir) y sin secuencias de "..".
func ValidarNombreCarpeta(nombre string) error {
	if nombre == "" {
		return fmt.Errorf("el nombre no puede estar vacío")
	}
	if strings.Contains(nombre, "..") {
		return fmt.Errorf("el nombre no puede contener '..'")
	}
	if strings.ContainsAny(nombre, "/\\") {
		return fmt.Errorf("el nombre no puede contener '/' ni '\\'")
	}
	if !patronNombreCarpetaValido.MatchString(nombre) {
		return fmt.Errorf("solo se permiten minúsculas, números, '-' y '_' (máx. 50 caracteres, debe empezar con letra o número)")
	}
	return nil
}
