package pdf_test

import (
	"testing"

	"github.com/tinywasm/pdf"
)

func TestAPI_Basic(t *testing.T) {
	doc := pdf.NewDocument()
	doc.SetPageHeader().SetLeftText("Test Header")
	doc.SetPageFooter().WithPageTotal("R")
	doc.AddPage()
	doc.AddHeader1("Hello World")
	doc.AddText("This is a test document.").Draw()
	doc.SpaceBefore(10)
	doc.AddText("Another paragraph.").Bold().AlignRight().Draw()

	err := doc.WritePdf("test_api.pdf")
	if err != nil {
		t.Errorf("WritePdf failed: %v", err)
	}
}
