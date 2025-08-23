package fontManager

import (
	. "github.com/cdvelop/tinystring"
)

// GetFontDesc returns the font descriptor, which can be used for
// example to find the baseline of a font. If familyStr is empty
// current font descriptor will be returned.
// See FontDescType for documentation about the font descriptor.
// See AddFont for details about familyStr and styleStr.
func (f *FontManager) GetFontDesc(familyStr, styleStr string) FontDescType {
	if familyStr == "" {
		return FontDescType{}
	}
	key := getFontKey(f.FontFamilyEscape(familyStr), styleStr)
	if def, ok := f.FindFontByKey(key); ok {
		return def.Desc
	}
	return FontDescType{}
}

// AllFonts returns a copy of the internal font definitions slice.
func (f *FontManager) AllFonts() []FontDefType {
	out := make([]FontDefType, len(f.fonts))
	copy(out, f.fonts)
	return out
}

// FindFontByKey searches the font slice for an entry with matching Key.
func (f *FontManager) FindFontByKey(key string) (FontDefType, bool) {
	for _, fd := range f.fonts {
		if fd.Key == key {
			return fd, true
		}
	}
	return FontDefType{}, false
}

// AllFiles returns a copy of the internal font files slice.
func (f *FontManager) AllFiles() []FontFileType {
	out := make([]FontFileType, len(f.fontFiles))
	copy(out, f.fontFiles)
	return out
}

// FindFileByKey searches the font files slice for an entry with matching Key.
func (f *FontManager) FindFileByKey(key string) (FontFileType, bool) {
	for _, ff := range f.fontFiles {
		if ff.Key == key {
			return ff, true
		}
	}
	return FontFileType{}, false
}

// ToUnicodeMap returns the static header for ToUnicode CMap used in PDFs
func (f *FontManager) ToUnicodeMap() string {
	return toUnicode
}

// AllDiffs returns a copy of the encoding differences array.
func (f *FontManager) AllDiffs() []string {
	out := make([]string, len(f.diffs))
	copy(out, f.diffs)
	return out
}

// getFontKey is used by AddFontFromReader and GetFontDesc
// SetFont proxies font selection to the owning TinyPDF instance via the
// SetFontCB callback provided in Config. This allows existing call sites
// that call pdf.Font().SetFont(...) to continue working while font loading
// is handled at initialization.
func (f *FontManager) SetFont(familyStr, styleStr string, size float64) {
	if f.Config != nil && f.Config.SetFontCB != nil {
		f.Config.SetFontCB(familyStr, styleStr, size)
	}
}

// getFontKey is used by GetFontDesc
func getFontKey(familyStr, styleStr string) string {
	familyStr = Convert(familyStr).ToLower().String()
	styleStr = Convert(styleStr).ToUpper().String()
	if styleStr == "IB" {
		styleStr = "BI"
	}
	return familyStr + styleStr
}
