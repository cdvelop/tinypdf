package tinypdf

// Represents an RGB color with red, green, and blue components
type RGBColor struct {
	R uint8 // Red component (0-255)
	G uint8 // Green component (0-255)
	B uint8 // Blue component (0-255)
}

// Defines the border style for a cell or table
type BorderStyle struct {
	Top      bool     // Whether to draw the top border
	Left     bool     // Whether to draw the left border
	Right    bool     // Whether to draw the right border
	Bottom   bool     // Whether to draw the bottom border
	Width    float64  // Width of the border line
	RGBColor RGBColor // Color of the border
}

// Defines the style for a cell, including border, fill, text, and font properties
type CellStyle struct {
	BorderStyle BorderStyle // Border style for the cell
	FillColor   RGBColor    // Background color of the cell
	TextColor   RGBColor    // Color of the text in the cell
	Font        string      // Font name for the cell text
	FontSize    float64     // Font size for the cell text
}

// PaintStyle represents the painting style for graphics
type PaintStyle string

const (
	// DrawPaintStyle is for drawing only
	DrawPaintStyle PaintStyle = "S"

	// FillPaintStyle is for filling only
	FillPaintStyle PaintStyle = "F"

	// DrawAndFillPaintStyle is for drawing and filling
	DrawAndFillPaintStyle PaintStyle = "B"
)

// parseStyle converts style strings to PaintStyle constants
func parseStyle(style string) PaintStyle {
	switch style {
	case "F":
		return FillPaintStyle
	case "DF", "FD":
		return DrawAndFillPaintStyle
	default: // "D" or any other string
		return DrawPaintStyle
	}
}
