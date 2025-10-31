//go:build wasm
// +build wasm

package tinypdf

import (
	"encoding/base64"
	"fmt"
	"syscall/js"
)

// initIO inicializa las funciones de IO para entorno frontend (wasm)
func (tp *TinyPDF) initIO() {
	// Inicializar logger para frontend usando console.log
	tp.logger = func(message ...any) {
		console := js.Global().Get("console")
		if !console.IsUndefined() {
			console.Call("log", fmt.Sprint(message...))
		}
	}
}

// writeFile escribe un archivo en localStorage
func (tp *TinyPDF) writeFile(filePath string, content []byte) error {
	localStorage := js.Global().Get("localStorage")
	if localStorage.IsUndefined() {
		return fmt.Errorf("localStorage no disponible")
	}

	// Codificar el contenido en base64 para almacenarlo
	encoded := base64.StdEncoding.EncodeToString(content)
	localStorage.Call("setItem", filePath, encoded)

	return nil
}

// readFile lee un archivo de localStorage
func (tp *TinyPDF) readFile(filePath string) ([]byte, error) {
	localStorage := js.Global().Get("localStorage")
	if localStorage.IsUndefined() {
		return nil, fmt.Errorf("localStorage no disponible")
	}

	encoded := localStorage.Call("getItem", filePath)
	if encoded.IsNull() {
		return nil, fmt.Errorf("archivo no encontrado: %s", filePath)
	}

	// Decodificar de base64
	decoded, err := base64.StdEncoding.DecodeString(encoded.String())
	if err != nil {
		return nil, fmt.Errorf("error decodificando archivo: %v", err)
	}

	return decoded, nil
}

// fileSize obtiene el tamaño de un archivo de localStorage
func (tp *TinyPDF) fileSize(filePath string) (int64, error) {
	content, err := tp.readFile(filePath)
	if err != nil {
		return 0, err
	}
	return int64(len(content)), nil
}
