package tinypdf

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/cdvelop/tinypdf/fontManager"
	. "github.com/cdvelop/tinystring"
)

func baseNoExt(fileStr string) string {
	str := filepath.Base(fileStr)
	extLen := len(filepath.Ext(str))
	if extLen > 0 {
		str = str[:len(str)-extLen]
	}
	return str
}

func loadMap(encodingFileStr string) (encList encListType, err error) {
	// printf("Encoding file string [%s]\n", encodingFileStr)
	var f *os.File
	// f, err = os.Open(encodingFilepath(encodingFileStr))
	f, err = os.Open(encodingFileStr)
	if err == nil {
		defer f.Close()
		for j := range encList {
			encList[j].uv = -1
			encList[j].name = ".notdef"
		}
		scanner := bufio.NewScanner(f)
		var enc encType
		var pos int
		for scanner.Scan() {
			// "!3F U+003F question"
			_, err = Sscanf(scanner.Text(), "!%x U+%x %s", &pos, &enc.uv, &enc.name)
			if err == nil {
				if pos < 256 {
					encList[pos] = enc
				} else {
					err = Errf("map position 0x%2X exceeds 0xFF", pos)
					return
				}
			} else {
				return
			}
		}
		if err = scanner.Err(); err != nil {
			return
		}
	}
	return
}

type segmentType struct {
	marker uint8
	tp     uint8
	size   uint32
	data   []byte
}

func segmentRead(r io.Reader) (s segmentType, err error) {
	if err = binary.Read(r, binary.LittleEndian, &s.marker); err != nil {
		return
	}
	if s.marker != 128 {
		err = Errf("font file is not a valid binary Type1")
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &s.tp); err != nil {
		return
	}
	if err = binary.Read(r, binary.LittleEndian, &s.size); err != nil {
		return
	}
	s.data = make([]byte, s.size)
	_, err = r.Read(s.data)
	return
}

// -rw-r--r-- 1 root root  9532 2010-04-22 11:27 /usr/share/fonts/type1/mathml/Symbol.afm
// -rw-r--r-- 1 root root 37744 2010-04-22 11:27 /usr/share/fonts/type1/mathml/Symbol.pfb

// getInfoFromType1 return information from a Type1 font
func getInfoFromType1(fileStr string, log func(...any), embed bool, encList encListType) (info fontInfoType, err error) {
	info.Widths = make([]int, 256)
	if embed {
		var f *os.File
		f, err = os.Open(fileStr)
		if err != nil {
			return
		}
		defer f.Close()
		// Read first segment
		var s1, s2 segmentType
		s1, err = segmentRead(f)
		if err != nil {
			return
		}
		s2, err = segmentRead(f)
		if err != nil {
			return
		}
		info.Data = s1.data
		info.Data = append(info.Data, s2.data...)
		info.Size1 = s1.size
		info.Size2 = s2.size
	}
	afmFileStr := fileStr[0:len(fileStr)-3] + "afm"
	size, ok := fileSize(afmFileStr)
	if !ok {
		err = Errf("font file (ATM) %s not found", afmFileStr)
		return
	} else if size == 0 {
		err = Errf("font file (AFM) %s empty or not readable", afmFileStr)
		return
	}

	f, err := os.Open(afmFileStr)
	if err != nil {
		return
	}
	defer f.Close()

	p := newAFMParser(f)
	err = p.parse(&info)
	if err != nil {
		return info, err
	}

	if info.FontName == "" {
		err = Errf("the field FontName missing in AFM file %s", afmFileStr)
		return
	}
	var (
		missingWd int
		wdMap     = p.wdmap
	)
	missingWd, ok = wdMap[".notdef"]
	if ok {
		info.Desc.MissingWidth = missingWd
	}
	for j := 0; j < len(info.Widths); j++ {
		info.Widths[j] = info.Desc.MissingWidth
	}
	for j := 0; j < len(info.Widths); j++ {
		name := encList[j].name
		if name != ".notdef" {
			wd, ok := wdMap[name]
			if ok {
				info.Widths[j] = wd
			} else {
				if log != nil {
					log(Fmt("Character %s is missing", name))
				}
			}
		}
	}
	// printf("getInfoFromType1/FontBBox\n")
	// dump(info.Desc.FontBBox)
	return
}

func makeFontDescriptor(info *fontInfoType) {
	if info.Desc.CapHeight == 0 {
		info.Desc.CapHeight = info.Desc.Ascent
	}
	info.Desc.Flags = 1 << 5
	if info.IsFixedPitch {
		info.Desc.Flags |= 1
	}
	if info.Desc.ItalicAngle != 0 {
		info.Desc.Flags |= 1 << 6
	}
	if info.Desc.StemV == 0 {
		if info.Bold {
			info.Desc.StemV = 120
		} else {
			info.Desc.StemV = 70
		}
	}
	// printf("makeFontDescriptor/FontBBox\n")
	// dump(info.Desc.FontBBox)
}

// makeFontEncoding builds differences from reference encoding
func makeFontEncoding(encList encListType, refEncFileStr string) (diffStr string, err error) {
	var refList encListType
	if refList, err = loadMap(refEncFileStr); err != nil {
		return
	}
	var buf fmtBuffer
	last := 0
	for j := 32; j < 256; j++ {
		if encList[j].name != refList[j].name {
			if j != last+1 {
				buf.printf("%d ", j)
			}
			last = j
			buf.printf("/%s ", encList[j].name)
		}
	}
	diffStr = Convert(buf.String()).TrimSpace().String()
	return
}

func makeDefinitionFile(fileStr, tpStr, encodingFileStr string, embed bool, encList encListType, info fontInfoType) error {
	var err error
	var def fontManager.FontDefType
	def.Tp = tpStr
	def.Name = info.FontName
	makeFontDescriptor(&info)
	def.Desc = info.Desc
	// printf("makeDefinitionFile/FontBBox\n")
	// dump(def.Desc.FontBBox)
	def.Up = info.UnderlinePosition
	def.Ut = info.UnderlineThickness
	def.Cw = info.Widths
	def.Enc = baseNoExt(encodingFileStr)
	// fmt.Printf("encodingFileStr [%s], def.Enc [%s]\n", encodingFileStr, def.Enc)
	// fmt.Printf("reference [%s]\n", filepath.Join(filepath.Dir(encodingFileStr), "cp1252.map"))
	def.Diff, err = makeFontEncoding(encList, filepath.Join(filepath.Dir(encodingFileStr), "cp1252.map"))
	if err != nil {
		return err
	}
	def.File = info.File
	def.Size1 = int(info.Size1)
	def.Size2 = int(info.Size2)
	def.OriginalSize = info.OriginalSize
	// printf("Font definition file [%s]\n", fileStr)
	var buf []byte
	buf, err = json.Marshal(def)
	if err != nil {
		return err
	}
	var f *os.File
	f, err = os.Create(fileStr)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(buf)
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil
}
