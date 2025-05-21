package fmt

import (
	"github.com/cdvelop/docpdf/log"
)

// Print escribe todos los args separados por espacios y añade '\n'
func Print(args ...any) {
	log.Print(args...)
}
