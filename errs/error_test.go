package errs

import (
	"testing"
)

func TestErrAllTypes(t *testing.T) {
	// Llamada al método newErr con varios tipos
	e := New(
		"stringTest",
		[]string{"array", "of", "strings"},
		rune(':'), // Solo se une sin espacio adicional
		42,
		3.14,
		true,
		New("customError"),
	)

	expected := "stringTest array of strings: 42 3.14 true customError"

	if e.Error() != expected {
		t.Errorf("se obtuvo: %q, se esperaba: %q", e.Error(), expected)
	}
}
