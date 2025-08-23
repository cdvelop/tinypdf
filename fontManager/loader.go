package fontManager

import (
	"bytes"
	"compress/zlib"

	. "github.com/cdvelop/tinystring"
)

// loadFonts is the single, shared loader for all build targets. Platform-specific
// files must provide the helper function `getFontData`.
// It is intentionally unexported; callers should create a FontManager via New
// which will invoke this during initialization.
func (fm *FontManager) loadFonts() error {

	// Recorrer directamente FontsPath
	for _, fontPath := range fm.FontsPath {
		if !HasSuffix(Convert(fontPath).ToLower().String(), ".ttf") {
			continue
		}

		fontData, err := fm.getFontData(fontPath)
		if err != nil {
			fm.Log("Warning: could not load font '%s': %v\n", fontPath, err)
			continue
		}
		// Parse TTF bytes directly and register a FontDef for the family/style
		familyName, style := parseFontName(fontPath)
		reader := fileReader{readerPosition: 0, array: fontData}
		utf8File := newUTF8Font(&reader)
		if err := utf8File.parseFile(); err != nil {
			fm.Log("Warning: could not parse ttf '%s': %v\n", fontPath, err)
			continue
		}
		desc := FontDesc{
			Ascent:       int(utf8File.Ascent),
			Descent:      int(utf8File.Descent),
			CapHeight:    utf8File.CapHeight,
			Flags:        utf8File.Flags,
			FontBBox:     FontBox{Xmin: utf8File.Bbox.Xmin, Ymin: utf8File.Bbox.Ymin, Xmax: utf8File.Bbox.Xmax, Ymax: utf8File.Bbox.Ymax},
			ItalicAngle:  utf8File.ItalicAngle,
			StemV:        utf8File.StemV,
			MissingWidth: round(utf8File.DefaultWidth),
		}

		// Compress font data for embedding (same approach as createFontDefFromTtf)
		var zbuf bytes.Buffer
		zw := zlib.NewWriter(&zbuf)
		if _, err := zw.Write(fontData); err != nil {
			zw.Close()
			fm.Log("Warning: could not compress font data for '%s': %v\n", fontPath, err)
			continue
		}
		zw.Close()

		fd := &FontDef{
			Tp:           "UTF8",
			Name:         getFontKey(fm.FontFamilyEscape(familyName), ""),
			Desc:         desc,
			Up:           int(round(utf8File.UnderlinePosition)),
			Ut:           round(utf8File.UnderlineThickness),
			Cw:           utf8File.CharWidths,
			File:         fontPath,
			OriginalSize: len(fontData),
			Data:         zbuf.Bytes(),
		}

		// Also prepare a FontDefType with utf8-specific fields used by tinypdf
		var fdt FontDefType
		fdt.Tp = "UTF8"
		fdt.Name = fd.Name
		fdt.Desc = FontDescType{Ascent: fd.Desc.Ascent, Descent: fd.Desc.Descent, CapHeight: fd.Desc.CapHeight, Flags: fd.Desc.Flags, FontBBox: fontBoxType{Xmin: fd.Desc.FontBBox.Xmin, Ymin: fd.Desc.FontBBox.Ymin, Xmax: fd.Desc.FontBBox.Xmax, Ymax: fd.Desc.FontBBox.Ymax}, ItalicAngle: fd.Desc.ItalicAngle, StemV: fd.Desc.StemV, MissingWidth: fd.Desc.MissingWidth}
		fdt.Up = fd.Up
		fdt.Ut = fd.Ut
		fdt.Cw = fd.Cw
		fdt.File = fd.File
		fdt.OriginalSize = fd.OriginalSize
		fdt.utf8File = utf8File
		// initialize UsedRunes with a sensible default subset range
		if fm.aliasNbPagesStr == "" {
			fdt.UsedRunes = makeSubsetRange(57)
		} else {
			fdt.UsedRunes = makeSubsetRange(32)
		}
		fdt.ListIndex, _ = generateFontID(fdt)
		// register in fm.fonts for potential later queries (append to slice)
		fontkey := getFontKey(fm.FontFamilyEscape(familyName), "")
		fdt.Key = fontkey
		fm.fonts = append(fm.fonts, fdt)
		// register fontFiles entries so embedding will work (store as slice entries)
		ff1 := FontFileType{Length1: int64(fd.OriginalSize), FontType: "UTF8", Key: fontkey, Content: zbuf.Bytes(), Embedded: true}
		ff2 := FontFileType{FontType: "UTF8", Key: fd.File, Content: zbuf.Bytes(), Embedded: true}
		fm.fontFiles = append(fm.fontFiles, ff1)
		fm.fontFiles = append(fm.fontFiles, ff2)

		// Find or create family entry and assign
		var family *FontFamily
		for i := range fm.fontFamilies {
			if fm.fontFamilies[i].Name == familyName {
				family = &fm.fontFamilies[i]
				break
			}
		}
		if family == nil {
			ff := FontFamily{
				Name:   familyName,
				Styles: make(map[fontStyle]*FontDef),
			}
			fm.fontFamilies = append(fm.fontFamilies, ff)
			family = &fm.fontFamilies[len(fm.fontFamilies)-1]
		}
		family.Styles[style] = fd
		if style == Regular {
			family.Regular = fd
		}
	}

	// If no font families were loaded, return an error so callers can detect
	// the failure to find/parse any fonts from the configured FontsPath.
	if len(fm.fontFamilies) == 0 {
		return Errf("no fonts loaded from FontsPath")
	}

	return nil
}

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
