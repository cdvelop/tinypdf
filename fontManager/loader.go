package fontManager

import (
	"bytes"
	"compress/zlib"

	. "github.com/cdvelop/tinystring"
)

// createFontDefFromTtf converts the parsed TTF data into a FontDef structure.
func createFontDefFromTtf(ttf TtfType, fontData []byte) (*FontDef, error) {
	def := &FontDef{
		Tp:   "TrueType",
		Name: ttf.PostScriptName,
		Up:   int(ttf.UnderlinePosition),
		Ut:   int(ttf.UnderlineThickness),
	}

	// Compress font data
	var zbuf bytes.Buffer
	zw := zlib.NewWriter(&zbuf)
	if _, err := zw.Write(fontData); err != nil {
		zw.Close()
		return nil, Errf("could not compress font data: %w", err)
	}
	zw.Close()
	def.Data = zbuf.Bytes()
	def.OriginalSize = len(fontData)

	// Descriptor
	k := 1000.0 / float64(ttf.UnitsPerEm)
	def.Desc = FontDesc{
		Ascent:      int(float64(ttf.TypoAscender) * k),
		Descent:     int(float64(ttf.TypoDescender) * k),
		CapHeight:   int(float64(ttf.CapHeight) * k),
		ItalicAngle: int(ttf.ItalicAngle),
		FontBBox: FontBox{
			Xmin: int(float64(ttf.Xmin) * k),
			Ymin: int(float64(ttf.Ymin) * k),
			Xmax: int(float64(ttf.Xmax) * k),
			Ymax: int(float64(ttf.Ymax) * k),
		},
		// Flags will be set later if needed, based on tinypdf's logic
	}
	if ttf.IsFixedPitch {
		def.Desc.Flags |= (1 << 0)
	}
	if ttf.Bold {
		def.Desc.Flags |= (1 << 18) // ForceBold
	}
	if def.Desc.ItalicAngle != 0 {
		def.Desc.Flags |= (1 << 6) // Italic
	}

	// Widths (for now, a simple conversion, may need more logic for encoding)
	def.Cw = make([]int, 256)
	// This part is complex because it depends on the character map (cmap).
	// The original `getInfoFromTrueType` did a complex mapping.
	// For now, we will simplify and assume a basic mapping.
	// A more robust solution would require handling the cmap properly.
	if len(ttf.Widths) > 0 {
		missingWidth := int(float64(ttf.Widths[0]) * k)
		def.Desc.MissingWidth = missingWidth
		for i := 0; i < 256; i++ {
			// This is a simplification. A real implementation needs to map
			// char codes to glyph indices and then to widths.
			if i < len(ttf.Widths) {
				def.Cw[i] = int(float64(ttf.Widths[i]) * k)
			} else {
				def.Cw[i] = missingWidth
			}
		}
	}

	return def, nil
}

// parseFontName is duplicated here from loader_std.go to avoid non-Wasm imports.
// A better solution might be to move it to a common file without os/fs imports.
func parseFontName(filename string) (family, style string) {
	basename := Convert(filename).TrimSuffix(".ttf").String()
	basename = Convert(basename).TrimSuffix(".TTF").String()

	parts := Convert(basename).Split("-")
	if len(parts) > 1 {
		style = parts[len(parts)-1]
		family = Convert(parts[:len(parts)-1]).Join("-").String()
		if len(style) > 0 {
			style = Convert(style[:1]).ToUpper().String() + Convert(style[1:]).ToLower().String()
		}
	} else {
		family = basename
		style = "Regular"
	}
	return family, style
}

// LoadFonts is the single, shared loader for all build targets. Platform-specific
// files must provide the helper function `getFontData`.
func (fm *FontManager) LoadFonts() error {
	// Clear any previously loaded fonts
	fm.fontFamilies = make([]FontFamily, 0)

	// Recorrer directamente fontsPath
	for _, fontPath := range fm.fontsPath {
		if !HasSuffix(Convert(fontPath).ToLower().String(), ".ttf") {
			continue
		}

		fontData, err := fm.getFontData(fontPath)
		if err != nil {
			if fm.log != nil {
				fm.log("Warning: could not load font '%s': %v\n", fontPath, err)
			}
			continue
		}

		// Usar la nueva API que recibe []byte directamente
		ttf, err := TtfParse(fontData)
		if err != nil {
			if fm.log != nil {
				fm.log("Warning: could not parse ttf '%s': %v\n", fontPath, err)
			}
			continue
		}

		fontDef, err := createFontDefFromTtf(ttf, fontData)
		if err != nil {
			if fm.log != nil {
				fm.log("Warning: could not create font definition for '%s': %v\n", fontPath, err)
			}
			continue
		}
		fontDef.File = fontPath

		if err := fm.setFontID(fontDef); err != nil {
			if fm.log != nil {
				fm.log("Warning: could not set font id for '%s': %v\n", fontPath, err)
			}
			continue
		}

		fontFamilyName, style := parseFontName(fontPath)

		// Find existing family or create a new one
		var family *FontFamily
		for i := range fm.fontFamilies {
			if fm.fontFamilies[i].Name == fontFamilyName {
				family = &fm.fontFamilies[i]
				break
			}
		}
		if family == nil {
			ff := FontFamily{
				Name:   fontFamilyName,
				Styles: make(map[string]*FontDef),
			}
			fm.fontFamilies = append(fm.fontFamilies, ff)
			family = &fm.fontFamilies[len(fm.fontFamilies)-1]
		}

		family.Styles[style] = fontDef
		if style == "Regular" {
			family.Regular = fontDef
		}
	}

	return nil
}
