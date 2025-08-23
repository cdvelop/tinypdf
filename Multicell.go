package tinypdf

import (
	"math"

	. "github.com/cdvelop/tinystring"
)

// MultiCell supports printing text with line breaks. They can be automatic (as
// soon as the text reaches the right border of the cell) or explicit (via the
// \n character). As many cells as necessary are output, one below the other.
//
// Text can be aligned, centered or justified. The cell block can be framed and
// the background painted. See CellFormat() for more details.
//
// The current position after calling MultiCell() is the beginning of the next
// line, equivalent to calling CellFormat with ln equal to 1.
//
// w is the width of the cells. A value of zero indicates cells that reach to
// the right margin.
//
// h indicates the line height of each cell in the Unit of measure specified in New().
//
// Note: this method has a known bug that treats UTF-8 fonts differently than
// non-UTF-8 fonts. With UTF-8 fonts, all trailing newlines in txtStr are
// removed. With a non-UTF-8 font, if txtStr has one or more trailing newlines,
// only the last is removed. In the next major module version, the UTF-8 logic
// will be changed to match the non-UTF-8 logic. To prepare for that change,
// applications that use UTF-8 fonts and depend on having all trailing newlines
// removed should call TrimRight(txtStr, "\r\n") before calling this
// method.
func (f *TinyPDF) MultiCell(w, h float64, txtStr, borderStr, alignStr string, fill bool) {
	if f.err != nil {
		return
	}
	// dbg("MultiCell")
	if alignStr == "" {
		alignStr = "J"
	}
	cw := f.currentFont.Cw
	if w == 0 {
		w = f.w - f.rMargin - f.x
	}
	wmax := int(math.Ceil((w - 2*f.cMargin) * 1000 / f.fontSize))
	s := Convert(txtStr).Replace("\r", "").String()
	srune := []rune(s)

	// remove extra line breaks
	var nb int
	if f.isCurrentUTF8 {
		nb = len(srune)
		for nb > 0 && srune[nb-1] == '\n' {
			nb--
		}
		srune = srune[0:nb]
	} else {
		nb = len(s)
		bytes2 := []byte(s)

		// for nb > 0 && bytes2[nb-1] == '\n' {

		// Prior to August 2019, if s ended with a newline, this code stripped it.
		// After that date, to be compatible with the UTF-8 code above, *all*
		// trailing newlines were removed. Because this regression caused at least
		// one application to break (see issue #333), the original behavior has been
		// reinstated with a caveat included in the documentation.
		if nb > 0 && bytes2[nb-1] == '\n' {
			nb--
		}
		s = s[0:nb]
	}
	// dbg("[%s]\n", s)
	var b, b2 string
	b = "0"
	if len(borderStr) > 0 {
		if borderStr == "1" {
			borderStr = "LTRB"
			b = "LRT"
			b2 = "LR"
		} else {
			b2 = ""
			if Contains(borderStr, "L") {
				b2 += "L"
			}
			if Contains(borderStr, "R") {
				b2 += "R"
			}
			if Contains(borderStr, "T") {
				b = b2 + "T"
			} else {
				b = b2
			}
		}
	}
	sep := -1
	i := 0
	j := 0
	l := 0
	ls := 0
	ns := 0
	nl := 1
	for i < nb {
		// Get next character
		var c rune
		if f.isCurrentUTF8 {
			c = srune[i]
		} else {
			c = rune(s[i])
		}
		if c == '\n' {
			// Explicit line break
			if f.ws > 0 {
				f.ws = 0
				f.out("0 Tw")
			}

			if f.isCurrentUTF8 {
				newAlignStr := alignStr
				if newAlignStr == "J" {
					if f.isRTL {
						newAlignStr = "R"
					} else {
						newAlignStr = "L"
					}
				}
				f.CellFormat(w, h, string(srune[j:i]), b, 2, newAlignStr, fill, 0, "")
			} else {
				f.CellFormat(w, h, s[j:i], b, 2, alignStr, fill, 0, "")
			}
			i++
			sep = -1
			j = i
			l = 0
			ns = 0
			nl++
			if len(borderStr) > 0 && nl == 2 {
				b = b2
			}
			continue
		}
		if c == ' ' || isChinese(c) {
			sep = i
			ls = l
			ns++
		}
		if int(c) >= len(cw) {
			f.err = Errf("character outside the supported range: %s", string(c))
			return
		}
		if cw[int(c)] == 0 { //Marker width 0 used for missing symbols
			l += f.currentFont.Desc.MissingWidth
		} else if cw[int(c)] != 65535 { //Marker width 65535 used for zero width symbols
			l += cw[int(c)]
		}
		if l > wmax {
			// Automatic line break
			if sep == -1 {
				if i == j {
					i++
				}
				if f.ws > 0 {
					f.ws = 0
					f.out("0 Tw")
				}
				if f.isCurrentUTF8 {
					f.CellFormat(w, h, string(srune[j:i]), b, 2, alignStr, fill, 0, "")
				} else {
					f.CellFormat(w, h, s[j:i], b, 2, alignStr, fill, 0, "")
				}
			} else {
				if alignStr == "J" {
					if ns > 1 {
						f.ws = float64((wmax-ls)/1000) * f.fontSize / float64(ns-1)
					} else {
						f.ws = 0
					}
					// f.outf("%.3f Tw", f.ws*f.k)
					f.putF64(f.ws*f.k, 3)
					f.put(" Tw\n")
				}
				if f.isCurrentUTF8 {
					f.CellFormat(w, h, string(srune[j:sep]), b, 2, alignStr, fill, 0, "")
				} else {
					f.CellFormat(w, h, s[j:sep], b, 2, alignStr, fill, 0, "")
				}
				i = sep + 1
			}
			sep = -1
			j = i
			l = 0
			ns = 0
			nl++
			if len(borderStr) > 0 && nl == 2 {
				b = b2
			}
		} else {
			i++
		}
	}
	// Last chunk
	if f.ws > 0 {
		f.ws = 0
		f.out("0 Tw")
	}
	if len(borderStr) > 0 && Contains(borderStr, "B") {
		b += "B"
	}
	if f.isCurrentUTF8 {
		if alignStr == "J" {
			if f.isRTL {
				alignStr = "R"
			} else {
				alignStr = ""
			}
		}
		f.CellFormat(w, h, string(srune[j:i]), b, 2, alignStr, fill, 0, "")
	} else {
		f.CellFormat(w, h, s[j:i], b, 2, alignStr, fill, 0, "")
	}
	f.x = f.lMargin
}
