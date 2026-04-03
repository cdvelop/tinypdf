//go:build wasm

package fpdf

import (
	. "github.com/tinywasm/fmt"
)

func generateImageID(info *ImageInfoType) (string, error) {
	// Simple deterministic ID for WASM to avoid crypto/sha1
	return Sprintf("img_%d_%d_%d", int(info.w), int(info.h), len(info.data)), nil
}

// generateFontID generates a font Id from the font definition
func generateFontID(fdt fontDefType) (string, error) {
	// Simple deterministic ID for WASM
	return fdt.Tp + "_" + fdt.Name, nil
}
