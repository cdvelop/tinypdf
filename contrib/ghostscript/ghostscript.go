package main

// This command demonstrates the use of ghotscript to reduce the size
// of generated PDFs. This is based on a comment made by farkerhaiku:
// https://github.com/phpdave11/gofpdf/issues/57#issuecomment-185843315

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cdvelop/tinypdf"
)

func report(fileStr string, err error) {
	if err == nil {
		var info os.FileInfo
		info, err = os.Stat(fileStr)
		if err == nil {
			fmt.Printf("%s: OK, size %d\n", fileStr, info.Size())
		} else {
			fmt.Printf("%s: bad stat\n", fileStr)
		}
	} else {
		fmt.Printf("%s: %s\n", fileStr, err)
	}
}

func newPdf() (pdf *tinypdf.TinyPDF) {
	// New API: New(fontsPath []string, logger func(...any))
	pdf = tinypdf.New([]string{"../../font"}, nil)
	pdf.SetCompression(false)
	// Note: AddFont was removed in the new API. Fonts are discovered/managed
	// by the font manager configured via tinypdf.New(fontsPath, logger).
	// For TrueType fonts (ttf) only basic support exists; see README for notes.
	pdf.AddPage()
	// The font manager proxies to TinyPDF's SetFont method; call SetFont directly.
	pdf.SetFont("Calligrapher", "", 35)
	pdf.Cell(0, 10, "Enjoy new fonts with FPDF!")
	return
}

func full(name string) {
	report(name, newPdf().OutputFileAndClose(name))
}

func min(name string) {
	cmd := exec.Command("gs", "-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.4",
		"-dPDFSETTINGS=/screen", "-dNOPAUSE", "-dQUIET",
		"-dBATCH", "-sOutputFile="+name, "-")
	inPipe, err := cmd.StdinPipe()
	if err == nil {
		errChan := make(chan error, 1)
		go func() {
			errChan <- cmd.Start()
		}()
		err = newPdf().Output(inPipe)
		if err == nil {
			report(name, <-errChan)
		} else {
			report(name, err)
		}
	} else {
		report(name, err)
	}
}

func main() {
	full("full.pdf")
	min("min.pdf")
}
