//go:build !wasm

package fontManager

import (
	"os"
	"path/filepath"

	. "github.com/cdvelop/tinystring"
)

// getFontList returns the list of font filenames from the base path (non-Wasm).
func (fm *FontManager) getFontList() ([]string, error) {
	files, err := os.ReadDir(fm.fontsBasePath)
	if err != nil {
		return nil, Errf("could not read font directory '%s': %w", fm.fontsBasePath, err)
	}

	var list []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		list = append(list, f.Name())
	}
	return list, nil
}

// getFontData reads the font file bytes for the given filename (non-Wasm).
func (fm *FontManager) getFontData(name string) ([]byte, error) {
	fullPath := filepath.Join(fm.fontsBasePath, name)
	return os.ReadFile(fullPath)
}
