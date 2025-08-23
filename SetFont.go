package tinypdf

import (
	"github.com/cdvelop/tinypdf/fontManager"
	. "github.com/cdvelop/tinystring"
)

// SetFont sets the font used to print character strings. It is mandatory to
// call this method at least once before printing text or the resulting
// document will not be valid.
//
// The font can be either a standard one or a font added via the AddFont()
// method or AddFontFromReader() method. Standard fonts use the Windows
// encoding cp1252 (Western Europe).
//
// The method can be called before the first page is created and the font is
// kept from page to page. If you just wish to change the current font size, it
// is simpler to call SetFontSize().
//
// Note: the font definition file must be accessible. An error is set if the
// file cannot be read.
//
// familyStr specifies the font family. It can be either a name defined by
// AddFont(), AddFontFromReader() or one of the standard families (case
// insensitive): "Courier" for fixed-width, "Helvetica" or "Arial" for sans
// serif, "Times" for serif, "Symbol" or "ZapfDingbats" for symbolic.
//
// styleStr can be "B" (bold), "I" (italic), "U" (underscore), "S" (strike-out)
// or any combination. The default value (specified with an empty string) is
// regular. Bold and italic styles do not apply to Symbol and ZapfDingbats.
//
// size is the font size measured in points. The default value is the current
// size. If no size has been specified since the beginning of the document, the
// value taken is 12.
func (f *TinyPDF) SetFont(familyStr, styleStr string, size float64) {
	// dbg("SetFont x %.2f, lMargin %.2f", f.x, f.lMargin)

	if f.err != nil {
		return
	}
	// dbg("SetFont")
	familyStr = f.fontFamilyEscape(familyStr)
	var ok bool
	if familyStr == "" {
		familyStr = f.GetFontFamily()
	} else {
		familyStr = Convert(familyStr).ToLower().String()
	}
	styleStr = Convert(styleStr).ToUpper().String()
	f.underline = Contains(styleStr, "U")
	if f.underline {
		styleStr = Convert(styleStr).Replace("U", "").String()
	}
	f.strikeout = Contains(styleStr, "S")
	if f.strikeout {
		styleStr = Convert(styleStr).Replace("S", "").String()
	}
	if styleStr == "IB" {
		styleStr = "BI"
	}
	if size == 0.0 {
		size = f.GetFontSizePt()
	}

	// Test if font is already loaded
	fontKey := familyStr + styleStr
	_, ok = f.findFont(fontKey)
	if !ok {
		// Handle core-font aliases (arial -> helvetica) and symbolic fonts.
		if familyStr == "arial" {
			familyStr = "helvetica"
		}
		// The previous implementation used a coreFonts map. That field was
		// removed; fonts (including standard/core fonts) must be registered
		// in the FontManager at initialization. Try to locate the font in the
		// FontManager and, if found, register it into this TinyPDF instance's
		// fonts slice so it can be used by the document.
		var fmDef fontManager.FontDefType
		// Try direct lookup in the font manager
		if fmDef, ok = f.fm.FindFontByKey(fontKey); ok {
			// ensure document has the font registered
			if _, found := f.findFont(fontKey); !found {
				f.fonts = append(f.fonts, fmDef)
			}
		} else {
			// try aliases and symbolic font mappings
			if familyStr == "symbol" {
				familyStr = "zapfdingbats"
			}
			if familyStr == "zapfdingbats" {
				styleStr = ""
			}
			fontKey = familyStr + styleStr
			if fmDef, ok = f.fm.FindFontByKey(fontKey); ok {
				if _, found := f.findFont(fontKey); !found {
					f.fonts = append(f.fonts, fmDef)
				}
			} else {
				f.err = Errf("undefined font: %s %s", familyStr, styleStr)
				return
			}
		}
	}
	// Select it
	f.fontFamily = familyStr
	f.textDecoration = styleStr
	f.fontSizePt = size
	f.fontSize = size / f.ConversionRatio()
	ff, _ := f.findFont(fontKey)
	f.currentFont = ff
	if f.currentFont.Tp == "UTF8" {
		f.isCurrentUTF8 = true
	} else {
		f.isCurrentUTF8 = false
	}
	if f.page > 0 {
		f.outf("BT /F%s %.2f Tf ET", f.currentFont.ListIndex, f.GetFontSizePt())
	}
}
