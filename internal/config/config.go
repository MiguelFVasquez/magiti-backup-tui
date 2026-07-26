// Package config para las configuraciones del proyecto
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
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
