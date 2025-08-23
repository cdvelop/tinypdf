package fontManager

// FontDescType (font descriptor) specifies metrics and other
// attributes of a font, as distinct from the metrics of individual
// glyphs (as defined in the pdf specification).
type FontDescType struct {
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
	FontBBox fontBoxType
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

type FontDefType struct {
	Tp           string        // "Core", "TrueType", ...
	Name         string        // "Courier-Bold", ...
	Desc         FontDescType  // Font descriptor
	Up           int           // Underline position
	Ut           int           // Underline thickness
	Cw           []int         // Character width by ordinal
	Enc          string        // "cp1252", ...
	Diff         string        // Differences from reference encoding
	File         string        // "Redressed.z"
	Size1, Size2 int           // Type1 values
	OriginalSize int           // Size of uncompressed font file
	N            int           // Set by font loader
	DiffN        int           // Position of diff in app array, set by font loader
	ListIndex    string        // 1-based position in font list, set by font loader, not this program
	utf8File     *utf8FontFile // UTF-8 font
	Utf8File     *utf8FontFile // exported pointer so tinypdf can call GenerateCutFont
	UsedRunes    map[int]int   // Array of used runes
	Key          string        // unique key (was map key) for lookup
}

type fontBoxType struct {
	Xmin, Ymin, Xmax, Ymax int
}

type FontFileType struct {
	Length1, Length2 int64
	N                int
	Embedded         bool
	Content          []byte
	FontType         string
	Key              string // the key under which this file was previously stored
}
