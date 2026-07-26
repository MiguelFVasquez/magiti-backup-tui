# MAGITI - Backup TUI

Interfaz de terminal (TUI) para gestionar y monitorear el daemon de respaldo automático de configuraciones críticas del Grupo de Investigación GRID.

Proyecto desarrollado en el marco del trabajo de grado **MAGITI**: *Modelo de Automatización para la Gestión de la Infraestructura de TI del Grupo de Investigación GRID*.

## Contexto y Justificación

Este proyecto complementa a [`magiti-config-backup`](https://github.com/MiguelFVasquez/magiti-config-backup), el daemon encargado de vigilar y respaldar automáticamente las configuraciones de Firewall y VPN hacia un repositorio Git. Si bien el daemon corre de forma autónoma como servicio `systemd`, el administrador de infraestructura necesita una forma rápida y visual de:

- Verificar si el servicio está activo, sin recordar comandos de `systemctl`.
- Revisar el registro de actividad (logs) sin usar `cat`/`tail` manualmente.
- Consultar el historial de respaldos realizados (commits) de forma legible.
- Agregar o quitar carpetas vigiladas sin editar archivos de configuración a mano.

Esta TUI resuelve esa necesidad, siguiendo el mismo enfoque de **interfaz de terminal rápida y basada en teclado** que otras herramientas del proyecto MAGITI (ver [`baky`](https://github.com/Josedzzz/baky), desarrollada en paralelo para el respaldo de Máquinas Virtuales).

## Alcance del Proyecto (v1)

### Dentro del alcance

- Pantalla de estado del servicio (`magiti-backup.service`): activo/inactivo, con acciones para iniciar, detener y reiniciar.
- Visualización de los logs del daemon.
- Visualización del historial de commits (respaldos realizados) del repositorio de configuraciones.
- Gestión de carpetas vigiladas: agregar o quitar entradas en `config.json` del daemon, con reinicio automático del servicio al guardar cambios.

### Fuera del alcance (fases posteriores)

- Edición del contenido de los archivos de configuración respaldados (Firewall/VPN) desde la TUI.
- Comparación de versiones (diffs) entre commits.
- Notificaciones o alertas proactivas.
- Soporte multi-usuario o multi-servidor desde una sola instancia de la TUI.

## Tecnologías Utilizadas

| Tecnología | Propósito |
|---|---|
| **Go** | Lenguaje principal del proyecto |
| **Bubble Tea** ([charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)) | Framework de la interfaz de terminal (TUI) |
| **Bubbles** ([charmbracelet/bubbles](https://github.com/charmbracelet/bubbles)) | Componentes de interfaz reutilizables (listas, spinners) |
| **Lip Gloss** ([charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)) | Estilos y colores en terminal |
| **systemd** (vía `systemctl`) | Control del ciclo de vida del daemon de respaldo |
| **Git** (vía comandos de sistema) | Lectura del historial de respaldos |

## Justificación de Decisiones Técnicas

- **¿Por qué Go y Bubble Tea, en vez de una TUI en Bash (`gum`/`whiptail`)?** Se optó por mantener consistencia con el stack ya utilizado en `baky`, otra herramienta del mismo proyecto MAGITI. Esto facilita el mantenimiento a largo plazo por parte del grupo de investigación y permite, a futuro, reutilizar componentes de interfaz entre ambos proyectos.

- **¿Por qué un repositorio separado del daemon (`magiti-config-backup`)?** El daemon y la TUI tienen ciclos de vida y lenguajes distintos: el daemon es un servicio en segundo plano sin interfaz (Bash), mientras que la TUI es una aplicación interactiva (Go) que el administrador ejecuta bajo demanda. Mantenerlos separados facilita la documentación individual y evita mezclar dependencias de ambos ecosistemas.

- **¿Por qué la TUI no reemplaza al daemon ni corre su lógica internamente?** La TUI actúa como panel de control, no como motor de respaldo. El daemon sigue siendo gestionado por `systemd` de forma independiente; la TUI únicamente consulta su estado y modifica su configuración, delegando el reinicio del servicio al propio sistema operativo.

## Estructura del Proyecto

```
magiti-backup-tui/
├── cmd/
│   └── main.go              Punto de entrada de la aplicación
├── internal/
│   ├── tui/                 Modelo, vistas y estilos de la interfaz
│   ├── daemon/               Integración con systemctl (estado, start/stop/restart)
│   └── config/                Lectura y escritura de config.json del daemon
├── go.mod
├── go.sum
└── README.md
```

## Requisitos

- Go 1.23 o superior.
- El proyecto [`magiti-config-backup`](https://github.com/MiguelFVasquez/magiti-config-backup) instalado y corriendo como servicio `systemd` en la misma máquina.
- Permisos de usuario para ejecutar `systemctl` sobre el servicio `magiti-backup.service`.

## Instalación y Uso

```bash
git clone git@github.com:MiguelFVasquez/magiti-backup-tui.git
cd magiti-backup-tui
go build -o magiti-tui ./cmd
./magiti-tui
```

### Controles

- `↑ / ↓` o `k / j`: navegar por los menús.
- `Enter`: seleccionar una opción.
- `q` o `Ctrl+C`: salir.

## Estado del Proyecto

🚧 En desarrollo — diseño inicial de pantallas en construcción.

## Autores

- Jose D. Amaya — Investigador del Proyecto MAGITI
- Juan M. Florez — Investigador del Proyecto MAGITI

Basado en entrevista con Luis E. Sepúlveda, Administrador de Sistemas GRID (22 de junio de 2026).
