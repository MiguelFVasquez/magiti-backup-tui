package daemon

import (
	"bufio"
	"os"
)

// LeerUltimasLineas lee un archivo y devuelve como máximo las últimas n
// líneas, en orden cronológico (la más antigua primero).
func LeerUltimasLineas(ruta string, n int) ([]string, error) {
	archivo, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer archivo.Close()

	var todas []string
	scanner := bufio.NewScanner(archivo)
	// Buffer más grande por si alguna línea de log es larga.
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		todas = append(todas, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if len(todas) <= n {
		return todas, nil
	}
	return todas[len(todas)-n:], nil
}
