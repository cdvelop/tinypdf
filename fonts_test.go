package docpdf_test

import "testing"

// Test_SetUnderlineThickness demonstrates how to adjust the text
// underline thickness.
func Test_SetUnderlineThickness(t *testing.T) {
	pdf := NewDocPdfTest() // 210mm x 297mm
	pdf.AddPage()
	pdf.SetFont("Arial", "U", 12)

	pdf.SetUnderlineThickness(0.5)
	pdf.CellFormat(0, 10, "Thin underline", "", 1, "", false, 0, "")

	pdf.SetUnderlineThickness(1)
	pdf.CellFormat(0, 10, "Normal underline", "", 1, "", false, 0, "")

	pdf.SetUnderlineThickness(2)
	pdf.CellFormat(0, 10, "Thicker underline", "", 1, "", false, 0, "")

	fileStr := Filename("Test_UnderlineThickness")
	err := pdf.OutputFileAndClose(fileStr)
	SummaryCompare(err, fileStr)
	// Output:
	// Successfully generated pdf/Test_UnderlineThickness.pdf
}
