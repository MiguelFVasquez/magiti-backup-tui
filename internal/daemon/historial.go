package daemon

import (
	"os/exec"
	"strconv"
	"strings"
)

// CommitInfo representa una entrada del historial de respaldos.
type CommitInfo struct {
	Hash    string
	Fecha   string
	Mensaje string
}

const separadorCommit = "|||"

// LeerHistorial ejecuta `git log` sobre repoDir y devuelve los últimos n
// commits, del más reciente al más antiguo.
func LeerHistorial(repoDir string, n int) ([]CommitInfo, error) {
	formato := "%h" + separadorCommit + "%ad" + separadorCommit + "%s"
	cmd := exec.Command(
		"git", "-C", repoDir,
		"log",
		"-n", strconv.Itoa(n),
		"--date=format:%Y-%m-%d %H:%M:%S",
		"--pretty=format:"+formato,
	)

	salida, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lineas := strings.Split(strings.TrimSpace(string(salida)), "\n")
	var commits []CommitInfo
	for _, linea := range lineas {
		if linea == "" {
			continue
		}
		partes := strings.SplitN(linea, separadorCommit, 3)
		if len(partes) != 3 {
			continue
		}
		commits = append(commits, CommitInfo{
			Hash:    partes[0],
			Fecha:   partes[1],
			Mensaje: partes[2],
		})
	}
	return commits, nil
}
