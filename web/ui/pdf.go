//go:build wasm
// +build wasm

package ui

import (
	"syscall/js"
)

// GeneratePDF genera el PDF con el título especificado
func GeneratePDF() {
	TP.Log("Iniciando generación de PDF...")

	// Obtener el título del input
	titleText := GetTitleText()

	// Generar PDF usando la función centralizada
	pdf, err := TP.GenerateSamplePDF(titleText)
	if err != nil {
		TP.Log("Error generando PDF:", err.Error())
		ShowError("Error al generar PDF: " + err.Error())
		return
	}

	// Obtener los bytes del PDF
	var buf []byte
	err = pdf.Output(&bytesWriter{data: &buf})
	if err != nil {
		TP.Log("Error obteniendo bytes del PDF:", err.Error())
		ShowError("Error al obtener bytes del PDF: " + err.Error())
		return
	}

	TP.Log("PDF generado, tamaño:", len(buf), "bytes")

	// Usar Blob API para mostrar el PDF (más eficiente que base64)
	ShowPDFFromBytes(buf)

	TP.Log("PDF generado exitosamente")
}

// bytesWriter implementa io.Writer para capturar los bytes del PDF
type bytesWriter struct {
	data *[]byte
}

func (bw *bytesWriter) Write(p []byte) (n int, err error) {
	*bw.data = append(*bw.data, p...)
	return len(p), nil
}

// ShowPDFFromBytes crea un Blob a partir de los bytes del PDF y lo muestra
// usando Blob API en lugar de base64 para mejor rendimiento
func ShowPDFFromBytes(pdfBytes []byte) {
	TP.Log("Creando Blob del PDF, tamaño:", len(pdfBytes), "bytes")

	// Crear Uint8Array desde los bytes de Go
	uint8Array := js.Global().Get("Uint8Array").New(len(pdfBytes))
	js.CopyBytesToJS(uint8Array, pdfBytes)

	// Crear el Blob con el tipo MIME correcto
	// El constructor de Blob espera: new Blob([array], {type: 'mime/type'})
	blobParts := []interface{}{uint8Array}
	blobOptions := map[string]interface{}{
		"type": "application/pdf",
	}

	blob := js.Global().Get("Blob").New(blobParts, blobOptions)

	// Crear URL del Blob usando la función integrada
	blobURL := CreateBlobURL(blob)

	TP.Log("Blob URL creado:", blobURL)

	// Mostrar el PDF usando el blob URL
	ShowPDF(blobURL)
}

// CreateBlobURL crea una URL Blob a partir de un blob.
func CreateBlobURL(blob any) string {
	jsBlob := js.ValueOf(blob)
	return js.Global().Get("URL").Call("createObjectURL", jsBlob).String()
}
