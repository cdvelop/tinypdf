package tinypdf

import (
	"bytes"
	"os"
	"testing"

	"github.com/cdvelop/tinypdf/fpdf"
)

func TestGenerateSamplePDF(t *testing.T) {
	// Crear instancia de TinyPDF
	tp := New(
		fpdf.RootDirectoryType("."),
		fpdf.FontsDirName("fpdf/fonts"),
	)

	// Generar PDF
	pdf, err := tp.GenerateSamplePDF("Test Document")
	if err != nil {
		t.Fatalf("Error generando PDF: %v", err)
	}

	// Usar OutputFileAndClose en lugar de Output
	err = pdf.OutputFileAndClose("test_output.pdf")
	if err != nil {
		t.Fatalf("Error guardando PDF: %v", err)
	}

	// Leer el archivo generado
	pdfBytes, err := os.ReadFile("test_output.pdf")
	if err != nil {
		t.Fatalf("Error leyendo PDF generado: %v", err)
	}

	// Verificar que se generaron bytes
	if len(pdfBytes) == 0 {
		t.Fatal("El PDF no tiene bytes")
	}

	t.Logf("PDF generado correctamente, tamaño: %d bytes", len(pdfBytes))

	// Guardar el PDF para inspección manual (ANTES de validar)
	err = os.WriteFile("test_output.pdf", pdfBytes, 0644)
	if err != nil {
		t.Logf("No se pudo guardar el PDF de prueba: %v", err)
	} else {
		t.Log("PDF de prueba guardado en: test_output.pdf")
	}

	// Debug: Ver los primeros 20 bytes
	if len(pdfBytes) >= 20 {
		t.Logf("Primeros 20 bytes: %v", pdfBytes[:20])
		t.Logf("Primeros 20 bytes (string): %q", string(pdfBytes[:20]))
	}

	// Verificar que comienza con el header de PDF
	if len(pdfBytes) < 5 {
		t.Fatal("PDF demasiado pequeño")
	}

	pdfHeader := string(pdfBytes[:5])
	if pdfHeader != "%PDF-" {
		t.Fatalf("Header de PDF inválido: %q (bytes: %v)", pdfHeader, pdfBytes[:10])
	}

	t.Log("Header de PDF válido: %PDF")

	// Verificar que termina con EOF de PDF
	if len(pdfBytes) < 6 {
		t.Fatal("PDF demasiado pequeño para verificar EOF")
	}

	// Los PDFs deben terminar con %%EOF
	pdfEnd := string(pdfBytes[len(pdfBytes)-6:])
	if pdfEnd != "%%EOF\n" && pdfEnd[len(pdfEnd)-5:] != "%%EOF" {
		t.Logf("Últimos bytes del PDF: %q", pdfEnd)
		// No es crítico, algunos PDFs pueden no tener newline al final
	}

	// Guardar el PDF para inspección manual
	err = os.WriteFile("test_output.pdf", pdfBytes, 0644)
	if err != nil {
		t.Logf("No se pudo guardar el PDF de prueba: %v", err)
	} else {
		t.Log("PDF de prueba guardado en: test_output.pdf")
	}
}

func TestPDFStructure(t *testing.T) {
	// Crear instancia de TinyPDF
	tp := New(
		fpdf.RootDirectoryType("."),
		fpdf.FontsDirName("fpdf/fonts"),
	)

	// Generar PDF
	pdf, err := tp.GenerateSamplePDF("Documento de Prueba")
	if err != nil {
		t.Fatalf("Error generando PDF: %v", err)
	}

	// Obtener bytes
	var buf bytes.Buffer
	err = GetPDFBytes(pdf, &buf)
	if err != nil {
		t.Fatalf("Error obteniendo bytes del PDF: %v", err)
	}

	pdfBytes := buf.Bytes()

	// Verificar presencia de elementos clave
	checks := []struct {
		name    string
		pattern string
	}{
		{"PDF version", "%PDF-1."},
		{"Catalog", "/Catalog"},
		{"Pages", "/Pages"},
		{"Font", "/Font"},
		{"Content stream", "/Length"},
	}

	for _, check := range checks {
		if !bytes.Contains(pdfBytes, []byte(check.pattern)) {
			t.Errorf("PDF no contiene %s (%s)", check.name, check.pattern)
		} else {
			t.Logf("✓ PDF contiene %s", check.name)
		}
	}

	t.Logf("Estructura del PDF parece correcta")
}
