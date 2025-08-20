package fontManager

// FontBox defines the coordinates and extent of the various page box types
type FontBox struct {
	Xmin, Ymin, Xmax, Ymax int
}

// FontDesc (font descriptor) specifies metrics and other
// attributes of a font, as distinct from the metrics of individual
// glyphs (as defined in the pdf specification).
type FontDesc struct {
	// The maximum height above the baseline reached by glyphs in this
	// font (for example for "S"). The height of glyphs for accented
	// characters shall be excluded.
	Ascent int
	// The maximum depth below the baseline reached by glyphs in this
	// font. The value shall be a negative number.
	Descent int
	// The vertical coordinate of the top of flat capital letters,
	// measured from the baseline (for example "H").
	CapHeight int
	// A collection of flags defining various characteristics of the
	// font. (See the FontFlag* constants.)
	Flags int
	// A rectangle, expressed in the glyph coordinate system, that
	// shall specify the font bounding box. This should be the smallest
	// rectangle enclosing the shape that would result if all of the
	// glyphs of the font were placed with their origins coincident
	// and then filled.
	FontBBox FontBox
	// The angle, expressed in degrees counterclockwise from the
	// vertical, of the dominant vertical strokes of the font. (The
	// 9-o’clock position is 90 degrees, and the 3-o’clock position
	// is –90 degrees.) The value shall be negative for fonts that
	// slope to the right, as almost all italic fonts do.
	ItalicAngle int
	// The thickness, measured horizontally, of the dominant vertical
	// stems of glyphs in the font.
	StemV int
	// The width to use for character codes whose widths are not
	// specified in a font dictionary’s Widths array. This shall have
	// a predictable effect only if all such codes map to glyphs whose
	// actual widths are the same as the value of the MissingWidth
	// entry. (Default value: 0.)
	MissingWidth int
}

// FontDef contains all the information needed by the PDF generator to use a font.
type FontDef struct {
	Tp           string   // "Core", "TrueType", ...
	Name         string   // "Courier-Bold", ...
	Desc         FontDesc // Font descriptor
	Up           int      // Underline position
	Ut           int      // Underline thickness
	Cw           []int    // Character width by ordinal
	Enc          string   // "cp1252", ...
	Diff         string   // Differences from reference encoding
	File         string   // "Redressed.z"
	Size1, Size2 int      // Type1 values
	OriginalSize int      // Size of uncompressed font file
	Data         []byte   // Raw, compressed font data for embedding.

	// Fields used by the PDF generator
	I     string // Unique font identifier, used for /F... resources
	N     int    // PDF object number for the font
	DiffN int    // Position of diff in app array
}

// FontFamily groups the different styles (Regular, Bold, etc.) for a single font.
type FontFamily struct {
	Name   string
	Styles map[string]*FontDef
	// Regular is a fallback for any missing styles
	Regular *FontDef
}
