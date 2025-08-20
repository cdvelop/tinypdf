package fontManager

import (
	. "github.com/cdvelop/tinystring"
)

// FontManager manages the loading and accessing of fonts for the PDF document.
// It scans a directory for TTF files and makes them available to the generator.
type FontManager struct {
	fontsPath    []string // List of font paths/URLs to load
	fontFamilies []FontFamily
	reader       osFile
	log          func(...any) // logging function, can be nil
}

// NewFontManager creates and initializes a new FontManager.
//
// fontsPath: slice of font file paths or URLs to load.
//
// For WASM builds, use URLs:
//
//	fontsPath := []string{"fonts/arial.ttf", "fonts/helvetica.ttf", "fonts/times.ttf"}
//
// For Server/Desktop builds, use file paths:
//
//	fontsPath := []string{"./fonts/arial.ttf", "/usr/share/fonts/truetype/arial.ttf"}
//
// Or use a directory pattern (for auto-discovery in non-WASM):
//
//	fontsPath := []string{"./fonts/"} // Will scan directory
//
// logger: optional logging function. Pass nil to disable logging.
func NewFontManager(fontsPath []string, logger func(...any)) *FontManager {
	return &FontManager{
		fontsPath:    fontsPath,
		fontFamilies: make([]FontFamily, 0),
		log:          logger,
	}
}

// getFontList
func (fm *FontManager) getFontList() ([]string, error) {
	if len(fm.fontsPath) == 0 {
		return nil, Err("no font paths provided for WASM build")
	}
	return fm.fontsPath, nil
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
