//go:build wasm
// +build wasm

package pdf

import (
	"encoding/base64"
	"github.com/tinywasm/fetch"
	. "github.com/tinywasm/fmt"
	"io"
	"syscall/js"
)

// initIO inicializa las funciones de IO para entorno frontend (wasm)
func (tp *TinyPDF) initIO() {
	// Inicializar logger para frontend usando console.log
	tp.logger = func(message ...any) {
		console := js.Global().Get("console")
		if !console.IsUndefined() {
			console.Call("log", Translate(message...))
		}
	}
}

// writeFile escribe un archivo en localStorage y dispara una descarga
func (tp *TinyPDF) writeFile(filePath string, content []byte) error {
	// 1. Guardar en localStorage (como backup/persistencia simple)
	localStorage := js.Global().Get("localStorage")
	if !localStorage.IsUndefined() {
		encoded := base64.StdEncoding.EncodeToString(content)
		localStorage.Call("setItem", filePath, encoded)
	}

	// 2. Disparar descarga del navegador
	uint8Array := js.Global().Get("Uint8Array").New(len(content))
	js.CopyBytesToJS(uint8Array, content)

	blob := js.Global().Get("Blob").New([]any{uint8Array}, map[string]any{"type": "application/pdf"})
	url := js.Global().Get("URL").Call("createObjectURL", blob)

	link := js.Global().Get("document").Call("createElement", "a")
	link.Set("href", url)
	link.Set("download", filePath)
	link.Call("click")

	return nil
}

// readFile lee un archivo usando fetch (para cargar recursos estáticos como fuentes e imágenes)
func (tp *TinyPDF) readFile(filePath string) ([]byte, error) {
	resp, err := fetch.Get(filePath)
	if err != nil {
		return nil, Errf("error fetching file %s: %v", filePath, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, Errf("error fetching file %s: status %d", filePath, resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, Errf("error reading response body for %s: %v", filePath, err)
	}

	return data, nil
}

// fileSize obtiene el tamaño de un archivo de localStorage
func (tp *TinyPDF) fileSize(filePath string) (int64, error) {
	content, err := tp.readFile(filePath)
	if err != nil {
		return 0, err
	}
	return int64(len(content)), nil
}
