//go:build !wasm

package fpdf

import (
	"bytes"
	"image"
	"image/gif"
	"image/png"
	"io"
)

// parsegif extracts info from a GIF data (via PNG conversion)
func (f *Fpdf) parsegif(r io.Reader) (info *ImageInfoType) {
	data, err := newRBuffer(r)
	if err != nil {
		f.err = err
		return
	}
	var img image.Image
	img, err = gif.Decode(data)
	if err != nil {
		f.err = err
		return
	}
	pngBuf := new(bytes.Buffer)
	err = png.Encode(pngBuf, img)
	if err != nil {
		f.err = err
		return
	}
	return f.parsepngstream(&rbuffer{p: pngBuf.Bytes()}, false)
}
