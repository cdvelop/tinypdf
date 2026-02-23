package pdf_test

import (
	"testing"

	"github.com/tinywasm/pdf"
)

func TestCharts(t *testing.T) {
	doc := pdf.NewDocument()
	doc.AddPage()

	doc.AddHeader1("Chart Examples")

	// Bar Chart
	doc.AddHeader2("Bar Chart")
	doc.Chart().Bar().
		Title("Monthly Sales").
		Height(100).
		AddBar(120, "Jan", pdf.ColorRGB(50, 100, 200)).
		AddBar(140, "Feb", pdf.ColorRGB(200, 100, 50)).
		AddBar(110, "Mar", pdf.ColorRGB(50, 200, 100)).
		Draw()

	doc.SpaceBefore(10)

	// Line Chart
	doc.AddHeader2("Line Chart")
	doc.Chart().Line().
		Title("Growth Trends").
		Height(100).
		AddSeries("Revenue", []float64{10, 15, 13, 17, 20, 25, 22}, pdf.ColorRGB(0, 0, 255)).
		Draw()

	doc.SpaceBefore(10)

	// Pie Chart
	doc.AddHeader2("Pie Chart")
	doc.Chart().Pie().
		Title("Market Share").
		Height(120).
		AddSlice("A", 40, pdf.ColorRGB(255, 0, 0)).
		AddSlice("B", 30, pdf.ColorRGB(0, 255, 0)).
		AddSlice("C", 30, pdf.ColorRGB(0, 0, 255)).
		Draw()

	err := doc.WritePdf("test_charts.pdf")
	if err != nil {
		t.Errorf("WritePdf failed: %v", err)
	}
}
