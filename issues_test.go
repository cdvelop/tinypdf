package tinypdf_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/cdvelop/tinypdf"
)

func init() {
	cleanup()
}

func cleanup() {
	_ = filepath.Walk(
		PdfDir(),
		func(path string, info os.FileInfo, err error) (reterr error) {
			if info.Mode().IsRegular() {
				dir, _ := filepath.Split(path)
				if filepath.Base(dir) != "reference" {
					if len(path) > 3 {
						if path[len(path)-4:] == ".pdf" {
							os.Remove(path)
						}
					}
				}
			}
			return
		},
	)
}

var summaryCompare = SummaryCompare

func init() {
	if runtime.GOOS == "windows" {
		summaryCompare = Summary
	}
}

func TestFpdfImplementPdf(t *testing.T) {
	// this will not compile if DocPDF and Tpl
	// do not implement Pdf
	var _ tinypdf.Pdf = (*tinypdf.DocPDF)(nil)
	var _ tinypdf.Pdf = (*tinypdf.Tpl)(nil)
}

// TestPagedTemplate ensures new paged templates work
func TestPagedTemplate(t *testing.T) {
	pdf := NewDocPdfTest()
	tpl := pdf.CreateTemplate(func(t *tinypdf.Tpl) {
		// this will be the second page, as a page is already
		// created by default
		t.AddPage()
		t.AddPage()
		t.AddPage()
	})

	if tpl.NumPages() != 4 {
		t.Fatalf("The template does not have the correct number of pages %d", tpl.NumPages())
	}

	tplPages := tpl.FromPages()
	for x := 0; x < len(tplPages); x++ {
		pdf.AddPage()
		pdf.UseTemplate(tplPages[x])
	}

	// get the last template
	tpl2, err := tpl.FromPage(tpl.NumPages())
	if err != nil {
		t.Fatal(err)
	}

	// the objects should be the exact same, as the
	// template will represent the last page by default
	// therefore no new id should be set, and the object
	// should be the same object
	if fmt.Sprintf("%p", tpl2) != fmt.Sprintf("%p", tpl) {
		t.Fatal("Template no longer respecting initial template object")
	}
}

// TestIssue0116 addresses issue 116 in which library silently fails after
// calling CellFormat when no font has been set.
func TestIssue0116(t *testing.T) {
	pdf := NewDocPdfTest()
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "OK")
	if pdf.Error() != nil {
		t.Fatalf("not expecting error when rendering text")
	}

	pdf = tinypdf.New(tinypdf.MM, "A4", "")
	pdf.AddPage()
	pdf.Cell(40, 10, "Not OK") // Font not set
	if pdf.Error() == nil {
		t.Fatalf("expecting error when rendering text without having set font")
	}
}

// TestIssue0193 addresses issue 193 in which the error io.EOF is incorrectly
// assigned to the FPDF instance error.
func TestIssue0193(t *testing.T) {
	var png []byte
	var pdf *tinypdf.DocPDF
	var err error
	var rdr *bytes.Reader

	png, err = os.ReadFile(ImageFile("sweden.png"))
	if err == nil {
		rdr = bytes.NewReader(png)
		pdf = tinypdf.New(tinypdf.MM, "A4", "")
		pdf.AddPage()
		_ = pdf.RegisterImageOptionsReader("sweden", tinypdf.ImageOptions{ImageType: "png", ReadDpi: true}, rdr)
		err = pdf.Error()
	}
	if err != nil {
		t.Fatalf("issue 193 error: %s", err)
	}

}

// TestIssue0209SplitLinesEqualMultiCell addresses issue 209
// make SplitLines and MultiCell split at the same place
func TestIssue0209SplitLinesEqualMultiCell(t *testing.T) {
	pdf := NewDocPdfTest()
	pdf.AddPage()
	pdf.SetFont("Arial", "", 8)
	// this sentence should not be splited
	str := "Guochin Amandine"
	lines := pdf.SplitLines([]byte(str), 26)
	_, FontSize := pdf.GetFontSize()
	y_start := pdf.GetY()
	pdf.MultiCell(26, FontSize, str, "", "L", false)
	y_end := pdf.GetY()

	if len(lines) != 1 {
		t.Fatalf("expect SplitLines split in one line")
	}
	if int(y_end-y_start) != int(FontSize) {
		t.Fatalf("expect MultiCell split in one line %.2f != %.2f", y_end-y_start, FontSize)
	}

	// this sentence should be splited in two lines
	str = "Guiochini Amandine"
	lines = pdf.SplitLines([]byte(str), 26)
	y_start = pdf.GetY()
	pdf.MultiCell(26, FontSize, str, "", "L", false)
	y_end = pdf.GetY()

	if len(lines) != 2 {
		t.Fatalf("expect SplitLines split in two lines")
	}
	if int(y_end-y_start) != int(FontSize*2) {
		t.Fatalf("expect MultiCell split in two lines %.2f != %.2f", y_end-y_start, FontSize*2)
	}
}

// TestFooterFuncLpi tests to make sure the footer is not call twice and SetFooterFuncLpi can work
// without SetFooterFunc.
func TestFooterFuncLpi(t *testing.T) {
	pdf := NewDocPdfTest()
	var (
		oldFooterFnc  = "oldFooterFnc"
		bothPages     = "bothPages"
		firstPageOnly = "firstPageOnly"
		lastPageOnly  = "lastPageOnly"
	)

	// This set just for testing, only set SetFooterFuncLpi.
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, oldFooterFnc,
			"", 0, "C", false, 0, "")
	})
	pdf.SetFooterFuncLpi(func(lastPage bool) {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 10, bothPages, "", 0, "L", false, 0, "")
		if !lastPage {
			pdf.CellFormat(0, 10, firstPageOnly, "", 0, "C", false, 0, "")
		} else {
			pdf.CellFormat(0, 10, lastPageOnly, "", 0, "C", false, 0, "")
		}
	})
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	for j := 1; j <= 40; j++ {
		pdf.CellFormat(0, 10, fmt.Sprintf("Printing line number %d", j),
			"", 1, "", false, 0, "")
	}
	if pdf.Error() != nil {
		t.Fatalf("not expecting error when rendering text")
	}
	w := &bytes.Buffer{}
	if err := pdf.Output(w); err != nil {
		t.Errorf("unexpected err: %s", err)
	}
	b := w.Bytes()
	if bytes.Contains(b, []byte(oldFooterFnc)) {
		t.Errorf("not expecting %s render on pdf when FooterFncLpi is set", oldFooterFnc)
	}
	got := bytes.Count(b, []byte("bothPages"))
	if got != 2 {
		t.Errorf("footer %s should render on two page got:%d", bothPages, got)
	}
	got = bytes.Count(b, []byte(firstPageOnly))
	if got != 1 {
		t.Errorf("footer %s should render only on first page got: %d", firstPageOnly, got)
	}
	got = bytes.Count(b, []byte(lastPageOnly))
	if got != 1 {
		t.Errorf("footer %s should render only on first page got: %d", lastPageOnly, got)
	}
	f := bytes.Index(b, []byte(firstPageOnly))
	l := bytes.Index(b, []byte(lastPageOnly))
	if f > l {
		t.Errorf("index %d (%s) should less than index %d (%s)", f, firstPageOnly, l, lastPageOnly)
	}
}

func TestIssue0069PanicOnSplitTextWithUnicode(t *testing.T) {
	var str string

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%q make SplitText panic", str)
		}
	}()

	pdf := NewDocPdfTest()
	pdf.AddPage()
	pdf.SetFont("Arial", "", 8)

	testChars := []string{"«", "»", "—"}

	for _, str = range testChars {
		_ = pdf.SplitText(str, 100)
	}

}

func TestSplitTextHandleCharacterNotInFontRange(t *testing.T) {
	var str string

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%q text make SplitText panic", str)
		}
	}()

	pdf := NewDocPdfTest()
	pdf.AddPage()
	pdf.SetFont("Arial", "", 8)

	// Test values in utf8 beyond the ascii range
	// I assuming that if the function can handle values in this range
	// it can handle others since the function basically use the rune codepoint
	// as a index for the font char width and 1_000_000 elements must be
	// enough (hopefully!) for the fonts used in the real world.
	for i := 128; i < 1_000_000; i++ {
		str = string(rune(i))
		_ = pdf.SplitText(str, 100)
	}

}

func TestAFMFontParser(t *testing.T) {
	const embed = true
	err := tinypdf.MakeFont(
		FontFile("cmmi10.pfb"),
		FontFile("cp1252.map"),
		FontsDirName(),
		nil, embed,
	)
	if err != nil {
		t.Fatalf("could not create cmmi10 font: %v", err)
	}

}

func BenchmarkLineTo(b *testing.B) {
	pdf := NewDocPdfTest()
	pdf.AddPage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pdf.LineTo(170, 20)
	}
}

func BenchmarkCurveTo(b *testing.B) {
	pdf := NewDocPdfTest()
	pdf.AddPage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pdf.CurveTo(190, 100, 105, 100)
	}
}
