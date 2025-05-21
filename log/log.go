package log

import "github.com/cdvelop/docpdf/env"

// función println interna para imprimir en consola
func Print(args ...any) {
	env.Logger(args...)
}
