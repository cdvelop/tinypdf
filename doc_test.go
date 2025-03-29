package tinypdf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentAPIUsage(t *testing.T) {
	// Create a simple document with default settings
	doc := NewDocument(func(a ...any) {
		// Simple logger that does nothing for this test
		t.Log(a...)
	})

	// add logo image
	doc.AddImage("test/res/logo.png").Height(35).Inline().Draw()

	// add date and time aligned to the right
	doc.AddText("date: 2023-10-01").AlignRight().Inline().Draw()

	// Add a centered header
	doc.AddHeader1("Example Document").AlignCenter().Draw()

	// Add a level 2 header
	doc.AddHeader2("Section 1: Introduction").Draw()

	// Add normal text
	doc.AddText("This is a normal text paragraph that shows the basic capabilities of the gopdf library. " +
		"We can create documents with different text styles and formats.").Draw()

	// Add text with different styles
	doc.AddText("This text is in bold.").Bold().Draw()

	doc.AddText("This text is in italic.").Italic().Draw()

	// Add right-aligned text (ensuring it's in regular style, not italic)
	doc.AddText("This text is right-aligned.").Regular().AlignRight().Draw()

	// Add justified text with multiple lines
	doc.AddHeader3("Example of justified text").Draw()
	doc.AddText("This is an example of justified text that extends across multiple lines. " +
		"Justification distributes text uniformly between the left and right margins, " +
		"which provides a more professional and organized look to the document. " +
		"This is especially useful for formal documents such as reports, contracts, or books. " +
		"As you can see, the spaces between words are automatically adjusted so that each line " +
		"(except the last one in each paragraph) occupies the full available width. " +
		"The last line is left-aligned by typographic convention.").Justify().Draw()

	// bar chart image
	doc.AddImage("test/res/barchart.png").Height(150).AlignCenter().Draw()
	// Add a footnote (in italic by default)
	doc.AddFootnote("This is a footnote.").AlignCenter().Draw()

	// add gopher image as a right-aligned inline image
	doc.AddImage("test/res/gopher-color.png").Height(50).Inline().AlignRight().Draw()
	// Add level 3 header
	doc.AddHeader3("Subsection 1.1: More examples").Draw()

	// Add text with a border
	doc.AddText("This text has a border around it.").WithBorder().Draw()

	// Compare justified vs non-justified
	doc.AddHeader1("Comparison: Normal Text vs Justified Text").Draw()

	// Normal text (left-aligned)
	doc.AddText("NORMAL TEXT (left-aligned):").Bold().Draw()
	doc.AddText("This is a sample text that demonstrates normal text flow. The text continues across multiple lines to show how words wrap naturally at the margins. This creates a simple left-aligned paragraph that is easy to read. When text is not justified, it maintains consistent spacing between words while keeping a ragged right edge.").Draw()

	// Space between examples
	doc.SpaceBefore(3)

	// Justified text
	doc.AddText("JUSTIFIED TEXT:").Bold().Draw()
	doc.AddText("This is an Example of a long paragraph of text that demonstrates justified text alignment. When text is justified, it is aligned evenly on both the left and right margins. This creates a clean, professional look that is commonly used in books, magazines, and formal documents. The spacing between words is automatically adjusted to ensure the text fills the entire width of the page from margin to margin.").Justify().Draw()
	doc.AddImage("test/res/barchart.png").Height(150).AlignCenter().Draw()
	doc.AddText("This is an Example of a long paragraph of text that demonstrates justified text alignment. When text is justified, it is aligned evenly on both the left and right margins. This creates a clean, professional look that is commonly used in books, magazines, and formal documents. The spacing between words is automatically adjusted to ensure the text fills the entire width of the page from margin to margin.").Justify().Draw()

	// Add example of table usage
	doc.AddHeader2("Section 2: Table Example").Draw()
	doc.AddText("Below is an example of a simple product table:").Draw()

	// Create a new table with headers
	table := doc.NewTable("CODE", "DESCRIPTION", "QTY.", "PRICE", "TOTAL")

	// Set column alignments (left for text, right for numbers)
	table.SetColumnAlignment(0, Left)  // CODE
	table.SetColumnAlignment(1, Left)  // DESCRIPTION
	table.SetColumnAlignment(2, Right) // QTY
	table.SetColumnAlignment(3, Right) // PRICE
	table.SetColumnAlignment(4, Right) // TOTAL

	// Add regular rows
	table.AddRow("001", "Product A", "2", "$10.00", "$20.00")
	table.AddRow("002", "Product B", "1", "$15.00", "$15.00")
	table.AddRow("003", "Product C", "3", "$5.00", "$15.00")

	// Create a styled cell for the total row
	boldStyle := CellStyle{
		TextColor: RGBColor{R: 0, G: 0, B: 128}, // Dark blue
		Font:      FontBold,
	}

	// Add a styled row for the total
	table.AddStyledRow(
		doc.NewStyledCell("", CellStyle{}),
		doc.NewStyledCell("TOTAL", boldStyle),
		doc.NewStyledCell("6", boldStyle),
		doc.NewStyledCell("", CellStyle{}),
		doc.NewStyledCell("$50.00", boldStyle),
	)

	// Draw the table
	table.Draw()

	// Add another table with right alignment
	doc.AddText("Here's another table with right alignment:").Draw()

	rightTable := doc.NewTable("Item", "Value")
	rightTable.AlignRight()
	rightTable.AddRow("Item 1", "$100.00")
	rightTable.AddRow("Item 2", "$200.00")
	rightTable.AddRow("Item 3", "$300.00")
	rightTable.Draw()

	// Add a centered footer with page number
	doc.AddPageFooter("Page").AlignCenter().WithPageNumber().Draw()

	// Create output directory if it doesn't exist
	outDir := "test/out"
	err := os.MkdirAll(outDir, 0755)
	if err != nil {
		t.Fatalf("Error creating output directory: %v", err)
	}

	// Set the output file path
	outFilePath := filepath.Join(outDir, "doc_test.pdf")

	// Save the document to the specified location
	err = doc.WritePdf(outFilePath)
	if err != nil {
		t.Fatalf("Error writing PDF: %v", err)
	}

	absPath, _ := filepath.Abs(outFilePath)
	t.Logf("PDF created successfully at: %s", absPath)
}
