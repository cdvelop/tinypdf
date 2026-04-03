//go:build !wasm

package fpdf

import (
	"os"
)

// SVGBasicFileParse parses a simple scalable vector graphics (SVG) file into a
// basic descriptor. The SVGBasicWrite() example demonstrates this method.
func SVGBasicFileParse(svgFileStr string) (sig SVGBasicType, err error) {
	var buf []byte
	buf, err = os.ReadFile(svgFileStr)
	if err == nil {
		sig, err = SVGBasicParse(buf)
	}
	return
}
