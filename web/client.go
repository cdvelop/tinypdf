//go:build wasm

package main

import (
	"github.com/tinywasm/pdf"
	"github.com/tinywasm/pdf/web/ui"
)

func main() {
	// Crear instancia de TinyWasmPDF
	tp := pdf.New()

	tp.Log("TinyWasmPDF inicializado...")

	// Configurar UI
	ui.Setup(tp)

	tp.Log("Aplicación lista")

	// Mantener el programa ejecutándose
	select {}
}
