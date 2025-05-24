package docpdf

// Embedded standard fonts

import (
	"embed"
	"io"
	"strings"
)

//go:embed font_embed/*.json font_embed/*.map
var embFS embed.FS

func (f *DocPDF) coreFontReader(familyStr, styleStr string) (r io.ReadCloser) {
	key := familyStr + styleStr
	key = strings.ToLower(key)
	emb, err := embFS.Open("font_embed/" + key + ".json")
	if err == nil {
		r = emb
	} else {
		f.SetErrorf("could not locate \"%s\" among embedded core font definition files", key)
	}
	return
}
