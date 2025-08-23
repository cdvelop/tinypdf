package tinypdf

// SetFontStyle sets the style of the current font. See also SetFont()
func (f *TinyPDF) SetFontStyle(styleStr string) {
	f.SetFont(f.fontFamily, styleStr, f.fontSizePt)
}

// GetFontFamily returns the family of the current font. See SetFont() for details.
func (f *TinyPDF) GetFontFamily() string {
	return f.fontFamily
}

// SetFontSize defines the size of the current font. Size is specified in
// points (1/ 72 inch). See also SetFontUnitSize().
func (f *TinyPDF) SetFontSize(size float64) {
	f.fontSizePt = size
	f.fontSize = size / f.ConversionRatio()
	if f.page > 0 {
		f.outf("BT /F%s %.2f Tf ET", f.currentFont.ListIndex, f.fontSizePt)
	}
}

// SetFontUnitSize defines the size of the current font. Size is specified in
// the Unit of measure specified in New(). See also SetFontSize().
func (f *TinyPDF) SetFontUnitSize(size float64) {
	f.fontSizePt = size * f.ConversionRatio()
	f.fontSize = size
	if f.page > 0 {
		f.outf("BT /F%s %.2f Tf ET", f.currentFont.ListIndex, f.fontSizePt)
	}
}

// GetFontSizes returns the size of the current font in points followed by the
// size in the Unit of measure specified in New(). The second value can be used
// as a line height value in drawing operations.
func (f *TinyPDF) GetFontSizes() (ptSize, unitSize float64) {
	return f.fontSizePt, f.fontSize
}

// GetFontSize returns the current font size in the user Unit (single value).
// This convenience method allows calling code to use `f.fm.GetFontSize()`
// when only the Unit-size is required. For both values use `GetFontSizes()`.
func (f *TinyPDF) GetFontSize() float64 {
	return f.fontSize
}

// GetFontSizePt returns the current font size in points (pt).
func (f *TinyPDF) GetFontSizePt() float64 {
	return f.fontSizePt
}
