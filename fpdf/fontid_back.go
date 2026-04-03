//go:build !wasm

package fpdf

import (
	"crypto/sha1"
	"encoding/json"
	. "github.com/tinywasm/fmt"
)

func generateImageID(info *ImageInfoType) (string, error) {
	sha := sha1.New()
	enc := newIDEncoder(sha)
	enc.bytes(info.data)
	enc.bytes(info.smask)
	enc.i64(int64(info.n))
	enc.f64(info.w)
	enc.f64(info.h)
	enc.str(info.cs)
	enc.bytes(info.pal)
	enc.i64(int64(info.bpc))
	enc.str(info.f)
	enc.str(info.dp)
	for _, v := range info.trns {
		enc.i64(int64(v))
	}
	enc.f64(info.scale)
	enc.f64(info.dpi)
	enc.str(info.i)

	return Sprintf("%x", sha.Sum(nil)), nil
}

// generateFontID generates a font Id from the font definition
func generateFontID(fdt fontDefType) (string, error) {
	// file can be different if generated in different instance
	fdt.File = ""
	b, err := json.Marshal(&fdt)
	return Sprintf("%x", sha1.Sum(b)), err
}
