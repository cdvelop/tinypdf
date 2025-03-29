package tinypdf

import (
	"image"
	"os"
	"path/filepath"
)

// imageElement represents an image to be added to the document
type imageElement struct {
	doc       *Document
	path      string
	width     float64
	height    float64
	keepRatio bool
	alignment int
	x, y      float64
	hasPos    bool
	inline    bool // New property to track inline status
	valign    int  // Vertical alignment for inline images
}

// AddImage creates a new image element
func (doc *Document) AddImage(path string) *imageElement {
	// Use absolute path if provided path is relative
	absolutePath, err := filepath.Abs(path)
	if err == nil && fileExists(absolutePath) {
		path = absolutePath
	}

	return &imageElement{
		doc:       doc,
		path:      path,
		keepRatio: true,
		alignment: Left,
	}
}

// Width sets the image width and maintains aspect ratio if height is not set
// eg: img.Width(50) will set the width to 50 and calculate height based on aspect ratio
func (img *imageElement) Width(w float64) *imageElement {
	img.width = w
	return img
}

// Height sets the image height and maintains aspect ratio if width is not set
// eg: img.Height(50) will set the height to 100 and calculate width based on aspect ratio
func (img *imageElement) Height(h float64) *imageElement {
	img.height = h
	return img
}

// Size sets both width and height explicitly (no aspect ratio preservation)
// eg: img.Size(50, 100) will set the width to 50 and height to 100
func (img *imageElement) Size(w, h float64) *imageElement {
	img.width = w
	img.height = h
	img.keepRatio = false
	return img
}

// FixedPosition places the image at specific coordinates
func (img *imageElement) FixedPosition(x, y float64) *imageElement {
	img.x = x
	img.y = y
	img.hasPos = true
	return img
}

// AlignLeft aligns the image to the left margin
func (img *imageElement) AlignLeft() *imageElement {
	img.alignment = Left
	return img
}

// AlignCenter centers the image horizontally
func (img *imageElement) AlignCenter() *imageElement {
	img.alignment = Center
	return img
}

// AlignRight aligns the image to the right margin
func (img *imageElement) AlignRight() *imageElement {
	img.alignment = Right
	return img
}

// Inline makes the image display inline with text rather than breaking to a new line
// The text will continue from the right side of the image
func (img *imageElement) Inline() *imageElement {
	img.inline = true
	return img
}

// VerticalAlignTop aligns the image with the top of the text line when inline
func (img *imageElement) VerticalAlignTop() *imageElement {
	img.valign = 0
	return img
}

// VerticalAlignMiddle aligns the image with the middle of the text line when inline
func (img *imageElement) VerticalAlignMiddle() *imageElement {
	img.valign = 1
	return img
}

// VerticalAlignBottom aligns the image with the bottom of the text line when inline
func (img *imageElement) VerticalAlignBottom() *imageElement {
	img.valign = 2
	return img
}

// Draw renders the image on the document to include page break handling
func (img *imageElement) Draw() error {
	// Get image dimensions to calculate aspect ratio if needed
	imgWidth, imgHeight, err := img.getImageDimensions()
	if err != nil {
		return err
	}

	// Calculate final dimensions
	finalWidth, finalHeight := img.calculateDimensions(imgWidth, imgHeight)

	// Check if the image has a fixed position
	if !img.hasPos {
		// HERE IS THE NEW PART: Check if the image fits on current page
		newY, _ := img.doc.ensureElementFits(finalHeight)

		// Only update Y position if this is not an inline element
		if !img.inline {
			img.doc.SetY(newY)
		}
	}

	// Determine position (after possible page break)
	x, y := img.calculatePosition(finalWidth)

	// Adjust vertical position for inline images based on alignment
	if img.inline {
		lineHeight := img.doc.GetLineHeight()

		switch img.valign {
		case 0: // Top alignment
			// No adjustment needed
		case 1: // Middle alignment
			y = y + (lineHeight-finalHeight)/2
		case 2: // Bottom alignment
			y = y + lineHeight - finalHeight
		default:
			// Default to middle alignment
			y = y + (lineHeight-finalHeight)/2
		}
	}

	// Create rectangle for the image
	rect := &Rect{
		W: finalWidth,
		H: finalHeight,
	}

	// Draw the image using the underlying GoPdf instance
	err = img.doc.Image(img.path, x, y, rect)
	if err != nil {
		return err
	}

	// Handle position updates based on inline setting
	if img.inline {
		// For inline images, advance X position but keep Y unchanged
		img.doc.SetX(x + finalWidth)

		// Store that we have an inline element active
		img.doc.inlineMode = true
	} else {
		// For block images, advance Y position to avoid text overlapping with the image
		if !img.hasPos {
			img.doc.newLineBreakBasedOnDefaultFont(y + finalHeight)
		}

		// Reset X position to left margin since this is a block element
		img.doc.SetX(img.doc.margins.Left)

		// Reset inline mode
		img.doc.inlineMode = false
	}

	return nil
}

// getImageDimensions returns the natural width and height of the image
func (img *imageElement) getImageDimensions() (float64, float64, error) {
	file, err := os.Open(img.path)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	imgConfig, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return float64(imgConfig.Width), float64(imgConfig.Height), nil
}

// calculateDimensions determines the final width and height of the image
func (img *imageElement) calculateDimensions(imgWidth, imgHeight float64) (float64, float64) {
	// Default to original dimensions
	finalWidth := imgWidth
	finalHeight := imgHeight

	// Scale down if original image is too large
	pageWidth := img.doc.pageWidth - img.doc.margins.Left - img.doc.margins.Right
	if finalWidth > pageWidth {
		ratio := pageWidth / finalWidth
		finalWidth = pageWidth
		finalHeight = finalHeight * ratio
	}

	// Apply user-specified dimensions
	if img.width > 0 && img.height > 0 {
		// Both dimensions specified
		finalWidth = img.width
		finalHeight = img.height
	} else if img.width > 0 && img.keepRatio {
		// Only width specified, calculate height to maintain aspect ratio
		ratio := img.width / finalWidth
		finalWidth = img.width
		finalHeight = finalHeight * ratio
	} else if img.height > 0 && img.keepRatio {
		// Only height specified, calculate width to maintain aspect ratio
		ratio := img.height / finalHeight
		finalHeight = img.height
		finalWidth = finalWidth * ratio
	}

	return finalWidth, finalHeight
}

// calculatePosition determines where to place the image
func (img *imageElement) calculatePosition(width float64) (float64, float64) {
	if img.hasPos {
		return img.x, img.y
	}

	x := img.doc.margins.Left
	y := img.doc.GetY()

	// Apply alignment
	availableWidth := img.doc.pageWidth
	switch img.alignment {
	case Center:
		x = img.doc.margins.Left + (availableWidth-width)/2
	case Right:
		x = img.doc.margins.Left + availableWidth - width
	}

	return x, y
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
