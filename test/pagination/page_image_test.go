package pagination

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/cdvelop/tinypdf"
)

func TestPageWithImage(t *testing.T) {
	var err error
	pdf := &tinypdf.GoPdf{}
	pdf.Start(tinypdf.Config{PageSize: *tinypdf.PageSizeA4})
	pdf.SetMargins(0, 20, 0, 10)
	pdf.AddPage()

	var x float64 = 100
	var y float64 = 20
	imgRect := &tinypdf.Rect{
		W: 354 * 72 / 120,
		H: 241 * 72 / 120,
	}
	for i := 0; i < 10; i++ {
		var imgHeight float64 = 241 * 72 / 120
		pdf.SetNewYIfNoOffset(y, imgHeight)
		y = pdf.GetY()
		err = pdf.Image("../res/gopher01.jpg", x, y, imgRect)
		if err != nil {
			log.Fatal(err)
		}
		y += imgHeight
	}

	err = pdf.WritePdf(fmt.Sprintf("page_image-%s.pdf", time.Now().Format("01-02-15-04-05")))
	if err != nil {
		log.Fatalln(err)
	}

}
