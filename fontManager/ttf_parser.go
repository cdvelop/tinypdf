package fontManager

import (
	"encoding/binary"

	. "github.com/cdvelop/tinystring"
)

// TtfType contains metrics of a TrueType font.
type TtfType struct {
	Embeddable             bool
	UnitsPerEm             uint16
	PostScriptName         string
	Bold                   bool
	ItalicAngle            int16
	IsFixedPitch           bool
	TypoAscender           int16
	TypoDescender          int16
	UnderlinePosition      int16
	UnderlineThickness     int16
	Xmin, Ymin, Xmax, Ymax int16
	CapHeight              int16
	Widths                 []uint16
	Chars                  map[uint16]uint16
}

type ttfParser struct {
	rec              TtfType
	data             []byte
	pos              int
	tables           map[string]uint32
	numberOfHMetrics uint16
	numGlyphs        uint16
}

// TtfParse extracts various metrics from a TrueType font file.
func TtfParse(data []byte) (TtfRec TtfType, err error) {
	var t ttfParser
	t.data = data
	t.pos = 0

	version, err := t.ReadStr(4)
	if err != nil {
		return
	}
	if version == "OTTO" {
		err = Errf("fonts based on PostScript outlines are not supported")
		return
	}
	if version != "\x00\x01\x00\x00" {
		err = Errf("unrecognized file format")
		return
	}
	numTables := int(t.ReadUShort())
	t.Skip(3 * 2) // searchRange, entrySelector, rangeShift
	t.tables = make(map[string]uint32)
	var tag string
	for j := 0; j < numTables; j++ {
		tag, err = t.ReadStr(4)
		if err != nil {
			return
		}
		t.Skip(4) // checkSum
		offset := t.ReadULong()
		t.Skip(4) // length
		t.tables[tag] = offset
	}
	err = t.ParseComponents()
	if err != nil {
		return
	}
	TtfRec = t.rec
	return
}

func (t *ttfParser) ParseComponents() (err error) {
	err = t.ParseHead()
	if err == nil {
		err = t.ParseHhea()
		if err == nil {
			err = t.ParseMaxp()
			if err == nil {
				err = t.ParseHmtx()
				if err == nil {
					err = t.ParseCmap()
					if err == nil {
						err = t.ParseName()
						if err == nil {
							err = t.ParseOS2()
							if err == nil {
								err = t.ParsePost()
							}
						}
					}
				}
			}
		}
	}
	return
}

func (t *ttfParser) ParseHead() (err error) {
	err = t.Seek("head")
	t.Skip(3 * 4) // version, fontRevision, checkSumAdjustment
	magicNumber := t.ReadULong()
	if magicNumber != 0x5F0F3CF5 {
		err = Errf("incorrect magic number")
		return
	}
	t.Skip(2) // flags
	t.rec.UnitsPerEm = t.ReadUShort()
	t.Skip(2 * 8) // created, modified
	t.rec.Xmin = t.ReadShort()
	t.rec.Ymin = t.ReadShort()
	t.rec.Xmax = t.ReadShort()
	t.rec.Ymax = t.ReadShort()
	return
}

func (t *ttfParser) ParseHhea() (err error) {
	err = t.Seek("hhea")
	if err == nil {
		t.Skip(4 + 15*2)
		t.numberOfHMetrics = t.ReadUShort()
	}
	return
}

func (t *ttfParser) ParseMaxp() (err error) {
	err = t.Seek("maxp")
	if err == nil {
		t.Skip(4)
		t.numGlyphs = t.ReadUShort()
	}
	return
}

func (t *ttfParser) ParseHmtx() (err error) {
	err = t.Seek("hmtx")
	if err == nil {
		t.rec.Widths = make([]uint16, 0, 8)
		for j := uint16(0); j < t.numberOfHMetrics; j++ {
			t.rec.Widths = append(t.rec.Widths, t.ReadUShort())
			t.Skip(2) // lsb
		}
		if t.numberOfHMetrics < t.numGlyphs {
			lastWidth := t.rec.Widths[t.numberOfHMetrics-1]
			for j := t.numberOfHMetrics; j < t.numGlyphs; j++ {
				t.rec.Widths = append(t.rec.Widths, lastWidth)
			}
		}
	}
	return
}

func (t *ttfParser) ParseCmap() (err error) {
	var offset int64
	if err = t.Seek("cmap"); err != nil {
		return
	}
	t.Skip(2) // version
	numTables := int(t.ReadUShort())
	offset31 := int64(0)
	for j := 0; j < numTables; j++ {
		platformID := t.ReadUShort()
		encodingID := t.ReadUShort()
		offset = int64(t.ReadULong())
		if platformID == 3 && encodingID == 1 {
			offset31 = offset
		}
	}
	if offset31 == 0 {
		err = Errf("no Unicode encoding found")
		return
	}
	startCount := make([]uint16, 0, 8)
	endCount := make([]uint16, 0, 8)
	idDelta := make([]int16, 0, 8)
	idRangeOffset := make([]uint16, 0, 8)
	t.rec.Chars = make(map[uint16]uint16)
	_, err = t.SeekToPos(int64(t.tables["cmap"]) + offset31)
	if err != nil {
		err = Errf("could not seek to cmap table: %w", err)
		return
	}
	format := t.ReadUShort()
	if format != 4 {
		err = Errf("unexpected subtable format: %d", format)
		return
	}
	t.Skip(2 * 2) // length, language
	segCount := int(t.ReadUShort() / 2)
	t.Skip(3 * 2) // searchRange, entrySelector, rangeShift
	for j := 0; j < segCount; j++ {
		endCount = append(endCount, t.ReadUShort())
	}
	t.Skip(2) // reservedPad
	for j := 0; j < segCount; j++ {
		startCount = append(startCount, t.ReadUShort())
	}
	for j := 0; j < segCount; j++ {
		idDelta = append(idDelta, t.ReadShort())
	}
	offset = t.GetPos()
	for j := 0; j < segCount; j++ {
		idRangeOffset = append(idRangeOffset, t.ReadUShort())
	}
	for j := 0; j < segCount; j++ {
		c1 := startCount[j]
		c2 := endCount[j]
		d := idDelta[j]
		ro := idRangeOffset[j]
		if ro > 0 {
			_, err = t.SeekToPos(offset + 2*int64(j) + int64(ro))
			if err != nil {
				return Errf("could not seek to id range offset: %w", err)
			}
		}
		for c := c1; c <= c2; c++ {
			if c == 0xFFFF {
				break
			}
			var gid int32
			if ro > 0 {
				gid = int32(t.ReadUShort())
				if gid > 0 {
					gid += int32(d)
				}
			} else {
				gid = int32(c) + int32(d)
			}
			if gid >= 65536 {
				gid -= 65536
			}
			if gid > 0 {
				t.rec.Chars[c] = uint16(gid)
			}
		}
	}
	return
}

func (t *ttfParser) ParseName() (err error) {
	err = t.Seek("name")
	if err == nil {
		tableOffset := t.GetPos()
		t.rec.PostScriptName = ""
		t.Skip(2) // format
		count := t.ReadUShort()
		stringOffset := t.ReadUShort()
		for j := uint16(0); j < count && t.rec.PostScriptName == ""; j++ {
			t.Skip(3 * 2) // platformID, encodingID, languageID
			nameID := t.ReadUShort()
			length := t.ReadUShort()
			offset := t.ReadUShort()
			if nameID == 6 {
				// PostScript name
				_, err = t.SeekToPos(int64(tableOffset) + int64(stringOffset) + int64(offset))
				if err != nil {
					return
				}
				var s string
				s, err = t.ReadStr(int(length))
				if err != nil {
					return
				}
				s = Convert(s).Replace("\x00", "", -1).String()
				t.rec.PostScriptName = cleanPostScriptName(s)
			}
		}
		if t.rec.PostScriptName == "" {
			err = Errf("the name PostScript was not found")
		}
	}
	return
}

func (t *ttfParser) ParseOS2() (err error) {
	err = t.Seek("OS/2")
	if err == nil {
		version := t.ReadUShort()
		t.Skip(3 * 2) // xAvgCharWidth, usWeightClass, usWidthClass
		fsType := t.ReadUShort()
		t.rec.Embeddable = (fsType != 2) && (fsType&0x200) == 0
		t.Skip(11*2 + 10 + 4*4 + 4)
		fsSelection := t.ReadUShort()
		t.rec.Bold = (fsSelection & 32) != 0
		t.Skip(2 * 2) // usFirstCharIndex, usLastCharIndex
		t.rec.TypoAscender = t.ReadShort()
		t.rec.TypoDescender = t.ReadShort()
		if version >= 2 {
			t.Skip(3*2 + 2*4 + 2)
			t.rec.CapHeight = t.ReadShort()
		} else {
			t.rec.CapHeight = 0
		}
	}
	return
}

func (t *ttfParser) ParsePost() (err error) {
	err = t.Seek("post")
	if err == nil {
		t.Skip(4) // version
		t.rec.ItalicAngle = t.ReadShort()
		t.Skip(2) // Skip decimal part
		t.rec.UnderlinePosition = t.ReadShort()
		t.rec.UnderlineThickness = t.ReadShort()
		t.rec.IsFixedPitch = t.ReadULong() != 0
	}
	return
}

func (t *ttfParser) SeekToPos(pos int64) (int64, error) {
	if pos < 0 || int(pos) >= len(t.data) {
		return 0, Errf("seek position %d out of bounds", pos)
	}
	t.pos = int(pos)
	return pos, nil
}

func (t *ttfParser) GetPos() int64 {
	return int64(t.pos)
}

func (t *ttfParser) Seek(tag string) (err error) {
	ofs, ok := t.tables[tag]
	if !ok {
		return Errf("table not found: %s", tag)
	}

	if int(ofs) >= len(t.data) {
		return Errf("seek position %d out of bounds", ofs)
	}
	t.pos = int(ofs)
	return
}

func (t *ttfParser) Skip(n int) {
	t.pos += n
	if t.pos > len(t.data) {
		panic(Errf("skip position %d out of bounds", t.pos))
	}
}

func (t *ttfParser) ReadStr(length int) (str string, err error) {
	if t.pos+length > len(t.data) {
		return "", Errf("unable to read %d bytes at position %d", length, t.pos)
	}
	str = string(t.data[t.pos : t.pos+length])
	t.pos += length
	return
}

func (t *ttfParser) ReadUShort() (val uint16) {
	if t.pos+2 > len(t.data) {
		panic(Errf("cannot read u16 at position %d", t.pos))
	}
	val = binary.BigEndian.Uint16(t.data[t.pos:])
	t.pos += 2
	return
}

func (t *ttfParser) ReadShort() (val int16) {
	if t.pos+2 > len(t.data) {
		panic(Errf("cannot read i16 at position %d", t.pos))
	}
	val = int16(binary.BigEndian.Uint16(t.data[t.pos:]))
	t.pos += 2
	return
}

func (t *ttfParser) ReadULong() (val uint32) {
	if t.pos+4 > len(t.data) {
		panic(Errf("cannot read u32 at position %d", t.pos))
	}
	val = binary.BigEndian.Uint32(t.data[t.pos:])
	t.pos += 4
	return
}

// cleanPostScriptName removes invalid characters from PostScript font names
// Characters to remove: () {} <> space / % [ ]
func cleanPostScriptName(s string) string {
	var result []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '(', ')', '{', '}', '<', '>', ' ', '/', '%', '[', ']':
			// Skip these characters
		default:
			result = append(result, c)
		}
	}
	return string(result)
}
