//go:build !wasm
// +build !wasm

package tinypdf

import (
	"fmt"
	"os"
)

// initIO inicializa las funciones de IO para entorno backend (no-wasm)
func (tp *TinyPDF) initIO() {
	// Inicializar logger para backend usando fmt.Println
	tp.logger = func(message ...any) {
		fmt.Println(message...)
	}

	// Inicializar fontLoader para backend usando os.ReadFile
	tp.fontLoader = func(fontPath string) ([]byte, error) {
		return os.ReadFile(fontPath)
	}
}

// writeFile escribe un archivo en el sistema de archivos usando os
func (tp *TinyPDF) writeFile(filePath string, content []byte) error {
	return os.WriteFile(filePath, content, 0644)
}

// readFile lee un archivo del sistema de archivos usando os
func (tp *TinyPDF) readFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// fileSize obtiene el tamaño de un archivo usando os.Stat
func (tp *TinyPDF) fileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}
