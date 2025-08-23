package tinypdf

import (
	. "github.com/cdvelop/tinystring"
)

func byteBound(v byte) byte {
	if v > 100 {
		return 100
	}
	return v
}

// AddSpotColor adds an ink-based CMYK color to the gofpdf instance and
// associates it with the specified name. The individual components specify
// percentages ranging from 0 to 100. Values above this are quietly capped to
// 100. An error occurs if the specified name is already associated with a
// color.
func (f *TinyPDF) AddSpotColor(nameStr string, c, m, y, k byte) {
	if f.err == nil {
		_, ok := f.spotColorMap[nameStr]
		if !ok {
			id := len(f.spotColorMap) + 1
			f.spotColorMap[nameStr] = spotColorType{
				id: id,
				val: cmykColorType{
					c: byteBound(c),
					m: byteBound(m),
					y: byteBound(y),
					k: byteBound(k),
				},
			}
		} else {
			f.err = Errf("name \"%s\" is already associated with a spot color", nameStr)
		}
	}
}

func (f *TinyPDF) getSpotColor(nameStr string) (clr spotColorType, ok bool) {
	if f.err == nil {
		clr, ok = f.spotColorMap[nameStr]
		if !ok {
			f.err = Errf("spot color name \"%s\" is not registered", nameStr)
		}
	}
	return
}

// SetDrawSpotColor sets the current draw color to the spot color associated
// with nameStr. An error occurs if the name is not associated with a color.
// The value for tint ranges from 0 (no intensity) to 100 (full intensity). It
// is quietly bounded to this range.
func (f *TinyPDF) SetDrawSpotColor(nameStr string, tint byte) {
	var clr spotColorType
	var ok bool

	clr, ok = f.getSpotColor(nameStr)
	if ok {
		f.color.draw.mode = colorModeSpot
		f.color.draw.spotStr = nameStr
		f.color.draw.str = Fmt("/CS%d CS %.3f SCN", clr.id, float64(byteBound(tint))/100)
		if f.page > 0 {
			f.out(f.color.draw.str)
		}
	}
}

// SetFillSpotColor sets the current fill color to the spot color associated
// with nameStr. An error occurs if the name is not associated with a color.
// The value for tint ranges from 0 (no intensity) to 100 (full intensity). It
// is quietly bounded to this range.
func (f *TinyPDF) SetFillSpotColor(nameStr string, tint byte) {
	var clr spotColorType
	var ok bool

	clr, ok = f.getSpotColor(nameStr)
	if ok {
		f.color.fill.mode = colorModeSpot
		f.color.fill.spotStr = nameStr
		f.color.fill.str = Fmt("/CS%d cs %.3f scn", clr.id, float64(byteBound(tint))/100)
		f.colorFlag = f.color.fill.str != f.color.text.str
		if f.page > 0 {
			f.out(f.color.fill.str)
		}
	}
}

// SetTextSpotColor sets the current text color to the spot color associated
// with nameStr. An error occurs if the name is not associated with a color.
// The value for tint ranges from 0 (no intensity) to 100 (full intensity). It
// is quietly bounded to this range.
func (f *TinyPDF) SetTextSpotColor(nameStr string, tint byte) {
	var clr spotColorType
	var ok bool

	clr, ok = f.getSpotColor(nameStr)
	if ok {
		f.color.text.mode = colorModeSpot
		f.color.text.spotStr = nameStr
		f.color.text.str = Fmt("/CS%d cs %.3f scn", clr.id, float64(byteBound(tint))/100)
		f.colorFlag = f.color.fill.str != f.color.text.str
	}
}

func (f *TinyPDF) returnSpotColor(clr colorType) (name string, c, m, y, k byte) {
	var spotClr spotColorType
	var ok bool

	name = clr.spotStr
	if name != "" {
		spotClr, ok = f.getSpotColor(name)
		if ok {
			c = spotClr.val.c
			m = spotClr.val.m
			y = spotClr.val.y
			k = spotClr.val.k
		}
	}
	return
}

// GetDrawSpotColor returns the most recently used spot color information for
// drawing. This will not be the current drawing color if some other color type
// such as RGB is active. If no spot color has been set for drawing, zero
// values are returned.
func (f *TinyPDF) GetDrawSpotColor() (name string, c, m, y, k byte) {
	return f.returnSpotColor(f.color.draw)
}

// GetTextSpotColor returns the most recently used spot color information for
// text output. This will not be the current text color if some other color
// type such as RGB is active. If no spot color has been set for text, zero
// values are returned.
func (f *TinyPDF) GetTextSpotColor() (name string, c, m, y, k byte) {
	return f.returnSpotColor(f.color.text)
}

// GetFillSpotColor returns the most recently used spot color information for
// fill output. This will not be the current fill color if some other color
// type such as RGB is active. If no fill spot color has been set, zero values
// are returned.
func (f *TinyPDF) GetFillSpotColor() (name string, c, m, y, k byte) {
	return f.returnSpotColor(f.color.fill)
}

func (f *TinyPDF) putSpotColors() {
	for k, v := range f.spotColorMap {
		f.newobj()
		f.outf("[/Separation /%s", Convert(k).Replace(" ", "#20", -1))
		f.out("/DeviceCMYK <<")
		f.out("/Range [0 1 0 1 0 1 0 1] /C0 [0 0 0 0] ")
		f.outf("/C1 [%.3f %.3f %.3f %.3f] ", float64(v.val.c)/100, float64(v.val.m)/100,
			float64(v.val.y)/100, float64(v.val.k)/100)
		f.out("/FunctionType 2 /Domain [0 1] /N 1>>]")
		f.out("endobj")
		v.objID = f.n
		f.spotColorMap[k] = v
	}
}

func (f *TinyPDF) spotColorPutResourceDict() {
	f.out("/ColorSpace <<")
	for _, clr := range f.spotColorMap {
		f.outf("/CS%d %d 0 R", clr.id, clr.objID)
	}
	f.out(">>")
}
