package fpdf_test

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cdvelop/tinystring"

	tinypdf "github.com/cdvelop/tinypdf/fpdf"
)

var rootTestDir tinypdf.RootDirectoryType

// setRoot assigns the relative path to the rootTestDir directory based on current working
// directory
func init() {
	wdStr, err := os.Getwd()
	if err == nil {
		rootTestDir = tinypdf.RootDirectoryType(wdStr)
	} else {
		panic(err)
	}
}

// default docpdf init for testing
func NewDocPdfTest(options ...any) *tinypdf.Fpdf {

	// add root directory to the options
	options = append(options, rootTestDir)

	// add default writeFile function using os for tests
	options = append(options, tinypdf.WriteFileFunc(func(filePath string, content []byte) error {
		return os.WriteFile(filePath, content, 0644)
	}))

	// add default readFile function using os for tests
	options = append(options, tinypdf.ReadFileFunc(func(filePath string) ([]byte, error) {
		return os.ReadFile(filePath)
	}))

	// add default fileSize function using os for tests
	options = append(options, tinypdf.FileSizeFunc(func(filePath string) (int64, error) {
		info, err := os.Stat(filePath)
		if err != nil {
			return 0, err
		}
		return info.Size(), nil
	}))

	pdf := tinypdf.New(options...)
	pdf.SetCompression(false)
	pdf.SetCatalogSort(true)
	pdf.SetCreationDate(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
	pdf.SetModificationDate(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))

	return pdf
}

// ImageFile returns a qualified filename in which the path to the image
// directory is prepended to the specified filename.
func ImageFile(fileStr string) string {
	return rootTestDir.MakePath("image", fileStr)
}

// FontsDirName returns the name to the font directory test.
func FontsDirName() string {
	return "fonts"
}

// FontsDirPath returns the path to the font directory test.
func FontsDirPath() string {
	return rootTestDir.MakePath(FontsDirName())
}

// FontFile returns a qualified filename in which the path to the font
// directory is prepended to the specified filename.
func FontFile(fileStr string) string {
	return filepath.Join(FontsDirPath(), fileStr)
}

// TextFile returns a qualified filename in which the path to the text
// directory is prepended to the specified filename.
func TextFile(fileStr string) string {
	return rootTestDir.MakePath("text", fileStr)
}

// PdfDir returns the path to the PDF output directory.
func PdfDir() string {
	return rootTestDir.MakePath("pdf")
}

// ICCFile returns a qualified filename in which the path to the ICC file
// directory is prepended to the specified filename.
func ICCFile(fileStr string) string {
	return rootTestDir.MakePath("icc", fileStr)
}

// PdfFile returns a qualified filename in which the path to the PDF output
// directory is prepended to the specified filename.
func PdfFile(fileStr string) string {
	return filepath.Join(PdfDir(), fileStr)
}

// Filename returns a qualified filename in which the example PDF directory
// path is prepended and the suffix ".pdf" is appended to the specified
// filename.
func Filename(baseStr string) string {
	return PdfFile(baseStr + ".pdf")
}

// referenceCompare compares the specified file with the file's reference copy
// located in the 'reference' subdirectory. All bytes of the two files are
// compared except for the value of the /CreationDate field in the PDF. This
// function succeeds if both files are equivalent except for their
// /CreationDate values or if the reference file does not exist.
func referenceCompare(fileStr string) (err error) {
	var refFileStr, refDirStr, dirStr, baseFileStr string
	dirStr, baseFileStr = filepath.Split(fileStr)
	refDirStr = filepath.Join(dirStr, "reference")
	err = os.MkdirAll(refDirStr, 0755)
	if err == nil {
		refFileStr = filepath.Join(refDirStr, baseFileStr)
		err = tinypdf.ComparePDFFiles(fileStr, refFileStr, false)
	}
	return
}

// Summary generates a predictable report for use by test examples. If the
// specified error is nil, the filename delimiters are normalized and the
// filename printed to standard output with a success message. If the specified
// error is not nil, its String() value is printed to standard output.
func Summary(err error, fileStr string) {
	if err == nil {
		// Convert absolute path to relative path for consistent output
		if relPath, relErr := filepath.Rel(string(rootTestDir), fileStr); relErr == nil {
			fileStr = relPath
		}
		fileStr = filepath.ToSlash(fileStr)
		fmt.Printf("Successfully generated %s\n", fileStr)
	} else {
		fmt.Println(err)
	}
}

// SummaryCompare generates a predictable report for use by test examples. If
// the specified error is nil, the generated file is compared with a reference
// copy for byte-for-byte equality. If the files match, then the filename
// delimiters are normalized and the filename printed to standard output with a
// success message. If the files do not match, this condition is reported on
// standard output. If the specified error is not nil, its String() value is
// printed to standard output.
func SummaryCompare(err error, fileStr string) {
	if err == nil {
		err = referenceCompare(fileStr)
	}
	if err == nil {
		// Convert absolute path to relative path for consistent output
		if relPath, relErr := filepath.Rel(string(rootTestDir), fileStr); relErr == nil {
			fileStr = relPath
		}
		fileStr = filepath.ToSlash(fileStr)
		fmt.Printf("Successfully generated %s\n", fileStr)
	} else {
		fmt.Println(err)
	}
}

// ExampleFilename tests the Filename() and Summary() functions.
func ExampleFilename() {
	fileStr := Filename("example")
	Summary(tinystring.Err("printer on fire"), fileStr)
	// Output:
	// printer on fire
}
