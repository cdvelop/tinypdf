package tinypdf_test

import (
	"testing"

	"github.com/cdvelop/tinypdf"
)

func TestTable(t *testing.T) {
	// Create a new PDF document
	pdf := &tinypdf.GoPdf{}
	// Start the PDF with a custom page size (we'll adjust it later)
	pdf.Start(tinypdf.Config{PageSize: tinypdf.Rect{W: 430, H: 200}})
	// Add a new page to the document
	pdf.AddPage()

	err := pdf.AddTTFFont("LiberationSerif-Regular", "./test/res/LiberationSerif-Regular.ttf")
	if err != nil {
		t.Fatalf("Error loading font: %v", err)
		return
	}

	err = pdf.SetFont("LiberationSerif-Regular", "", 11)
	if err != nil {
		t.Fatalf("Error set font: %v", err)
		return
	}
	err = pdf.AddTTFFont("Ubuntu-L.ttf", "./examples/outline_example/Ubuntu-L.ttf")
	if err != nil {
		t.Fatalf("Error loading font: %v", err)
		return
	}

	err = pdf.SetFont("Ubuntu-L.ttf", "", 11)
	if err != nil {
		t.Fatalf("Error set font: %v", err)
		return
	}

	// Set the starting Y position for the table
	tableStartY := 10.0
	// Set the left margin for the table
	marginLeft := 10.0

	// Create a new table layout
	table := pdf.NewTableLayout(marginLeft, tableStartY, 25, 5)

	// Add columns to the table
	table.AddColumn("CODE", 50, "left")
	table.AddColumn("DESCRIPTION", 200, "left")
	table.AddColumn("QTY.", 40, "right")
	table.AddColumn("PRICE", 60, "right")
	table.AddColumn("TOTAL", 60, "right")

	// Add rows to the table
	table.AddRow([]string{"001", "Product A", "2", "10.00", "20.00"})
	table.AddRow([]string{"002", "Product B", "1", "15.00", "15.00"})
	table.AddRow([]string{"003", "Product C", "3", "5.00", "15.00"})

	// Set the style for table cells
	table.SetTableStyle(tinypdf.CellStyle{
		BorderStyle: tinypdf.BorderStyle{
			Top:    true,
			Left:   true,
			Bottom: true,
			Right:  true,
			Width:  1.0,
		},
		FillColor: tinypdf.RGBColor{R: 255, G: 255, B: 255},
		TextColor: tinypdf.RGBColor{R: 0, G: 0, B: 0},
		FontSize:  10,
	})

	// Set the style for table header
	table.SetHeaderStyle(tinypdf.CellStyle{
		BorderStyle: tinypdf.BorderStyle{
			Top:      true,
			Left:     true,
			Bottom:   true,
			Right:    true,
			Width:    2.0,
			RGBColor: tinypdf.RGBColor{R: 100, G: 150, B: 255},
		},
		FillColor: tinypdf.RGBColor{R: 255, G: 200, B: 200},
		TextColor: tinypdf.RGBColor{R: 255, G: 100, B: 100},
		Font:      "Ubuntu-L.ttf",
		FontSize:  12,
	})

	table.SetCellStyle(tinypdf.CellStyle{
		BorderStyle: tinypdf.BorderStyle{
			Right:    true,
			Bottom:   true,
			Width:    0.5,
			RGBColor: tinypdf.RGBColor{R: 0, G: 0, B: 0},
		},
		FillColor: tinypdf.RGBColor{R: 255, G: 255, B: 255},
		TextColor: tinypdf.RGBColor{R: 0, G: 0, B: 0},
		Font:      "LiberationSerif-Regular",
		FontSize:  10,
	})

	// Draw the table
	err = table.DrawTable()
	if err != nil {
		t.Errorf("Error drawing table: %v", err)
	}

	// Save the PDF to the specified path
	err = pdf.WritePdf("test/out/table_custom.pdf")
	if err != nil {
		t.Errorf("Error saving PDF: %v", err)
	}
}

func TestTableCenter(t *testing.T) {
	// Create a new PDF document
	pdf := &tinypdf.GoPdf{}
	// Start the PDF with a custom page size (we'll adjust it later)
	pdf.Start(tinypdf.Config{PageSize: tinypdf.Rect{W: 430, H: 200}})
	// Add a new page to the document
	pdf.AddPage()

	err := pdf.AddTTFFont("LiberationSerif-Regular", "./test/res/LiberationSerif-Regular.ttf")
	if err != nil {
		t.Fatalf("Error loading font: %v", err)
		return
	}

	err = pdf.SetFont("LiberationSerif-Regular", "", 11)
	if err != nil {
		t.Fatalf("Error set font: %v", err)
		return
	}
	err = pdf.AddTTFFont("Ubuntu-L.ttf", "./examples/outline_example/Ubuntu-L.ttf")
	if err != nil {
		t.Fatalf("Error loading font: %v", err)
		return
	}

	err = pdf.SetFont("Ubuntu-L.ttf", "", 11)
	if err != nil {
		t.Fatalf("Error set font: %v", err)
		return
	}

	// Set the starting Y position for the table
	tableStartY := 10.0
	// Set the left margin for the table
	marginLeft := 10.0

	// Create a new table layout
	table := pdf.NewTableLayout(marginLeft, tableStartY, 25, 5)

	// Add columns to the table
	table.AddColumn("CODE", 50, "center")
	table.AddColumn("DESCRIPTION", 200, "center")
	table.AddColumn("QTY.", 40, "center")
	table.AddColumn("PRICE", 60, "center")
	table.AddColumn("TOTAL", 60, "center")

	// Add rows to the table
	table.AddRow([]string{"001", "Product A", "2", "10.00", "20.00"})
	table.AddRow([]string{"002", "Product B", "1", "15.00", "15.00"})
	table.AddRow([]string{"003", "Product C", "3", "5.00", "15.00"})

	// Set the style for table cells
	table.SetTableStyle(tinypdf.CellStyle{
		BorderStyle: tinypdf.BorderStyle{
			Top:    true,
			Left:   true,
			Bottom: true,
			Right:  true,
			Width:  1.0,
		},
		FillColor: tinypdf.RGBColor{R: 255, G: 255, B: 255},
		TextColor: tinypdf.RGBColor{R: 0, G: 0, B: 0},
		FontSize:  10,
	})

	// Set the style for table header
	table.SetHeaderStyle(tinypdf.CellStyle{
		BorderStyle: tinypdf.BorderStyle{
			Top:      true,
			Left:     true,
			Bottom:   true,
			Right:    true,
			Width:    2.0,
			RGBColor: tinypdf.RGBColor{R: 100, G: 150, B: 255},
		},
		FillColor: tinypdf.RGBColor{R: 255, G: 200, B: 200},
		TextColor: tinypdf.RGBColor{R: 255, G: 100, B: 100},
		Font:      "Ubuntu-L.ttf",
		FontSize:  12,
	})

	table.SetCellStyle(tinypdf.CellStyle{
		BorderStyle: tinypdf.BorderStyle{
			Right:    true,
			Bottom:   true,
			Width:    0.5,
			RGBColor: tinypdf.RGBColor{R: 0, G: 0, B: 0},
		},
		FillColor: tinypdf.RGBColor{R: 255, G: 255, B: 255},
		TextColor: tinypdf.RGBColor{R: 0, G: 0, B: 0},
		Font:      "LiberationSerif-Regular",
		FontSize:  10,
	})

	// Draw the table
	err = table.DrawTable()
	if err != nil {
		t.Errorf("Error drawing table: %v", err)
	}

	// Save the PDF to the specified path
	err = pdf.WritePdf("test/out/table_center.pdf")
	if err != nil {
		t.Errorf("Error saving PDF: %v", err)
	}
}

func TestTableWithStyledRows(t *testing.T) {
	// Create a new PDF document
	pdf := &tinypdf.GoPdf{}
	// Start the PDF with a custom page size (we'll adjust it later)
	pdf.Start(tinypdf.Config{PageSize: tinypdf.Rect{W: 430, H: 200}})
	// Add a new page to the document
	pdf.AddPage()

	err := pdf.AddTTFFont("LiberationSerif-Regular", "./test/res/LiberationSerif-Regular.ttf")
	if err != nil {
		t.Fatalf("Error loading font: %v", err)
		return
	}

	err = pdf.SetFont("LiberationSerif-Regular", "", 11)
	if err != nil {
		t.Fatalf("Error set font: %v", err)
		return
	}
	err = pdf.AddTTFFont("Ubuntu-L.ttf", "./examples/outline_example/Ubuntu-L.ttf")
	if err != nil {
		t.Fatalf("Error loading font: %v", err)
		return
	}

	err = pdf.SetFont("Ubuntu-L.ttf", "", 11)
	if err != nil {
		t.Fatalf("Error set font: %v", err)
		return
	}

	// Set the starting Y position for the table
	tableStartY := 10.0
	// Set the left margin for the table
	marginLeft := 10.0

	// Create a new table layout
	table := pdf.NewTableLayout(marginLeft, tableStartY, 25, 5)

	// Add columns to the table
	table.AddColumn("CODE", 50, "left")
	table.AddColumn("DESCRIPTION", 200, "left")
	table.AddColumn("QTY.", 40, "right")
	table.AddColumn("PRICE", 60, "right")
	table.AddColumn("TOTAL", 60, "right")

	// Add rows to the table
	table.AddStyledRow([]tinypdf.RowCell{
		tinypdf.NewRowCell("001", tinypdf.CellStyle{
			TextColor: tinypdf.RGBColor{R: 255, G: 0, B: 0},
		}),
		tinypdf.NewRowCell("Product A", tinypdf.CellStyle{
			TextColor: tinypdf.RGBColor{R: 255, G: 0, B: 0},
		}),
		tinypdf.NewRowCell("2", tinypdf.CellStyle{
			TextColor: tinypdf.RGBColor{R: 255, G: 0, B: 0},
		}),
		tinypdf.NewRowCell("10.00", tinypdf.CellStyle{
			TextColor: tinypdf.RGBColor{R: 255, G: 0, B: 0},
		}),
		tinypdf.NewRowCell("20.00", tinypdf.CellStyle{
			TextColor: tinypdf.RGBColor{R: 255, G: 0, B: 0},
		}),
	})
	table.AddStyledRow([]tinypdf.RowCell{
		tinypdf.NewRowCell("002", tinypdf.CellStyle{
			TextColor: tinypdf.RGBColor{R: 0, G: 255, B: 0},
		}),
		tinypdf.NewRowCell("Product B", tinypdf.CellStyle{
			TextColor: tinypdf.RGBColor{R: 0, G: 255, B: 0},
		}),
		tinypdf.NewRowCell("1", tinypdf.CellStyle{
			TextColor: tinypdf.RGBColor{R: 0, G: 255, B: 0},
		}),
		tinypdf.NewRowCell("15.00", tinypdf.CellStyle{
			TextColor: tinypdf.RGBColor{R: 0, G: 255, B: 0},
		}),
		tinypdf.NewRowCell("15.00", tinypdf.CellStyle{
			TextColor: tinypdf.RGBColor{R: 0, G: 255, B: 0},
		}),
	})
	table.AddStyledRow([]tinypdf.RowCell{
		tinypdf.NewRowCell("003", tinypdf.CellStyle{
			TextColor: tinypdf.RGBColor{R: 255, G: 0, B: 0},
		}),
		tinypdf.NewRowCell("Product C", tinypdf.CellStyle{
			TextColor: tinypdf.RGBColor{R: 0, G: 255, B: 0},
		}),
		tinypdf.NewRowCell("3", tinypdf.CellStyle{
			TextColor: tinypdf.RGBColor{R: 0, G: 0, B: 255},
		}),
		tinypdf.NewRowCell("5.00", tinypdf.CellStyle{
			TextColor: tinypdf.RGBColor{R: 255, G: 0, B: 0},
		}),
		tinypdf.NewRowCell("15.00", tinypdf.CellStyle{
			TextColor: tinypdf.RGBColor{R: 0, G: 255, B: 0},
		}),
	})

	table.AddRow([]string{"004", "Product D", "7", "51.00", "1.00"})

	// Set the style for table cells
	table.SetTableStyle(tinypdf.CellStyle{
		BorderStyle: tinypdf.BorderStyle{
			Top:    true,
			Left:   true,
			Bottom: true,
			Right:  true,
			Width:  1.0,
		},
		FillColor: tinypdf.RGBColor{R: 255, G: 255, B: 255},
		TextColor: tinypdf.RGBColor{R: 0, G: 0, B: 0},
		FontSize:  10,
	})

	// Set the style for table header
	table.SetHeaderStyle(tinypdf.CellStyle{
		BorderStyle: tinypdf.BorderStyle{
			Top:      true,
			Left:     true,
			Bottom:   true,
			Right:    true,
			Width:    2.0,
			RGBColor: tinypdf.RGBColor{R: 100, G: 150, B: 255},
		},
		FillColor: tinypdf.RGBColor{R: 255, G: 200, B: 200},
		TextColor: tinypdf.RGBColor{R: 255, G: 100, B: 100},
		Font:      "Ubuntu-L.ttf",
		FontSize:  12,
	})

	table.SetCellStyle(tinypdf.CellStyle{
		BorderStyle: tinypdf.BorderStyle{
			Right:    true,
			Bottom:   true,
			Width:    0.5,
			RGBColor: tinypdf.RGBColor{R: 0, G: 0, B: 0},
		},
		FillColor: tinypdf.RGBColor{R: 255, G: 255, B: 255},
		TextColor: tinypdf.RGBColor{R: 0, G: 0, B: 0},
		Font:      "LiberationSerif-Regular",
		FontSize:  10,
	})

	// Draw the table
	err = table.DrawTable()
	if err != nil {
		t.Errorf("Error drawing table: %v", err)
	}

	// Save the PDF to the specified path
	err = pdf.WritePdf("test/out/table_with_styled_row.pdf")
	if err != nil {
		t.Errorf("Error saving PDF: %v", err)
	}
}
