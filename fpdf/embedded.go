package fpdf

// Embedded standard fonts

import (
	"embed"
	"io"

	. "github.com/tinywasm/fmt"
)

//go:embed font_embed/*.json font_embed/*.map
var embFS embed.FS

func (f *Fpdf) coreFontReader(familyStr, styleStr string) (r io.ReadCloser) {
	key := familyStr + styleStr
	key = Convert(key).Low().String()
	emb, err := embFS.Open("font_embed/" + key + ".json")
	if err == nil {
		r = emb
	} else if f.err == nil {
		f.err = Err("core font definition", key, "missing")
	}
	return
}
