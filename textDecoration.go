package tinypdf

// GetFontDecorations returns decorations (underline/strikeout) appended to
// the font style flags. The value is the style string (e.g. "BI") plus
// optional decoration letters: "U" for underline and "S" for strikeout.
//
// This clarifies that 'style' here is a set of flags (Bold/Italic) and that
// underline/strikeout are decorations. To preserve the public API the
// original method name is kept as a thin wrapper (see below).
func (f *TinyPDF) GetTextDecorations() string {
	styleStr := f.textDecoration

	if f.underline {
		styleStr += "U"
	}
	if f.strikeout {
		styleStr += "S"
	}

	return styleStr
}
