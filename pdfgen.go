package tinypdf

import (
	"fmt"
	"io"

	"github.com/cdvelop/tinypdf/fpdf"
)

// GenerateSamplePDF genera un PDF de ejemplo con el título especificado
// Esta función es independiente del entorno (WASM o backend)
// IMPORTANTE: Crea una nueva instancia de PDF cada vez
func (tp *TinyPDF) GenerateSamplePDF(title string) (*fpdf.Fpdf, error) {
	// Usar la instancia PDF ya configurada y agregar las fuentes requeridas
	pdf := tp.Fpdf

	// Agregar las fuentes TTF requeridas
	pdf.AddFont("Arial", "", "Arial.ttf")
	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("error adding Arial font: %v", err)
	}
	pdf.AddFont("Arial", "B", "Arial_Bold.ttf")
	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("error adding Arial Bold font: %v", err)
	}
	pdf.AddFont("Arial", "I", "Arial_Italic.ttf")
	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("error adding Arial Italic font: %v", err)
	}
	pdf.AddFont("DejaVu", "", "DejaVuSansCondensed.ttf")
	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("error adding DejaVu font: %v", err)
	}

	// Configurar márgenes y header/footer
	pdf.SetTopMargin(30)
	pdf.SetHeaderFuncMode(func() {
		pdf.SetY(5)
		pdf.SetFont("Arial", "B", 15)
		pdf.CellFormat(0, 10, title, "", 0, "C", false, 0, "")
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
	pdf.SetFont("DejaVu", "", 12)
	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("error setting DejaVu font: %v", err)
	}

	// Agregar contenido
	pdf.SetFont("Arial", "B", 14)
	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("error setting Arial Bold font for content: %v", err)
	}
	pdf.CellFormat(0, 10, "Contenido del documento:", "", 1, "", false, 0, "")
	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("error writing content title: %v", err)
	}
	pdf.Ln(5)

	pdf.SetFont("DejaVu", "", 12)
	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("error setting DejaVu font for lines: %v", err)
	}
	for j := 1; j <= 40; j++ {
		pdf.CellFormat(0, 10, fmt.Sprintf("Línea de contenido número %d", j),
			"", 1, "", false, 0, "")
		if err := pdf.Error(); err != nil {
			return nil, fmt.Errorf("error writing line %d: %v", j, err)
		}
	}

	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("error generating PDF: %v", err)
	}

	return pdf, nil
}

// GetPDFBytes genera el PDF y retorna los bytes
// IMPORTANTE: Cierra el PDF internamente
func GetPDFBytes(pdf *fpdf.Fpdf, w io.Writer) error {
	// El PDF se cierra automáticamente en Output si es necesario
	return pdf.Output(w)
}
