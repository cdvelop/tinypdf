package fontManager

import (
	. "github.com/cdvelop/tinystring"
)

// FontManager manages the loading and accessing of fonts for the PDF document.
// It scans a directory for TTF files and makes them available to the generator.
type FontManager struct {
	fontsBasePath string
	fontFamilies  []FontFamily

	reader Reader

	log func(...any) // logging function, can be nil
}

// NewFontManager creates and initializes a new FontManager.
//
// fontsBasePath: path to the directory that contains font files (TTF/OTF).
// If an empty string is provided, the default directory "fonts/" is used.
//
// logger: optional logging function. Pass nil to disable logging. The logger,
// if provided, will be called with variadic values (similar to fmt.Println)
// for informational or debugging messages emitted by the manager.
//
// Note: NewFontManager only constructs the manager and sets its fields. It
// does not scan or load font files from disk automatically; call
// fm.LoadFonts() after creation to scan the configured directory and register
// available fonts.
//
// eg:
//
//	fm := NewFontManager("assets/myfonts/", func(a ...any) { fmt.Println(a...) })
//	if err := fm.LoadFonts(); err != nil {
//	    // handle the error
//	}
//	def, err := fm.GetFontDef("Roboto", "Bold")
//	if err != nil {
//	    // handle the error
//	}
func NewFontManager(fontsBasePath string, logger func(...any)) *FontManager {
	if fontsBasePath == "" {
		// Default runtime font directory
		fontsBasePath = "fonts/"
	}
	return &FontManager{
		fontsBasePath: fontsBasePath,
		fontFamilies:  make([]FontFamily, 0),
		log:           logger,
	}
}

// GetFontDef retrieves a font definition for a given family and style.
// If the exact style is not found, it attempts to fall back to the Regular style for that family.
func (fm *FontManager) GetFontDef(family, style string) (*FontDef, error) {

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

// ChangeFontsDir updates the font directory path and reloads the fonts.
func (fm *FontManager) ChangeFontsDir(newPath string) *FontManager {
	fm.fontsBasePath = newPath
	fm.fontFamilies = make([]FontFamily, 0)
	fm.LoadFonts()
	return fm
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
