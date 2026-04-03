//go:build wasm

package fpdf

import (
	"io"
)

// parsegif is a stub for WASM that returns an error
func (f *Fpdf) parsegif(r io.Reader) (info *ImageInfoType) {
	f.SetErrorf("GIF images are not supported in WASM")
	return nil
}
