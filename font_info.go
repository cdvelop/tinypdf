package tinypdf

import "github.com/cdvelop/tinypdf/fontManager"

// fontInfoType holds temporary font information used during parsing/loading.
type fontInfoType struct {
	Data               []byte
	File               string
	OriginalSize       int
	FontName           string
	Bold               bool
	IsFixedPitch       bool
	UnderlineThickness int
	UnderlinePosition  int
	Widths             []int
	Size1, Size2       uint32
	Desc               fontManager.FontDescType
}
