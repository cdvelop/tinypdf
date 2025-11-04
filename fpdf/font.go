package fpdf

import (
	"encoding/json"

	. "github.com/cdvelop/tinystring"
)

// AddFont loads a TrueType font and makes it available. It is necessary to
// call this method for every font that is used in the document.
func (f *Fpdf) AddFont(family, style, fileStr string) {
	if f.err != nil {
		return
	}
	family = Convert(family).ToLower().String()
	if family == "arial" {
		family = "helvetica"
	}
	style = Convert(style).ToUpper().String()
	if style == "IB" {
		style = "BI"
	}
	fontkey := family + style
	if _, ok := f.fonts[fontkey]; ok {
		return
	}

	fontPath := fileStr
	if fontPath == "" {
		fontPath = Convert(style).Replace("I", "i").Replace("B", "b").String()
		fontPath = family + fontPath + ".ttf"
	}

	// Check cache first
	fontData, found := f.getCachedFont(fontPath)
	if !found {
		var err error
		fontData, err = f.fontLoader(fontPath)
		if err != nil {
			f.SetError(Errf("failed to load font %s: %w", fontPath, err))
			return
		}
		f.addFontToCache(fontPath, fontData)
	}

	ttf, err := TtfParseBytes(fontData)
	if err != nil {
		f.SetError(err)
		return
	}

	var def fontDefType
	def.Tp = "TrueType"
	def.Name = ttf.PostScriptName
	def.Desc.Ascent = int(ttf.TypoAscender)
	def.Desc.Descent = int(ttf.TypoDescender)
	def.Desc.CapHeight = int(ttf.CapHeight)
	def.Desc.Flags = 1<<5 | 1<<2 // Nonsymbolic and Symbolic flags
	def.Desc.FontBBox.Xmin = int(ttf.Xmin)
	def.Desc.FontBBox.Ymin = int(ttf.Ymin)
	def.Desc.FontBBox.Xmax = int(ttf.Xmax)
	def.Desc.FontBBox.Ymax = int(ttf.Ymax)
	def.Desc.ItalicAngle = int(ttf.ItalicAngle)
	def.Desc.StemV = 70
	if ttf.Bold {
		def.Desc.StemV = 120
	}
	def.Up = int(ttf.UnderlinePosition)
	def.Ut = int(ttf.UnderlineThickness)
	def.Cw = make([]int, 256)
	for i := 0; i < 256; i++ {
		def.Cw[i] = 600 // Default width
	}
	def.File = fontPath
	def.OriginalSize = len(fontData)
	def.usedRunes = make(map[int]int)

	def.i = Fmt("%d", len(f.fonts)+1)
	f.fonts[fontkey] = def
}

// SetFont sets the font used to print character strings. It is mandatory to
// call this method at least once before printing text or the resulting
// document will not be valid.
//
// The font can be either a standard one or a font added by AddFont(). Standard
// fonts use the Windows encoding cp1252 (Western Europe).
//
// The font is selected by specifying its family and a style in any
// combination.
//
// It is also possible to modify the current size. If the given size is not
// specified (or is zero), the size of the current font is retained.
//
// If a font has not been loaded by a prior call to AddFont(), an error is
// returned.
func (f *Fpdf) SetFont(familyStr, styleStr string, size float64) {
	if f.err != nil {
		return
	}
	familyStr = Convert(familyStr).ToLower().String()
	if familyStr == "arial" {
		familyStr = "helvetica"
	} else if familyStr == "" {
		familyStr = f.fontFamily
	}
	styleStr = Convert(styleStr).ToUpper().String()
	if styleStr == "IB" {
		styleStr = "BI"
	}
	fontkey := familyStr + styleStr
	if _, ok := f.fonts[fontkey]; !ok {
		f.err = Errf("font not found: %s %s", familyStr, styleStr)
		return
	}
	f.fontFamily = familyStr
	f.fontStyle = styleStr
	f.currentFont = f.fonts[fontkey]
	f.SetFontSize(size)
	f.isCurrentUTF8 = true
}

// SetFontSize defines the size of the current font. size is specified in the
// unit of measure passed to New().
func (f *Fpdf) SetFontSize(size float64) {
	if size == 0.0 {
		size = f.fontSize
	} else {
		f.fontSize = size
	}
	if f.k > 0 {
		f.fontSizePt = f.fontSize * f.k
	}
}

// GetFontSize returns the size of the current font in both points (pt) and
// the unit of measure specified in New() (u).
func (f *Fpdf) GetFontSize() (pt, u float64) {
	return f.fontSizePt, f.fontSize
}


func (f *Fpdf) loadFontDef(def *fontDefType, data []byte) error {
	return json.Unmarshal(data, def)
}
