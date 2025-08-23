package tinypdf

import (
	"sort"

	"github.com/cdvelop/tinypdf/fontManager"
	. "github.com/cdvelop/tinystring"
)

func (f *TinyPDF) fontsPut() {
	if f.err != nil {
		return
	}
	nf := f.CurrentObjectNumber()
	// Local index mapping file key -> file entry (populated below).
	var fileIndex map[string]*fontManager.FontFileType
	for _, diff := range f.fm.AllDiffs() {
		// Encodings
		f.newobj()
		f.outf("<</Type /Encoding /BaseEncoding /WinAnsiEncoding /Differences [%s]>>", diff)
		f.out("endobj")
	}
	{
		// Use font files discovered by the font manager (order is stable).
		// Build a local index so we can assign object numbers and reference
		// files when writing FontDescriptor objects.
		files := f.fm.AllFiles()
		fileIndex = make(map[string]*fontManager.FontFileType)
		for i := range files {
			info := files[i]
			if info.FontType == "UTF8" {
				continue
			}
			// create new object for the file stream
			f.newobj()
			// assign object number into the local copy and store pointer in index
			info.N = f.n
			// store pointer into slice copy so callers can find the N
			fileIndex[info.Key] = &files[i]

			// loader has already embedded/compressed font data into info.Content
			font := info.Content

			compressed := len(info.Key) >= 2 && info.Key[len(info.Key)-2:] == ".z"
			if !compressed && info.Length2 > 0 {
				buf := font[6:info.Length1]
				buf = append(buf, font[6+info.Length1+6:info.Length2]...)
				font = buf
			}
			f.outf("<</Length %d", len(font))
			if compressed {
				f.out("/Filter /FlateDecode")
			}
			f.outf("/Length1 %d", info.Length1)
			if info.Length2 > 0 {
				f.outf("/Length2 %d /Length3 0", info.Length2)
			}
			f.out(">>")
			f.putstream(font)
			f.out("endobj")
		}
	}
	{
		// Iterate fonts stored in TinyPDF (slice)
		var fonts []fontManager.FontDefType
		fonts = f.fonts
		// Optionally sort by key
		if f.catalogSort {
			sort.SliceStable(fonts, func(i, j int) bool { return fonts[i].Key < fonts[j].Key })
		}
		for _, font := range fonts {
			// Font objects
			font.N = f.n + 1
			tp := font.Tp
			name := font.Name
			switch tp {
			case "Core":
				// Core font
				f.newobj()
				f.out("<</Type /Font")
				f.outf("/BaseFont /%s", name)
				f.out("/Subtype /Type1")
				if name != "Symbol" && name != "ZapfDingbats" {
					f.out("/Encoding /WinAnsiEncoding")
				}
				f.out(">>")
				f.out("endobj")
			case "Type1":
				fallthrough
			case "TrueType":
				// Additional Type1 or TrueType/OpenType font
				f.newobj()
				f.out("<</Type /Font")
				f.outf("/BaseFont /%s", name)
				f.outf("/Subtype /%s", tp)
				f.out("/FirstChar 32 /LastChar 255")
				f.outf("/Widths %d 0 R", f.n+1)
				f.outf("/FontDescriptor %d 0 R", f.n+2)
				if font.DiffN > 0 {
					f.outf("/Encoding %d 0 R", nf+font.DiffN)
				} else {
					f.out("/Encoding /WinAnsiEncoding")
				}
				f.out(">>")
				f.out("endobj")
				// Widths
				f.newobj()
				var s fmtBuffer
				s.WriteString("[")
				for j := 32; j < 256; j++ {
					s.printf("%d ", font.Cw[j])
				}
				s.WriteString("]")
				f.out(s.String())
				f.out("endobj")
				// Descriptor
				f.newobj()
				s.Truncate(0)
				s.printf("<</Type /FontDescriptor /FontName /%s ", name)
				s.printf("/Ascent %d ", font.Desc.Ascent)
				s.printf("/Descent %d ", font.Desc.Descent)
				s.printf("/CapHeight %d ", font.Desc.CapHeight)
				s.printf("/Flags %d ", font.Desc.Flags)
				s.printf("/FontBBox [%d %d %d %d] ", font.Desc.FontBBox.Xmin, font.Desc.FontBBox.Ymin,
					font.Desc.FontBBox.Xmax, font.Desc.FontBBox.Ymax)
				s.printf("/ItalicAngle %d ", font.Desc.ItalicAngle)
				s.printf("/StemV %d ", font.Desc.StemV)
				s.printf("/MissingWidth %d ", font.Desc.MissingWidth)
				var suffix string
				if tp != "Type1" {
					suffix = "2"
				}
				// Lookup file object by key using local index populated above
				ffPtr, ok := fileIndex[font.File]
				if !ok {
					// Fall back to asking font manager (non-fast path)
					ff, ok2 := f.fm.FindFileByKey(font.File)
					if !ok2 {
						f.err = Errf("font file not found: %s", font.File)
						return
					}
					s.printf("/FontFile%s %d 0 R>>", suffix, ff.N)
				} else {
					s.printf("/FontFile%s %d 0 R>>", suffix, ffPtr.N)
				}
				f.out(s.String())
				f.out("endobj")
			case "UTF8":
				fontName := "utf8" + font.Name
				UsedRunes := font.UsedRunes
				delete(UsedRunes, 0)
				utf8FontStream := font.Utf8File.GenerateCutFont(UsedRunes)
				utf8FontSize := len(utf8FontStream)
				CodeSignDictionary := font.Utf8File.CodeSymbolDictionary
				delete(CodeSignDictionary, 0)

				f.newobj()
				f.out(Fmt("<</Type /Font\n/Subtype /Type0\n/BaseFont /%s\n/Encoding /Identity-H\n/DescendantFonts [%d 0 R]\n/ToUnicode %d 0 R>>\nendobj", fontName, f.n+1, f.n+2))

				f.newobj()
				f.out("<</Type /Font\n/Subtype /CIDFontType2\n/BaseFont /" + fontName + "\n" +
					"/CIDSystemInfo " + Convert(f.n+2).String() + " 0 R\n/FontDescriptor " + Convert(f.n+3).String() + " 0 R")
				if font.Desc.MissingWidth != 0 {
					f.out("/DW " + Convert(font.Desc.MissingWidth).String())
				}
				f.generateCIDFontMap(&font, font.Utf8File.LastRune)
				f.out("/CIDToGIDMap " + Convert(f.n+4).String() + " 0 R>>")
				f.out("endobj")

				f.newobj()
				toUnicode := f.fm.ToUnicodeMap()
				f.out("<</Length " + Convert(len(toUnicode)).String() + ">>")
				f.putstream([]byte(toUnicode))
				f.out("endobj")

				// CIDInfo
				f.newobj()
				f.out("<</Registry (Adobe)\n/Ordering (UCS)\n/Supplement 0>>")
				f.out("endobj")

				// Font descriptor
				f.newobj()
				var s fmtBuffer
				s.printf("<</Type /FontDescriptor /FontName /%s\n /Ascent %d", fontName, font.Desc.Ascent)
				s.printf(" /Descent %d", font.Desc.Descent)
				s.printf(" /CapHeight %d", font.Desc.CapHeight)
				v := font.Desc.Flags
				v = v | 4
				v = v &^ 32
				s.printf(" /Flags %d", v)
				s.printf("/FontBBox [%d %d %d %d] ", font.Desc.FontBBox.Xmin, font.Desc.FontBBox.Ymin,
					font.Desc.FontBBox.Xmax, font.Desc.FontBBox.Ymax)
				s.printf(" /ItalicAngle %d", font.Desc.ItalicAngle)
				s.printf(" /StemV %d", font.Desc.StemV)
				s.printf(" /MissingWidth %d", font.Desc.MissingWidth)
				s.printf("/FontFile2 %d 0 R", f.n+2)
				s.printf(">>")
				f.out(s.String())
				f.out("endobj")

				// Embed CIDToGIDMap
				cidToGidMap := make([]byte, 256*256*2)

				for cc, glyph := range CodeSignDictionary {
					cidToGidMap[cc*2] = byte(glyph >> 8)
					cidToGidMap[cc*2+1] = byte(glyph & 0xFF)
				}

				mem := xmem.compress(cidToGidMap)
				cidToGidMap = mem.bytes()
				f.newobj()
				f.out("<</Length " + Convert(len(cidToGidMap)).String() + "/Filter /FlateDecode>>")
				f.putstream(cidToGidMap)
				f.out("endobj")
				mem.release()

				//Font file
				mem = xmem.compress(utf8FontStream)
				compressedFontStream := mem.bytes()
				f.newobj()
				f.out("<</Length " + Convert(len(compressedFontStream)).String())
				f.out("/Filter /FlateDecode")
				f.out("/Length1 " + Convert(utf8FontSize).String())
				f.out(">>")
				f.putstream(compressedFontStream)
				f.out("endobj")
				mem.release()
			default:
				f.err = Errf("unsupported font type: %s", tp)
				return
			}
		}
	}
}
