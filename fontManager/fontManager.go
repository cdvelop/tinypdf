package fontManager

import (
	. "github.com/cdvelop/tinystring"
)

// FontManager manages the loading and accessing of fonts for the PDF document.
// It scans a directory for TTF files and makes them available to the generator.
type FontManager struct {
	*Config

	fontFamilies []FontFamily

	// fontFiles holds entries for font files available for embedding.
	// Using a slice keeps this layout compatible with tinygo and
	// avoids map iteration order issues; each entry has a Key field.
	fontFiles []FontFileType // slice of font files

	// fonts holds detailed font definitions (used by PDF generator logic)
	// stored as a slice; each FontDefType has a Key field for lookup.
	fonts []FontDefType

	// aliasNbPagesStr mirrors the same configuration used by tinypdf; kept
	// here for compatibility with existing font helpers that reference it.
	aliasNbPagesStr string

	err   error
	diffs []string // array of encoding differences
}

type Config struct {
	Log                 func(...any) // logging function
	FontsPath           []string     // list of font paths eg: ["./fonts/arial.ttf"]
	ConversionRatio     func() float64
	CurrentObjectNumber func() int // returns the current PDF object number
	// SetFontCB is an optional callback that will be called when callers
	// invoke SetFont() on the FontManager. This allows the FontManager to
	// proxy font selection back to the TinyPDF instance without circular
	// dependencies between packages.
	SetFontCB func(family, style string, size float64)
	// FontFamilyEscape is a helper function to ensure font family strings are
	// compliant with PDF naming conventions.
	FontFamilyEscape func(familyStr string) (escStr string)
}

func New(c *Config) *FontManager {

	// Initialize font manager
	fm := &FontManager{
		Config:       c,
		fontFiles:    make([]FontFileType, 0),
		fonts:        make([]FontDefType, 0),
		fontFamilies: make([]FontFamily, 0),
		diffs:        make([]string, 0, 8),
	}

	// Load fonts during initialization. Keep the error on the manager so callers
	// can inspect it if needed; also log a warning if a logger is provided.
	if fm.err = fm.loadFonts(); fm.err != nil {
		fm.Log("Error: FontManager.loadFonts failed: %v", fm.err)
	}

	return fm
}

// GetFontDef retrieves a font definition for a given family and style.
// If the exact style is not found, it attempts to fall back to the Regular style for that family.
func (fm *FontManager) GetFontDef(family string, style fontStyle) (*FontDef, error) {

	// Find family by name in the slice
	var fontFamily *FontFamily
	for i := range fm.fontFamilies {
		if fm.fontFamilies[i].Name == family {
			fontFamily = &fm.fontFamilies[i]
			break
		}
	}
	if fontFamily == nil {
		return nil, Errf("font family '%s' not found", family)
	}

	fontDef, ok := fontFamily.Styles[style]
	if !ok {
		// Fallback to regular style if the requested style is not available
		if fontFamily.Regular != nil {
			return fontFamily.Regular, nil
		}
		return nil, Errf("font style '%s' not found for family '%s' and no regular fallback is available", style, family)
	}

	return fontDef, nil
}

// GetAllFontDefs returns a slice of all loaded font definitions.
func (fm *FontManager) GetAllFontDefs() []*FontDef {

	var defs []*FontDef
	for i := range fm.fontFamilies {
		for _, def := range fm.fontFamilies[i].Styles {
			defs = append(defs, def)
		}
	}
	return defs
}
