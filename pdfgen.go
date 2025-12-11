package pdf

import (
	"fmt"
	"io"

	"github.com/tinywasm/pdf/fpdf"
)

// GenerateSamplePDF genera un PDF de ejemplo con el título especificado
// Esta función es independiente del entorno (WASM o backend)
// IMPORTANTE: Crea una nueva instancia de PDF cada vez
func (tp *TinyPDF) GenerateSamplePDF(title string) (*fpdf.Fpdf, error) {
	// Crear una nueva instancia de PDF para evitar problemas de estado
	pdf := fpdf.New(
		fpdf.FontsDirName("fonts"),
		fpdf.WriteFileFunc(tp.writeFile),
		fpdf.ReadFileFunc(tp.readFile),
		fpdf.FileSizeFunc(tp.fileSize),
	)

	// Configurar márgenes y header/footer
	pdf.SetTopMargin(30)
	pdf.SetHeaderFuncMode(func() {
		pdf.SetY(5)
		pdf.SetFont("Arial", "B", 15)
		pdf.Cell(80, 0, "")
		pdf.CellFormat(80, 10, title, "1", 0, "C", false, 0, "")
		pdf.Ln(20)
	}, true)

	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, fmt.Sprintf("Página %d/{nb}", pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})

	pdf.AliasNbPages("")
	pdf.AddPage()
	pdf.SetFont("Times", "", 12)

	// Agregar contenido
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(0, 10, "Contenido del documento:", "", 1, "", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Times", "", 12)
	for j := 1; j <= 40; j++ {
		pdf.CellFormat(0, 10, fmt.Sprintf("Línea de contenido número %d", j),
			"", 1, "", false, 0, "")
	}

	return pdf, nil
}

// GetPDFBytes genera el PDF y retorna los bytes
// IMPORTANTE: Cierra el PDF internamente
func GetPDFBytes(pdf *fpdf.Fpdf, w io.Writer) error {
	// El PDF se cierra automáticamente en Output si es necesario
	return pdf.Output(w)
}
