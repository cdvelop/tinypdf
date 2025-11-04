//go:build wasm
// +build wasm

package tinypdf

import (
	"encoding/base64"
	"fmt"
	"syscall/js"

	"github.com/cdvelop/fetchgo"
	. "github.com/cdvelop/tinystring"
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

	// Inicializar fontLoader para frontend usando fetchgo
	tp.fontLoader = tp.loadFontFromURL
}

// loadFontFromURL loads TTF fonts from current domain using fetchgo
func (tp *TinyPDF) loadFontFromURL(fontPath string) ([]byte, error) {
	location := js.Global().Get("location")
	if location.IsUndefined() {
		return nil, fmt.Errorf("window.location not available")
	}

	origin := location.Get("origin").String()
	fullURL := origin + "/" + fontPath

	client := &fetchgo.Client{
		RequestType: fetchgo.RequestRaw,
	}

	resultChan := make(chan []byte, 1)
	errorChan := make(chan error, 1)

	client.SendRequest("GET", fullURL, nil, func(result any, err error) {
		if err != nil {
			errorChan <- fmt.Errorf("failed to fetch font %s: %w", fontPath, err)
			return
		}

		if fontData, ok := result.([]byte); ok {
			resultChan <- fontData
		} else {
			errorChan <- fmt.Errorf("unexpected result type from fetchgo: %T", result)
		}
	})

	select {
	case data := <-resultChan:
		return data, nil
	case err := <-errorChan:
		return nil, err
	}
}

// writeFile escribe un archivo en localStorage
func (tp *TinyPDF) writeFile(filePath string, content []byte) error {
	localStorage := js.Global().Get("localStorage")
	if localStorage.IsUndefined() {
		return Errf("localStorage no disponible")
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
		return nil, Errf("localStorage no disponible")
	}

	encoded := localStorage.Call("getItem", filePath)
	if encoded.IsNull() {
		return nil, Errf("archivo no encontrado: %s", filePath)
	}

	// Decodificar de base64
	decoded, err := base64.StdEncoding.DecodeString(encoded.String())
	if err != nil {
		return nil, Errf("error decodificando archivo: %v", err)
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
