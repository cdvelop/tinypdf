package fontManager

import (
	"crypto/sha1"
	"encoding/json"

	"github.com/cdvelop/tinystring"
	. "github.com/cdvelop/tinystring"
)

// generateFontID generates a unique identifier for a font definition and assigns it.
// This is used to reference the font within the PDF's resource dictionary.
func (fm *FontManager) setFontID(def *FontDef) error {
	// A font's PostScriptName should be unique for a given family and style.
	// We create a hash from it to get a consistent and unique ID.
	if def.Name == "" {
		return Errf("font definition is missing a name")
	}

	h := sha1.New()
	h.Write([]byte(def.Name))
	h.Write([]byte(def.Tp))

	// The hash is used as the font's unique identifier 'i' field.
	def.I = Fmt("%x", h.Sum(nil))
	return nil
}

// generateFontID generates a font Id from the font definition
func generateFontID(fdt FontDefType) (string, error) {
	// file can be different if generated in different instance
	fdt.File = ""
	b, err := json.Marshal(&fdt)
	return tinystring.Fmt("%x", sha1.Sum(b)), err
}
