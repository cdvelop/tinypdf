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
	Key   string // unique key used previously when stored in map
}

// FontFamily groups the different styles (Regular, Bold, etc.) for a single font.
type FontFamily struct {
	Name   string
	Styles map[fontStyle]*FontDef
	// Regular is a fallback for any missing styles
	Regular *FontDef
}

// Font flags for FontDescType.Flags as defined in the pdf specification.
const (
	// FontFlagFixedPitch is set if all glyphs have the same width (as
	// opposed to proportional or variable-pitch fonts, which have
	// different widths).
	FontFlagFixedPitch = 1 << 0
	// FontFlagSerif is set if glyphs have serifs, which are short
	// strokes drawn at an angle on the top and bottom of glyph stems.
	// (Sans serif fonts do not have serifs.)
	FontFlagSerif = 1 << 1
	// FontFlagSymbolic is set if font contains glyphs outside the
	// Adobe standard Latin character set. This flag and the
	// Nonsymbolic flag shall not both be set or both be clear.
	FontFlagSymbolic = 1 << 2
	// FontFlagScript is set if glyphs resemble cursive handwriting.
	FontFlagScript = 1 << 3
	// FontFlagNonsymbolic is set if font uses the Adobe standard
	// Latin character set or a subset of it.
	FontFlagNonsymbolic = 1 << 5
	// FontFlagItalic is set if glyphs have dominant vertical strokes
	// that are slanted.
	FontFlagItalic = 1 << 6
	// FontFlagAllCap is set if font contains no lowercase letters;
	// typically used for display purposes, such as for titles or
	// headlines.
	FontFlagAllCap = 1 << 16
	// SmallCap is set if font contains both uppercase and lowercase
	// letters. The uppercase letters are similar to those in the
	// regular version of the same typeface family. The glyphs for the
	// lowercase letters have the same shapes as the corresponding
	// uppercase letters, but they are sized and their proportions
	// adjusted so that they have the same size and stroke weight as
	// lowercase glyphs in the same typeface family.
	SmallCap = 1 << 18
	// ForceBold determines whether bold glyphs shall be painted with
	// extra pixels even at very small text sizes by a conforming
	// reader. If the ForceBold flag is set, features of bold glyphs
	// may be thickened at small text sizes.
	ForceBold = 1 << 18
)
