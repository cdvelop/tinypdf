//go:build wasm

package main

import (
	"github.com/tinywasm/pdf"
	"github.com/tinywasm/pdf/web/ui"
)

func main() {
	// Crear instancia de Document
	doc := pdf.NewDocument()
	doc.SetLog(func(msg ...any) {
		// En WASM initIO ya configura un logger por defecto que usa console.log
		// Pero si queremos personalizarlo:
		// console := js.Global().Get("console")
		// if !console.IsUndefined() {
		// 	console.Call("log", Translate(msg...))
		// }
	})

	doc.Log("Document inicializado...")

	// Configurar UI
	ui.Setup(doc)

	doc.Log("Aplicación lista")

	// Mantener el programa ejecutándose
	select {}
}
