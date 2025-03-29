package tinypdf

import (
	"fmt"
	"io"
)

// IFont represents a font interface.
type IFont interface {
	Init()
	GetType() string
	GetName() string
	GetDesc() []FontDescItem
	GetUp() int
	GetUt() int
	GetCw() FontCw
	GetEnc() string
	GetDiff() string
	GetOriginalsize() int

	SetFamily(family string)
	GetFamily() string
}

// FontCw maps characters to integers.
type FontCw map[byte]int

// FontDescItem is a (key, value) pair.
type FontDescItem struct {
	Key string
	Val string
}

// // Chr
// func Chr(n int) byte {
// 	return byte(n) //ToByte(fmt.Sprintf("%c", n ))
// }

// ToByte returns the first byte of a string.
func ToByte(chr string) byte {
	return []byte(chr)[0]
}

// FontObj font obj
type FontObj struct {
	Family string
	//Style string
	//Size int
	IsEmbedFont bool

	indexObjWidth          int
	indexObjFontDescriptor int
	indexObjEncoding       int

	Font        IFont
	CountOfFont int
}

func (f *FontObj) init(funcGetRoot func() *GoPdf) {
	f.IsEmbedFont = false
	//me.CountOfFont = -1
}

func (f *FontObj) write(w io.Writer, objID int) error {
	baseFont := f.Family
	if f.Font != nil {
		baseFont = f.Font.GetName()
	}

	io.WriteString(w, "<<\n")
	fmt.Fprintf(w, "  /Type /%s\n", f.getType())
	io.WriteString(w, "  /Subtype /TrueType\n")
	fmt.Fprintf(w, "  /BaseFont /%s\n", baseFont)
	if f.IsEmbedFont {
		io.WriteString(w, "  /FirstChar 32 /LastChar 255\n")
		fmt.Fprintf(w, "  /Widths %d 0 R\n", f.indexObjWidth)
		fmt.Fprintf(w, "  /FontDescriptor %d 0 R\n", f.indexObjFontDescriptor)
		fmt.Fprintf(w, "  /Encoding %d 0 R\n", f.indexObjEncoding)
	}
	io.WriteString(w, ">>\n")
	return nil
}

func (f *FontObj) getType() string {
	return "Font"
}

// SetIndexObjWidth sets the width of a font object.
func (f *FontObj) SetIndexObjWidth(index int) {
	f.indexObjWidth = index
}

// SetIndexObjFontDescriptor sets the font descriptor.
func (f *FontObj) SetIndexObjFontDescriptor(index int) {
	f.indexObjFontDescriptor = index
}

// SetIndexObjEncoding sets the encoding.
func (f *FontObj) SetIndexObjEncoding(index int) {
	f.indexObjEncoding = index
}

// SetFontWithStyle : set font style support Regular or Underline
// for Bold|Italic should be loaded appropriate fonts with same styles defined
// size MUST be uint*, int* or float64*
func (gp *GoPdf) SetFontWithStyle(family string, style int, size interface{}) error {
	fontSize, err := convertNumericToFloat64(size)
	if err != nil {
		return err
	}
	found := false
	i := 0
	max := len(gp.pdfObjs)
	for i < max {
		if gp.pdfObjs[i].getType() == subsetFont {
			obj := gp.pdfObjs[i]
			sub, ok := obj.(*SubsetFontObj)
			if ok {
				if sub.GetFamily() == family && sub.GetTtfFontOption().Style == style&^Underline {
					gp.curr.FontSize = fontSize
					gp.curr.FontStyle = style
					gp.curr.FontFontCount = sub.CountOfFont
					gp.curr.FontISubset = sub
					found = true
					break
				}
			}
		}
		i++
	}

	if !found {
		return errMissingFontFamily
	}

	return nil
}

// SetFont : set font style support "" or "U"
// for "B" and "I" should be loaded appropriate fonts with same styles defined
// size MUST be uint*, int* or float64*
func (gp *GoPdf) SetFont(family string, style string, size interface{}) error {
	return gp.SetFontWithStyle(family, getConvertedStyle(style), size)
}

// SetFontSize : set the font size (and only the font size) of the currently
// active font
func (gp *GoPdf) SetFontSize(fontSize float64) error {
	gp.curr.FontSize = fontSize
	return nil
}

// SetCharSpacing : set the character spacing of the currently active font
func (gp *GoPdf) SetCharSpacing(charSpacing float64) error {
	gp.UnitsToPointsVar(&charSpacing)
	gp.curr.CharSpacing = charSpacing
	return nil
}
