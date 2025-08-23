package tinypdf

import (
	"github.com/cdvelop/tinypdf/fontManager"
	. "github.com/cdvelop/tinystring"
)

// ConversionRatio returns the current scale factor k (number of points per user Unit).
func (f *TinyPDF) ConversionRatio() float64 {
	return f.k
}

func (f *TinyPDF) Font() *fontManager.FontManager {
	return f.fm
}

// GetObjectNumber returns the current PDF object number.
func (f *TinyPDF) CurrentObjectNumber() int {
	return f.n
}

// Condition font family string to PDF name compliance. See section 5.3 (Names)
// in https://resources.infosecinstitute.com/pdf-file-format-basic-structure/
func (f *TinyPDF) fontFamilyEscape(familyStr string) (escStr string) {
	escStr = Convert(familyStr).Replace(" ", "#20", -1).String()
	// Additional replacements can take place here
	return
}

// SetUnit sets the unit type for the TinyPDF instance and updates the
// conversion ratio accordingly.
func (f *TinyPDF) SetUnit(u Unit) {
	f.unitType = u
	switch f.unitType {
	case POINT:
		f.k = 1.0
	case MM:
		f.k = 72.0 / 25.4
	case CM:
		f.k = 72.0 / 2.54
	case IN:
		f.k = 72.0
	default:
		// leave unchanged
	}
}
