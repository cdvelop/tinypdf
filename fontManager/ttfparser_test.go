package fontManager_test

import (
	_ "embed"
	"fmt"

	"github.com/cdvelop/tinypdf/fontManager"
)

//go:embed fonts/calligra.ttf
var testFontData []byte

func ExampleTtfParse() {
	ttf, err := fontManager.TtfParse(testFontData)
	if err == nil {
		fmt.Printf("Postscript name:  %s\n", ttf.PostScriptName)
		fmt.Printf("unitsPerEm:       %8d\n", ttf.UnitsPerEm)
		fmt.Printf("Xmin:             %8d\n", ttf.Xmin)
		fmt.Printf("Ymin:             %8d\n", ttf.Ymin)
		fmt.Printf("Xmax:             %8d\n", ttf.Xmax)
		fmt.Printf("Ymax:             %8d\n", ttf.Ymax)
	} else {
		fmt.Printf("%s\n", err)
	}
	// Output:
	// Postscript name:  CalligrapherRegular
	// unitsPerEm:           1000
	// Xmin:                 -173
	// Ymin:                 -234
	// Xmax:                 1328
	// Ymax:                  899
}
