//go:build wasm

package fpdf

import (
	"github.com/tinywasm/json"
)

func unmarshalFontDef(data []byte, def *fontDefType) error {
	return json.Decode(data, def)
}
