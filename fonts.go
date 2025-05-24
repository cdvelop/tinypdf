package docpdf

import (
	"io"
	"os"
	"path"
	"strings"
)

// AddFont imports a TrueType, OpenType or Type1 font and makes it available.
// It is necessary to generate a font definition file first with the makefont
// utility. It is not necessary to call this function for the core PDF fonts
// (courier, helvetica, times, zapfdingbats).
//
// The JSON definition file (and the font file itself when embedding) must be
// present in the font directory. If it is not found, the error "Could not
// include font definition file" is set.
//
// family specifies the font family. The name can be chosen arbitrarily. If it
// is a standard family name, it will override the corresponding font. This
// string is used to subsequently set the font with the SetFont method.
//
// style specifies the font style. Acceptable values are (case insensitive) the
// empty string for regular style, "B" for bold, "I" for italic, or "BI" or
// "IB" for bold and italic combined.
//
// fileStr specifies the base name with ".json" extension of the font
// definition file to be added. The file will be loaded from the font directory
// specified in the call to New() or SetFontLocation().
func (f *DocPDF) AddFont(familyStr, styleStr, fileStr string) {
	f.addFont(fontFamilyEscape(familyStr), styleStr, fileStr, false)
}

// AddUTF8Font imports a TrueType font with utf-8 symbols and makes it available.
// It is necessary to generate a font definition file first with the makefont
// utility. It is not necessary to call this function for the core PDF fonts
// (courier, helvetica, times, zapfdingbats).
//
// The JSON definition file (and the font file itself when embedding) must be
// present in the font directory. If it is not found, the error "Could not
// include font definition file" is set.
//
// family specifies the font family. The name can be chosen arbitrarily. If it
// is a standard family name, it will override the corresponding font. This
// string is used to subsequently set the font with the SetFont method.
//
// style specifies the font style. Acceptable values are (case insensitive) the
// empty string for regular style, "B" for bold, "I" for italic, or "BI" or
// "IB" for bold and italic combined.
//
// fileStr specifies the base name with ".json" extension of the font
// definition file to be added. The file will be loaded from the font directory
// specified in the call to New() or SetFontLocation().
func (f *DocPDF) AddUTF8Font(familyStr, styleStr, fileStr string) {
	f.addFont(fontFamilyEscape(familyStr), styleStr, fileStr, true)
}

func (f *DocPDF) addFont(familyStr, styleStr, fileStr string, isUTF8 bool) {
	if fileStr == "" {
		if isUTF8 {
			fileStr = strings.Replace(familyStr, " ", "", -1) + strings.ToLower(styleStr) + ".ttf"
		} else {
			fileStr = strings.Replace(familyStr, " ", "", -1) + strings.ToLower(styleStr) + ".json"
		}
	}
	if isUTF8 {
		fontKey := getFontKey(familyStr, styleStr)
		_, ok := f.fonts[fontKey]
		if ok {
			return
		}
		var ttfStat os.FileInfo
		var err error
		fileStr = path.Join(f.fontsPath, fileStr)
		ttfStat, err = os.Stat(fileStr)
		if err != nil {
			f.SetError(err)
			return
		}
		originalSize := ttfStat.Size()
		Type := "UTF8"
		var utf8Bytes []byte
		utf8Bytes, err = os.ReadFile(fileStr)
		if err != nil {
			f.SetError(err)
			return
		}
		reader := fileReader{readerPosition: 0, array: utf8Bytes}
		utf8File := newUTF8Font(&reader)
		err = utf8File.parseFile()
		if err != nil {
			f.SetError(err)
			return
		}

		desc := FontDescType{
			Ascent:       int(utf8File.Ascent),
			Descent:      int(utf8File.Descent),
			CapHeight:    utf8File.CapHeight,
			Flags:        utf8File.Flags,
			FontBBox:     utf8File.Bbox,
			ItalicAngle:  utf8File.ItalicAngle,
			StemV:        utf8File.StemV,
			MissingWidth: round(utf8File.DefaultWidth),
		}

		var sbarr map[int]int
		if f.aliasNbPagesStr == "" {
			sbarr = makeSubsetRange(57)
		} else {
			sbarr = makeSubsetRange(32)
		}
		def := fontDefType{
			Tp:        Type,
			Name:      fontKey,
			Desc:      desc,
			Up:        int(round(utf8File.UnderlinePosition)),
			Ut:        round(utf8File.UnderlineThickness),
			Cw:        utf8File.CharWidths,
			usedRunes: sbarr,
			File:      fileStr,
			utf8File:  utf8File,
		}
		def.i, _ = generateFontID(def)
		f.fonts[fontKey] = def
		f.fontFiles[fontKey] = fontFileType{
			length1:  originalSize,
			fontType: "UTF8",
		}
		f.fontFiles[fileStr] = fontFileType{
			fontType: "UTF8",
		}
	} else {
		if f.fontLoader != nil {
			reader, err := f.fontLoader.Open(fileStr)
			if err == nil {
				f.AddFontFromReader(familyStr, styleStr, reader)
				if closer, ok := reader.(io.Closer); ok {
					closer.Close()
				}
				return
			}
		}

		fileStr = path.Join(f.fontsPath, fileStr)
		file, err := os.Open(fileStr)
		if err != nil {
			f.err = err
			return
		}
		defer file.Close()

		f.AddFontFromReader(familyStr, styleStr, file)
	}
}

// GetFontLocation returns the location in the file system of the font and font
// definition files.
func (f *DocPDF) GetFontLocation() string {
	return f.fontsPath
}

// SetFontLocation sets the location in the file system of the font and font
// definition files.
func (f *DocPDF) SetFontLocation(fontDirStr string) {
	f.fontsPath = fontDirStr
}

func (f *DocPDF) loadFontFile(name string) ([]byte, error) {
	if f.fontLoader != nil {
		reader, err := f.fontLoader.Open(name)
		if err == nil {
			data, err := io.ReadAll(reader)
			if closer, ok := reader.(io.Closer); ok {
				closer.Close()
			}
			return data, err
		}
	}
	return os.ReadFile(path.Join(f.fontsPath, name))
}
