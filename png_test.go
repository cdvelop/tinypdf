package tinypdf_test

import (
	"bytes"
	"os"
	"testing"
)

func BenchmarkParsePNG_rgb(b *testing.B) {
	raw, err := os.ReadFile("image/golang-gopher.png")
	if err != nil {
		b.Fatal(err)
	}

	pdf := NewDocPdfTest()
	pdf.AddPage()

	const readDPI = true
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pdf.ParsePNG(bytes.NewReader(raw), readDPI)
	}
}

func BenchmarkParsePNG_gray(b *testing.B) {
	raw, err := os.ReadFile("image/logo-gray.png")
	if err != nil {
		b.Fatal(err)
	}

	pdf := NewDocPdfTest()
	pdf.AddPage()

	const readDPI = true
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pdf.ParsePNG(bytes.NewReader(raw), readDPI)
	}
}

func BenchmarkParsePNG_small(b *testing.B) {
	raw, err := os.ReadFile("image/logo.png")
	if err != nil {
		b.Fatal(err)
	}

	pdf := NewDocPdfTest()
	pdf.AddPage()

	const readDPI = true
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pdf.ParsePNG(bytes.NewReader(raw), readDPI)
	}
}

func BenchmarkParseJPG(b *testing.B) {
	raw, err := os.ReadFile("image/logo_gofpdf.jpg")
	if err != nil {
		b.Fatal(err)
	}

	pdf := NewDocPdfTest()
	pdf.AddPage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pdf.ParseJPG(bytes.NewReader(raw))
	}
}

func BenchmarkParseGIF(b *testing.B) {
	raw, err := os.ReadFile("image/logo.gif")
	if err != nil {
		b.Fatal(err)
	}

	pdf := NewDocPdfTest()
	pdf.AddPage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pdf.ParseGIF(bytes.NewReader(raw))
	}
}
