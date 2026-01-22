package main

// This command demonstrates the use of ghotscript to reduce the size
// of generated PDFs. This is based on a comment made by farkerhaiku:
// https://github.com/phpdave11/gofpdf/issues/57#issuecomment-185843315

import (
	"os"
	"os/exec"

	. "github.com/tinywasm/fmt"
	fpdf "github.com/tinywasm/pdf/fpdf"
)

func report(fileStr string, err error) {
	if err == nil {
		var info os.FileInfo
		info, err = os.Stat(fileStr)
		if err == nil {
			Sprintf("%s: OK, size %d\n", fileStr, info.Size())
		} else {
			Sprintf("%s: bad stat\n", fileStr)
		}
	} else {
		Sprintf("%s: %s\n", fileStr, err)
	}
}

func newPdf() (f *fpdf.Fpdf) {
	f = fpdf.New("mm", "A4", "../../font")
	f.SetCompression(false)
	f.AddFont("Calligrapher", "", "calligra.json")
	f.AddPage()
	f.SetFont("Calligrapher", "", 35)
	f.Cell(0, 10, "Enjoy new fonts with FPDF!")
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
