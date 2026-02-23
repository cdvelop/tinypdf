package pdf_test

import (
	"testing"

	"github.com/tinywasm/pdf"
)

func TestTable(t *testing.T) {
	doc := pdf.NewDocument()
	doc.AddPage()

	table := doc.AddTable().
		AddColumn("Code").Width(20).AlignCenter().
		AddColumn("Product").Width(80).AlignLeft().
		AddColumn("Price").Width(30).AlignRight().Prefix("$")

	table.HeaderStyle(pdf.Style{
		FillColor: pdf.ColorRGB(200, 200, 200),
		TextColor: pdf.ColorRGB(0, 0, 0),
		Font:      pdf.FontBold,
		FontSize:  12,
	})

	table.AddRow("001", "Widget A", "10.00")
	table.AddRow("002", "Widget B", "20.50")
	table.AddRow("003", "Widget C", "5.99")

	table.Draw()

	err := doc.WritePdf("test_table.pdf")
	if err != nil {
		t.Errorf("WritePdf failed: %v", err)
	}
}
