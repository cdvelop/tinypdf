//go:build !wasm

package fpdf

import (
	"encoding/json"
)

func unmarshalFontDef(data []byte, def *fontDefType) error {
	return json.Unmarshal(data, def)
}
