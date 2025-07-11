package tinypdf_test

import (
	"bytes"
	"io"
	"sync"
	"testing"

	"github.com/cdvelop/tinypdf"
	"github.com/cdvelop/tinypdf/contrib/gofpdi"
)

func ExampleNewImporter() {
	// create new pdf
	pdf := NewDocPdfTest("pt", "A4")

	// for testing purposes, get an arbitrary template pdf as stream
	rs, _ := getTemplatePdf()

	// create a new Importer instance
	imp := gofpdi.NewImporter()

	// import first page and determine page sizes
	tpl := imp.ImportPageFromStream(pdf, &rs, 1, "/MediaBox")
	pageSizes := imp.GetPageSizes()
	nrPages := len(imp.GetPageSizes())

	// add all pages from template pdf
	for i := 1; i <= nrPages; i++ {
		pdf.AddPage()
		if i > 1 {
			tpl = imp.ImportPageFromStream(pdf, &rs, i, "/MediaBox")
		}
		imp.UseImportedTemplate(pdf, tpl, 0, 0, pageSizes[i]["/MediaBox"]["w"], pageSizes[i]["/MediaBox"]["h"])
	}
	// output
	fileStr := Filename("contrib_gofpdi_Importer")
	err := pdf.OutputFileAndClose(fileStr)
	Summary(err, fileStr)
	// Output:
	// Successfully generated pdf/contrib_gofpdi_Importer.pdf
}

func TestGofpdiConcurrent(t *testing.T) {
	wg := sync.WaitGroup{}
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pdf := NewDocPdfTest()
			pdf.AddPage()
			rs, _ := getTemplatePdf()
			imp := gofpdi.NewImporter()
			tpl := imp.ImportPageFromStream(pdf, &rs, 1, "/MediaBox")
			imp.UseImportedTemplate(pdf, tpl, 0, 0, 210.0, 297.0)
			// write to bytes buffer
			buf := bytes.Buffer{}
			if err := pdf.Output(&buf); err != nil {
				t.Fail()
			}
		}()
	}
	wg.Wait()
}

func getTemplatePdf() (io.ReadSeeker, error) {
	tpdf := tinypdf.New(tinypdf.POINT, "A4", "")
	tpdf.AddPage()
	tpdf.SetFont("Arial", "", 12)
	tpdf.Text(20, 20, "Example Page 1")
	tpdf.AddPage()
	tpdf.Text(20, 20, "Example Page 2")
	tbuf := bytes.Buffer{}
	err := tpdf.Output(&tbuf)
	return bytes.NewReader(tbuf.Bytes()), err
}
