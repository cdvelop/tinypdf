package tinypdf

import "strconv"

// AddHeader - add a header function, if present this will be automatically called by AddPage()
func (gp *GoPdf) AddHeader(f func()) {
	gp.headerFunc = f
}

// AddFooter - add a footer function, if present this will be automatically called by AddPage()
func (gp *GoPdf) AddFooter(f func()) {
	gp.footerFunc = f
}

// AddPageHeader adds a header to the document
func (d *Document) AddPageHeader(text string) *docText {
	// Create text builder with header style
	builder := d.newTextBuilder(text, d.fontConfig.PageHeader, FontRegular)

	// Mark as fixed position so it doesn't trigger page breaks
	builder.positioning = fixedPosition

	// Store the initial builder for reuse across pages
	d.AddHeader(func() {
		// Position at top of page with proper margin
		d.SetY(d.margins.Top / 2) // Position header in top margin area

		// Draw header
		builder.Draw()

		// Position content to start at top margin
		d.SetXY(d.margins.Left, d.margins.Top)

		// Reset inline mode if it was set by the header
		d.inlineMode = false
	})

	// Return the builder for method chaining
	return builder
}

// AddPageFooter adds a footer to the document
func (d *Document) AddPageFooter(text string) *docText {
	// Create text builder with footer style
	builder := d.newTextBuilder(text, d.fontConfig.PageFooter, FontRegular)

	// Mark as fixed position so it doesn't trigger page breaks
	builder.positioning = fixedPosition

	d.AddFooter(func() {
		// Calculate footer position - place it within bottom margin area
		pageHeight := d.config.PageSize.H
		footerY := pageHeight - (d.margins.Bottom / 2) - d.fontConfig.PageFooter.Size

		// Save current drawing position
		prevX, prevY := d.GetX(), d.GetY()

		// Position footer
		d.SetY(footerY)

		// Draw footer
		builder.Draw()

		// Restore original position after drawing footer
		d.SetXY(prevX, prevY)

		// Reset inline mode if it was set by the footer
		d.inlineMode = false
	})

	// Return the builder for method chaining
	return builder
}

// WithPageNumber adds page number to the text builder
func (dt *docText) WithPageNumber() *docText {
	// Get current text
	currentText := dt.text

	// Add page number
	if currentText != "" {
		currentText += " "
	}
	currentText += strconv.Itoa(dt.doc.GetNumberOfPages())

	// Update text in the builder
	dt.text = currentText

	return dt
}
