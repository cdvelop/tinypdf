package fpdf_test

import (
	fpdf "github.com/tinywasm/pdf/fpdf"
	"github.com/tinywasm/pdf/fpdf/contrib/httpimg"
)

func ExampleRegister() {
	pdf := NewDocPdfTest("mm", "A4", fpdf.Landscape)
	pdf.SetFont("Helvetica", "", 12)
	pdf.SetFillColor(200, 200, 220)
	pdf.AddPage()

	url := "https://github.com/tinywasm/pdf/raw/main/fpdf/image/logo_gofpdf.jpg"
	httpimg.Register(pdf, url, "")
	pdf.Image(url, 15, 15, 267, 0, false, "", 0, "")
	fileStr := Filename("contrib_httpimg_Register")
	err := pdf.OutputFileAndClose(fileStr)
	Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/contrib_httpimg_Register.pdf
}
