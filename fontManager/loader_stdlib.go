//go:build !wasm

package fontManager

import (
	"os"
)

// getFontData reads the font file bytes for the given filename (non-Wasm).
func (fm *FontManager) getFontData(path string) ([]byte, error) {
	return os.ReadFile(path)
}
