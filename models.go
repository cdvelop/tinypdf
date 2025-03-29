package tinypdf

import (
	"bytes"
	"io"
)

// IObj inteface for all pdf object
type IObj interface {
	init(func() *GoPdf)
	getType() string
	write(w io.Writer, objID int) error
}

// Margins type.
type Margins struct {
	Left, Top, Right, Bottom float64
}

// GoPdf : core library for generating PDF
type GoPdf struct {

	// Page margins
	margins Margins

	pdfObjs []IObj
	config  Config
	anchors map[string]anchorOption

	indexOfCatalogObj int

	/*--- Important obj indexes stored to reduce search loops ---*/
	// Index of pages obj
	indexOfPagesObj int

	// Number of pages obj
	numOfPagesObj int

	// Index of first page obj
	indexOfFirstPageObj int

	// Current position
	curr Current

	indexEncodingObjFonts []int
	indexOfContent        int

	// Index of procset which should be unique
	indexOfProcSet int

	// Buffer for io.Reader compliance
	buf bytes.Buffer

	// PDF protection
	pdfProtection   *PDFProtection
	encryptionObjID int

	// Content streams only
	compressLevel int

	// Document info
	isUseInfo bool
	info      *PdfInfo

	// Outlines/bookmarks
	outlines           *OutlinesObj
	indexOfOutlinesObj int

	// Header and footer functions
	headerFunc func()
	footerFunc func()

	// gofpdi free pdf document importer
	fpdi *importer

	// Placeholder text
	placeHolderTexts map[string]([]placeHolderTextInfo)

	// Log function for debugging
	log func(...any)
}

// Current current state
type Current struct {
	setXCount int //many times we go func SetX()
	X         float64
	Y         float64

	//font
	IndexOfFontObj int
	CountOfFont    int
	CountOfL       int

	FontSize      float64
	FontStyle     int // Regular|Bold|Italic|Underline
	FontFontCount int
	FontType      int // CURRENT_FONT_TYPE_IFONT or  CURRENT_FONT_TYPE_SUBSET

	CharSpacing float64

	FontISubset *SubsetFontObj // FontType == CURRENT_FONT_TYPE_SUBSET

	//page
	IndexOfPageObj int

	//img
	CountOfImg int
	//cache of image in pdf file
	ImgCaches map[int]ImageCache

	//text color mode
	txtColorMode string //color, gray

	//text color
	txtColor ICacheColorText

	//text grayscale
	grayFill float64
	//draw grayscale
	grayStroke float64

	lineWidth float64

	//current page size
	pageSize *Rect

	//current trim box
	trimBox *Box

	sMasksMap       SMaskMap
	extGStatesMap   ExtGStatesMap
	transparency    *Transparency
	transparencyMap TransparencyMap
}

// Box represents a rectangular area with explicit coordinates for all four sides.
// It is used for defining boundaries in PDF documents, such as margins, trim boxes, etc.
// The coordinates are stored in the current unit system (points by default, but can be mm, cm, inches, or pixels).
type Box struct {
	Left, Top, Right, Bottom float64
	unitOverride             defaultUnitConfig
}

// Rect defines a rectangle by its width and height.
// This is used for defining page sizes, content areas, and other rectangular regions in PDF documents.
// The dimensions are stored in the current unit system (points by default, but can be mm, cm, inches, or pixels).
type Rect struct {
	W            float64 // Width of the rectangle
	H            float64 // Height of the rectangle
	unitOverride defaultUnitConfig
}

// defaultUnitConfig is the standard implementation of the unitConfigurator interface.
// It stores the unit type and an optional custom conversion factor.
type defaultUnitConfig struct {
	// Unit specifies the unit type (UnitPT, UnitMM, UnitCM, UnitIN, UnitPX)
	Unit int

	// ConversionForUnit is an optional custom conversion factor
	ConversionForUnit float64
}
