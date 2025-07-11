package tinypdf

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"io"
	"math"
	"path"
	"time"

	"github.com/cdvelop/tinystring"
)

// Version of FPDF from which this package is derived
const (
	cnFpdfVersion = "1.7"
)

type blendModeType struct {
	strokeStr, fillStr, modeStr string
	objNum                      int
}

type gradientType struct {
	tp                int // 2: linear, 3: radial
	clr1Str, clr2Str  string
	x1, y1, x2, y2, r float64
	objNum            int
}

type RootDirectoryType string // RootDirectoryType is the root directory of the executable default is "." but test can set it to a different directory

// MakePath joins the root directory with one or more path elements.
// For example, with root "/home/user/docpdf":
//
//	r.MakePath("file.txt") returns "/home/user/docpdf/file.txt"
//	r.MakePath("dir", "file.txt") returns "/home/user/docpdf/dir/file.txt"
func (r RootDirectoryType) MakePath(pathElements ...string) string {
	elements := append([]string{string(r)}, pathElements...)
	return path.Join(elements...)
}

type FontsDirName string // FontsDirName is the name of the font directory default is "fonts"

type orientationType string

const (
	// Portrait represents the portrait orientation.
	Portrait orientationType = "p"

	// Landscape represents the landscape orientation.
	Landscape orientationType = "l"
)

type unit string

const (
	// POINT represents the size unit point
	POINT unit = "pt"
	// MM represents the size unit millimeter
	MM unit = "mm"
	// CM represents the size unit centimeter
	CM unit = "cm"
	// IN represents the size unit inch
	IN unit = "inch"
)

// Standard page sizes in points (1/72 inch)
var (
	// A3 represents DIN/ISO A3 page size
	A3 = PageSize{Wd: 841.89, Ht: 1190.55, AutoHt: false}
	// A4 represents DIN/ISO A4 page size
	A4 = PageSize{Wd: 595.28, Ht: 841.89, AutoHt: false}
	// A5 represents DIN/ISO A5 page size
	A5 = PageSize{Wd: 420.94, Ht: 595.28, AutoHt: false}
	// A6 represents DIN/ISO A6 page size
	A6 = PageSize{Wd: 297.64, Ht: 420.94, AutoHt: false}
	// A7 represents DIN/ISO A7 page size
	A7 = PageSize{Wd: 209.76, Ht: 297.64, AutoHt: false}
	// A2 represents DIN/ISO A2 page size
	A2 = PageSize{Wd: 1190.55, Ht: 1683.78, AutoHt: false}
	// A1 represents DIN/ISO A1 page size
	A1 = PageSize{Wd: 1683.78, Ht: 2383.94, AutoHt: false}
	// Letter represents US Letter page size
	Letter = PageSize{Wd: 612, Ht: 792, AutoHt: false}
	// Legal represents US Legal page size
	Legal = PageSize{Wd: 612, Ht: 1008, AutoHt: false}
	// Tabloid represents US Tabloid page size
	Tabloid = PageSize{Wd: 792, Ht: 1224, AutoHt: false}
)

const (
	// BorderNone set no border
	BorderNone = ""
	// BorderFull sets a full border
	BorderFull = "1"
	// BorderLeft sets the border on the left side
	BorderLeft = "L"
	// BorderTop sets the border at the top
	BorderTop = "T"
	// BorderRight sets the border on the right side
	BorderRight = "R"
	// BorderBottom sets the border on the bottom
	BorderBottom = "B"
)

const (
	// LineBreakNone disables linebreak
	LineBreakNone = 0
	// LineBreakNormal enables normal linebreak
	LineBreakNormal = 1
	// LineBreakBelow enables linebreak below
	LineBreakBelow = 2
)

const (
	// AlignLeft left aligns the cell
	AlignLeft = "L"
	// AlignRight right aligns the cell
	AlignRight = "R"
	// AlignCenter centers the cell
	AlignCenter = "C"
	// AlignTop aligns the cell to the top
	AlignTop = "T"
	// AlignBottom aligns the cell to the bottom
	AlignBottom = "B"
	// AlignMiddle aligns the cell to the middle
	AlignMiddle = "M"
	// AlignBaseline aligns the cell to the baseline
	AlignBaseline = "B"
)

type colorMode int

const (
	colorModeRGB colorMode = iota
	colorModeSpot
)

type colorType struct {
	r, g, b    float64
	ir, ig, ib int
	mode       colorMode
	spotStr    string // name of current spot color
	gray       bool
	str        string
}

// SpotColorType specifies a named spot color value
type spotColorType struct {
	id, objID int
	val       cmykColorType
}

// cmykColorType specifies an ink-based CMYK color value
type cmykColorType struct {
	c, m, y, k byte // 0% to 100%
}

// SizeType fields Wd and Ht specify the horizontal and vertical extents of a
// document element such as a page.
type SizeType struct {
	Wd, Ht float64
}

// PageSize specifies the dimensions and properties of a page.
// Wd and Ht specify the horizontal and vertical extents in points.
// AutoHt indicates if the page height should grow automatically based on content.
type PageSize struct {
	Wd, Ht float64
	AutoHt bool // For cases where page size needs to grow automatically (e.g., thermal printer paper)
}

// ToSizeType converts a PageSize to SizeType for compatibility
func (ps PageSize) ToSizeType() SizeType {
	return SizeType{Wd: ps.Wd, Ht: ps.Ht}
}

// PointType fields X and Y specify the horizontal and vertical coordinates of
// a point, typically used in drawing.
type PointType struct {
	X, Y float64
}

// XY returns the X and Y components of the receiver point.
func (p PointType) XY() (float64, float64) {
	return p.X, p.Y
}

// Extent returns the width and height of the page size
func (ps PageSize) Extent() (wd, ht float64) {
	return ps.Wd, ps.Ht
}

// Width returns the width of the page size
func (ps PageSize) Width() float64 {
	return ps.Wd
}

// Height returns the height of the page size
func (ps PageSize) Height() float64 {
	return ps.Ht
}

// ImageInfoType contains size, color and other information about an image.
// Changes to this structure should be reflected in its GobEncode and GobDecode
// methods.
type ImageInfoType struct {
	data  []byte  // Raw image data
	smask []byte  // Soft Mask, an 8bit per-pixel transparency mask
	n     int     // Image object number
	w     float64 // Width
	h     float64 // Height
	cs    string  // Color space
	pal   []byte  // Image color palette
	bpc   int     // Bits Per Component
	f     string  // Image filter
	dp    string  // DecodeParms
	trns  []int   // Transparency mask
	scale float64 // Document scale factor
	dpi   float64 // Dots-per-inch found from image file (png only)
	i     string  // SHA-1 checksum of the above values.
}

type idEncoder struct {
	w   io.Writer
	buf []byte
	err error
}

func newIDEncoder(w io.Writer) *idEncoder {
	return &idEncoder{
		w:   w,
		buf: make([]byte, 8),
	}
}

func (enc *idEncoder) i64(v int64) {
	if enc.err != nil {
		return
	}
	binary.LittleEndian.PutUint64(enc.buf, uint64(v))
	_, enc.err = enc.w.Write(enc.buf)
}

func (enc *idEncoder) f64(v float64) {
	if enc.err != nil {
		return
	}
	binary.LittleEndian.PutUint64(enc.buf, math.Float64bits(v))
	_, enc.err = enc.w.Write(enc.buf)
}

func (enc *idEncoder) str(v string) {
	if enc.err != nil {
		return
	}
	_, enc.err = enc.w.Write([]byte(v))
}

func (enc *idEncoder) bytes(v []byte) {
	if enc.err != nil {
		return
	}
	_, enc.err = enc.w.Write(v)
}

func generateImageID(info *ImageInfoType) (string, error) {
	sha := sha1.New()
	enc := newIDEncoder(sha)
	enc.bytes(info.data)
	enc.bytes(info.smask)
	enc.i64(int64(info.n))
	enc.f64(info.w)
	enc.f64(info.h)
	enc.str(info.cs)
	enc.bytes(info.pal)
	enc.i64(int64(info.bpc))
	enc.str(info.f)
	enc.str(info.dp)
	for _, v := range info.trns {
		enc.i64(int64(v))
	}
	enc.f64(info.scale)
	enc.f64(info.dpi)
	enc.str(info.i)

	return tinystring.Fmt("%x", sha.Sum(nil)), nil
}

// GobEncode encodes the receiving image to a byte slice.
func (info *ImageInfoType) GobEncode() (buf []byte, err error) {
	fields := []interface{}{info.data, info.smask, info.n, info.w, info.h, info.cs,
		info.pal, info.bpc, info.f, info.dp, info.trns, info.scale, info.dpi}
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	for j := 0; j < len(fields) && err == nil; j++ {
		err = encoder.Encode(fields[j])
	}
	if err == nil {
		buf = w.Bytes()
	}
	return
}

// GobDecode decodes the specified byte buffer (generated by GobEncode) into
// the receiving image.
func (info *ImageInfoType) GobDecode(buf []byte) (err error) {
	fields := []interface{}{&info.data, &info.smask, &info.n, &info.w, &info.h,
		&info.cs, &info.pal, &info.bpc, &info.f, &info.dp, &info.trns, &info.scale, &info.dpi}
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	for j := 0; j < len(fields) && err == nil; j++ {
		err = decoder.Decode(fields[j])
	}

	info.i, err = generateImageID(info)
	return
}

// PointConvert returns the value of pt, expressed in points (1/72 inch), as a
// value expressed in the unit of measure specified in New(). Since font
// management in DocPDF uses points, this method can help with line height
// calculations and other methods that require user units.
func (f *DocPDF) PointConvert(pt float64) (u float64) {
	return pt / f.k
}

// PointToUnitConvert is an alias for PointConvert.
func (f *DocPDF) PointToUnitConvert(pt float64) (u float64) {
	return pt / f.k
}

// UnitToPointConvert returns the value of u, expressed in the unit of measure
// specified in New(), as a value expressed in points (1/72 inch). Since font
// management in DocPDF uses points, this method can help with setting font sizes
// based on the sizes of other non-font page elements.
func (f *DocPDF) UnitToPointConvert(u float64) (pt float64) {
	return u * f.k
}

// Extent returns the width and height of the image in the units of the DocPDF
// object.
func (info *ImageInfoType) Extent() (wd, ht float64) {
	return info.Width(), info.Height()
}

// Width returns the width of the image in the units of the DocPDF object.
func (info *ImageInfoType) Width() float64 {
	return info.w / (info.scale * info.dpi / 72)
}

// Height returns the height of the image in the units of the DocPDF object.
func (info *ImageInfoType) Height() float64 {
	return info.h / (info.scale * info.dpi / 72)
}

// SetDpi sets the dots per inch for an image. PNG images MAY have their dpi
// set automatically, if the image specifies it. DPI information is not
// currently available automatically for JPG and GIF images, so if it's
// important to you, you can set it here. It defaults to 72 dpi.
func (info *ImageInfoType) SetDpi(dpi float64) {
	info.dpi = dpi
}

type fontFileType struct {
	length1, length2 int64
	n                int
	embedded         bool
	content          []byte
	fontType         string
}

type linkType struct {
	x, y, wd, ht float64
	link         int    // Auto-generated internal link ID or...
	linkStr      string // ...application-provided external link string
}

type intLinkType struct {
	page int
	y    float64
}

// outlineType is used for a sidebar outline of bookmarks
type outlineType struct {
	text                                   string
	level, parent, first, last, next, prev int
	y                                      float64
	p                                      int
}

// InitType is used with NewCustom() to customize an DocPDF instance.
// OrientationStr, UnitType correspond to the arguments accepted by New().
// If the Wd and Ht fields of Size are each greater than zero, Size will be used
// to set the default page size. Wd and Ht are specified in the units of measure
// indicated by UnitType.
type InitType struct {
	OrientationStr orientationType // Landscape or Portrait
	UnitType       unit
	Size           PageSize
	RootDirectory  RootDirectoryType // Root directory of the executable default is "." but test can set it to a different directory
	FontDirName    string            // name to the font directory default is "fonts"
}

// FontLoader is used to read fonts (JSON font specification and zlib compressed font binaries)
// from arbitrary locations (e.g. files, zip files, embedded font resources).
//
// Open provides an io.Reader for the specified font file (.json or .z). The file name
// never includes a path. Open returns an error if the specified file cannot be opened.
type FontLoader interface {
	Open(name string) (io.Reader, error)
}

// OutputIntentSubtype any of the pre defined types below or a value defined by ISO 32000 extension.
type OutputIntentSubtype string

const (
	OutputIntent_GTS_PDFX  OutputIntentSubtype = "GTS_PDFX"
	OutputIntent_GTS_PDFA1 OutputIntentSubtype = "GTS_PDFA1"
	OutputIntent_GTS_PDFE1 OutputIntentSubtype = "GTS_PDFE1"
)

// OutputIntentType defines an output intent with name and ICC color profile.
type OutputIntentType struct {
	SubtypeIdent              OutputIntentSubtype
	OutputConditionIdentifier string
	Info                      string
	ICCProfile                []byte
}

// Pdf defines the interface used for various methods. It is implemented by the
// main FPDF instance as well as templates.
type Pdf interface {
	AddFont(familyStr, styleStr, fileStr string)
	AddFontFromBytes(familyStr, styleStr string, jsonFileBytes, zFileBytes []byte)
	AddFontFromReader(familyStr, styleStr string, r io.Reader)
	AddLayer(name string, visible bool) (layerID int)
	AddLink() int
	AddPage()
	AddPageFormat(orientationStr orientationType, size PageSize)
	AddSpotColor(nameStr string, c, m, y, k byte)
	AliasNbPages(aliasStr string)
	ArcTo(x, y, rx, ry, degRotate, degStart, degEnd float64)
	Arc(x, y, rx, ry, degRotate, degStart, degEnd float64, styleStr string)
	BeginLayer(id int)
	Beziergon(points []PointType, styleStr string)
	Bookmark(txtStr string, level int, y float64)
	CellFormat(w, h float64, txtStr, borderStr string, ln int, alignStr string, fill bool, link int, linkStr string)
	Cellf(w, h float64, fmtStr string, args ...interface{})
	Cell(w, h float64, txtStr string)
	Circle(x, y, r float64, styleStr string)
	ClearError()
	ClipCircle(x, y, r float64, outline bool)
	ClipEllipse(x, y, rx, ry float64, outline bool)
	ClipEnd()
	ClipPolygon(points []PointType, outline bool)
	ClipRect(x, y, w, h float64, outline bool)
	ClipRoundedRect(x, y, w, h, r float64, outline bool)
	ClipText(x, y float64, txtStr string, outline bool)
	Close()
	ClosePath()
	CreateTemplateCustom(corner PointType, size SizeType, fn func(*Tpl)) Template
	CreateTemplate(fn func(*Tpl)) Template
	CurveBezierCubicTo(cx0, cy0, cx1, cy1, x, y float64)
	CurveBezierCubic(x0, y0, cx0, cy0, cx1, cy1, x1, y1 float64, styleStr string)
	CurveCubic(x0, y0, cx0, cy0, x1, y1, cx1, cy1 float64, styleStr string)
	CurveTo(cx, cy, x, y float64)
	Curve(x0, y0, cx, cy, x1, y1 float64, styleStr string)
	DrawPath(styleStr string)
	Ellipse(x, y, rx, ry, degRotate float64, styleStr string)
	EndLayer()
	Err() bool
	Error() error
	GetAlpha() (alpha float64, blendModeStr string)
	GetAuthor() string
	GetAutoPageBreak() (auto bool, margin float64)
	GetCatalogSort() bool
	GetCellMargin() float64
	GetCompression() bool
	GetConversionRatio() float64
	GetCreationDate() time.Time
	GetCreator() string
	GetDisplayMode() (zoomStr, layoutStr string)
	GetDrawColor() (int, int, int)
	GetDrawSpotColor() (name string, c, m, y, k byte)
	GetFillColor() (int, int, int)
	GetFillSpotColor() (name string, c, m, y, k byte)
	GetFontDesc(familyStr, styleStr string) FontDescType
	GetFontFamily() string
	GetFontLoader() FontLoader
	GetFontLocation() string
	GetFontSize() (ptSize, unitSize float64)
	GetFontStyle() string
	GetImageInfo(imageStr string) (info *ImageInfoType)
	GetJavascript() string
	GetKeywords() string
	GetLang() string
	GetLineCapStyle() string
	GetLineJoinStyle() string
	GetLineWidth() float64
	GetMargins() (left, top, right, bottom float64)
	GetModificationDate() time.Time
	GetPageSize() (width, height float64)
	GetPageSizeStr(sizeStr string) (size PageSize)
	GetProducer() string
	GetStringWidth(s string) float64
	GetSubject() string
	GetTextColor() (int, int, int)
	GetTextSpotColor() (name string, c, m, y, k byte)
	GetTitle() string
	GetUnderlineThickness() float64
	GetWordSpacing() float64
	GetX() float64
	GetXmpMetadata() []byte
	GetXY() (float64, float64)
	GetY() float64
	HTMLBasicNew() (html HTMLBasicType)
	Image(imageNameStr string, x, y, w, h float64, flow bool, tp string, link int, linkStr string)
	ImageOptions(imageNameStr string, x, y, w, h float64, flow bool, options ImageOptions, link int, linkStr string)
	ImageTypeFromMime(mimeStr string) (tp string)
	LinearGradient(x, y, w, h float64, r1, g1, b1, r2, g2, b2 int, x1, y1, x2, y2 float64)
	LineTo(x, y float64)
	Line(x1, y1, x2, y2 float64)
	LinkString(x, y, w, h float64, linkStr string)
	Link(x, y, w, h float64, link int)
	Ln(h float64)
	MoveTo(x, y float64)
	MultiCell(w, h float64, txtStr, borderStr, alignStr string, fill bool)
	Ok() bool
	OpenLayerPane()
	OutputAndClose(w io.WriteCloser) error
	OutputFileAndClose(fileStr string) error
	Output(w io.Writer) error
	PageCount() int
	PageNo() int
	PageSize(pageNum int) (wd, ht float64, unitStr string)
	PointConvert(pt float64) (u float64)
	PointToUnitConvert(pt float64) (u float64)
	Polygon(points []PointType, styleStr string)
	RadialGradient(x, y, w, h float64, r1, g1, b1, r2, g2, b2 int, x1, y1, x2, y2, r float64)
	RawWriteBuf(r io.Reader)
	RawWriteStr(str string)
	Rect(x, y, w, h float64, styleStr string)
	RegisterAlias(alias, replacement string)
	RegisterImage(fileStr, tp string) (info *ImageInfoType)
	RegisterImageOptions(fileStr string, options ImageOptions) (info *ImageInfoType)
	RegisterImageOptionsReader(imgName string, options ImageOptions, r io.Reader) (info *ImageInfoType)
	RegisterImageReader(imgName, tp string, r io.Reader) (info *ImageInfoType)
	SetAcceptPageBreakFunc(fnc func() bool)
	SetAlpha(alpha float64, blendModeStr string)
	SetAuthor(authorStr string, isUTF8 bool)
	SetAutoPageBreak(auto bool, margin float64)
	SetCatalogSort(flag bool)
	SetCellMargin(margin float64)
	SetCompression(compress bool)
	SetCreationDate(tm time.Time)
	SetCreator(creatorStr string, isUTF8 bool)
	SetDashPattern(dashArray []float64, dashPhase float64)
	SetDisplayMode(zoomStr, layoutStr string)
	SetLang(lang string)
	SetDrawColor(r, g, b int)
	SetDrawSpotColor(nameStr string, tint byte)
	SetError(err error)
	SetErrorf(fmtStr string, args ...interface{})
	SetFillColor(r, g, b int)
	SetFillSpotColor(nameStr string, tint byte)
	SetFont(familyStr, styleStr string, size float64)
	SetFontLoader(loader FontLoader)
	SetFontLocation(fontDirStr string)
	SetFontSize(size float64)
	SetFontStyle(styleStr string)
	SetFontUnitSize(size float64)
	SetFooterFunc(fnc func())
	SetFooterFuncLpi(fnc func(lastPage bool))
	SetHeaderFunc(fnc func())
	SetHeaderFuncMode(fnc func(), homeMode bool)
	SetHomeXY()
	SetJavascript(script string)
	SetKeywords(keywordsStr string, isUTF8 bool)
	SetLeftMargin(margin float64)
	SetLineCapStyle(styleStr string)
	SetLineJoinStyle(styleStr string)
	SetLineWidth(width float64)
	SetLink(link int, y float64, page int)
	SetMargins(left, top, right float64)
	SetPageBoxRec(t string, pb PageBox)
	SetPageBox(t string, x, y, wd, ht float64)
	SetPage(pageNum int)
	SetProtection(actionFlag byte, userPassStr, ownerPassStr string)
	SetRightMargin(margin float64)
	SetSubject(subjectStr string, isUTF8 bool)
	SetTextColor(r, g, b int)
	SetTextSpotColor(nameStr string, tint byte)
	SetTitle(titleStr string, isUTF8 bool)
	SetTopMargin(margin float64)
	SetUnderlineThickness(thickness float64)
	SetXmpMetadata(xmpStream []byte)
	SetX(x float64)
	SetXY(x, y float64)
	SetY(y float64)
	SplitLines(txt []byte, w float64) [][]byte
	String() string
	SVGBasicWrite(sb *SVGBasicType, scale float64)
	Text(x, y float64, txtStr string)
	TransformBegin()
	TransformEnd()
	TransformMirrorHorizontal(x float64)
	TransformMirrorLine(angle, x, y float64)
	TransformMirrorPoint(x, y float64)
	TransformMirrorVertical(y float64)
	TransformRotate(angle, x, y float64)
	TransformScale(scaleWd, scaleHt, x, y float64)
	TransformScaleX(scaleWd, x, y float64)
	TransformScaleXY(s, x, y float64)
	TransformScaleY(scaleHt, x, y float64)
	TransformSkew(angleX, angleY, x, y float64)
	TransformSkewX(angleX, x, y float64)
	TransformSkewY(angleY, x, y float64)
	Transform(tm TransformMatrix)
	TransformTranslate(tx, ty float64)
	TransformTranslateX(tx float64)
	TransformTranslateY(ty float64)
	UnicodeTranslatorFromDescriptor(cpStr string) (rep func(string) string)
	UnitToPointConvert(u float64) (pt float64)
	UseTemplateScaled(t Template, corner PointType, size SizeType)
	UseTemplate(t Template)
	WriteAligned(width, lineHeight float64, textStr, alignStr string)
	Writef(h float64, fmtStr string, args ...interface{})
	Write(h float64, txtStr string)
	WriteLinkID(h float64, displayStr string, linkID int)
	WriteLinkString(h float64, displayStr, targetStr string)
}

// PageBox defines the coordinates and extent of the various page box types
type PageBox struct {
	SizeType
	PointType
}

// DocPDF is the principal structure for creating a single PDF document
type DocPDF struct {
	isCurrentUTF8    bool                       // is current font used in utf-8 mode
	isRTL            bool                       // is is right to left mode enabled
	page             int                        // current page number
	n                int                        // current object number
	offsets          []int                      // array of object offsets
	templates        map[string]Template        // templates used in this document
	templateObjects  map[string]int             // template object IDs within this document
	importedObjs     map[string][]byte          // imported template objects (gofpdi)
	importedObjPos   map[string]map[int]string  // imported template objects hashes and their positions (gofpdi)
	importedTplObjs  map[string]string          // imported template names and IDs (hashed) (gofpdi)
	importedTplIDs   map[string]int             // imported template ids hash to object id int (gofpdi)
	buffer           fmtBuffer                  // buffer holding in-memory PDF
	pages            []*bytes.Buffer            // slice[page] of page content; 1-based
	state            int                        // current document state
	compress         bool                       // compression flag
	k                float64                    // scale factor (number of points in user unit)
	defOrientation   orientationType            // default orientation
	curOrientation   orientationType            // current orientation
	stdPageSizes     map[string]PageSize        // standard page sizes
	defPageSize      PageSize                   // default page size
	defPageBoxes     map[string]PageBox         // default page size
	curPageSize      PageSize                   // current page size
	pageSizes        map[int]PageSize           // used for pages with non default sizes or orientations
	pageBoxes        map[int]map[string]PageBox // used to define the crop, trim, bleed and art boxes
	unitType         unit                       // unit of measure for all rendered objects except fonts
	wPt, hPt         float64                    // dimensions of current page in points
	w, h             float64                    // dimensions of current page in user unit
	lMargin          float64                    // left margin
	tMargin          float64                    // top margin
	rMargin          float64                    // right margin
	bMargin          float64                    // page break margin
	cMargin          float64                    // cell margin
	x, y             float64                    // current position in user unit
	lasth            float64                    // height of last printed cell
	lineWidth        float64                    // line width in user unit
	rootDirectory    RootDirectoryType          // root directory of the executable default is "." for test change
	fontsDirName     FontsDirName               // fonts directory name default is "fonts"
	fontsPath        string                     // full path containing fonts directory included rootDirectory eg. "/home/user/docpdf/fonts"
	fontLoader       FontLoader                 // used to load font files from arbitrary locations
	coreFonts        map[string]bool            // array of core font names
	fonts            map[string]fontDefType     // array of used fonts
	fontFiles        map[string]fontFileType    // array of font files
	diffs            []string                   // array of encoding differences
	fontFamily       string                     // current font family
	fontStyle        string                     // current font style
	underline        bool                       // underlining flag
	strikeout        bool                       // strike out flag
	currentFont      fontDefType                // current font info
	fontSizePt       float64                    // current font size in points
	fontSize         float64                    // current font size in user unit
	ws               float64                    // word spacing
	images           map[string]*ImageInfoType  // array of used images
	aliasMap         map[string]string          // map of alias->replacement
	pageLinks        [][]linkType               // pageLinks[page][link], both 1-based
	links            []intLinkType              // array of internal links
	attachments      []Attachment               // slice of content to embed globally
	pageAttachments  [][]annotationAttach       // 1-based array of annotation for file attachments (per page)
	outlines         []outlineType              // array of outlines
	outlineRoot      int                        // root of outlines
	autoPageBreak    bool                       // automatic page breaking
	acceptPageBreak  func() bool                // returns true to accept page break
	pageBreakTrigger float64                    // threshold used to trigger page breaks
	inHeader         bool                       // flag set when processing header
	headerFnc        func()                     // function provided by app and called to write header
	headerHomeMode   bool                       // set position to home after headerFnc is called
	inFooter         bool                       // flag set when processing footer
	footerFnc        func()                     // function provided by app and called to write footer
	footerFncLpi     func(bool)                 // function provided by app and called to write footer with last page flag
	zoomMode         string                     // zoom display mode
	layoutMode       string                     // layout display mode
	nXMP             int                        // XMP object number
	xmp              []byte                     // XMP metadata
	producer         string                     // producer
	title            string                     // title
	subject          string                     // subject
	author           string                     // author
	lang             string                     // lang
	keywords         string                     // keywords
	creator          string                     // creator
	creationDate     time.Time                  // override for document CreationDate value
	modDate          time.Time                  // override for document ModDate value
	aliasNbPagesStr  string                     // alias for total number of pages
	pdfVersion       pdfVersion                 // PDF version number
	capStyle         int                        // line cap style: butt 0, round 1, square 2
	joinStyle        int                        // line segment join style: miter 0, round 1, bevel 2
	dashArray        []float64                  // dash array
	dashPhase        float64                    // dash phase
	blendList        []blendModeType            // slice[idx] of alpha transparency modes, 1-based
	blendMap         map[string]int             // map into blendList
	blendMode        string                     // current blend mode
	alpha            float64                    // current transpacency
	gradientList     []gradientType             // slice[idx] of gradient records
	clipNest         int                        // Number of active clipping contexts
	transformNest    int                        // Number of active transformation contexts
	err              error                      // Set if error occurs during life cycle of instance
	protect          protectType                // document protection structure
	layer            layerRecType               // manages optional layers in document
	catalogSort      bool                       // sort resource catalogs in document
	nJs              int                        // JavaScript object number
	javascript       *string                    // JavaScript code to include in the PDF
	colorFlag        bool                       // indicates whether fill and text colors are different
	color            struct {
		// Composite values of colors
		draw, fill, text colorType
	}
	spotColorMap           map[string]spotColorType // Map of named ink-based colors
	outputIntents          []OutputIntentType       // OutputIntents
	outputIntentStartN     int                      // Start object number for
	userUnderlineThickness float64                  // A custom user underline thickness multiplier.

	fmt struct {
		buf []byte       // buffer used to format numbers.
		col bytes.Buffer // buffer used to build color strings.
	}
}

const (
	pdfVers1_3 = pdfVersion(uint16(1)<<8 | uint16(3))
	pdfVers1_4 = pdfVersion(uint16(1)<<8 | uint16(4))
	pdfVers1_5 = pdfVersion(uint16(1)<<8 | uint16(5))
)

type pdfVersion uint16

func pdfVersionFrom(maj, min uint) pdfVersion {
	if min > 255 {
		panic(tinystring.Err(tinystring.D.Format, tinystring.D.Invalid, maj, min))
	}
	return pdfVersion(uint16(maj)<<8 | uint16(min))
}

func (v pdfVersion) String() string {
	maj := int64(byte(v >> 8))
	min := int64(byte(v))
	return tinystring.Fmt("%d.%d", maj, min)
}

type encType struct {
	uv   int
	name string
}

type encListType [256]encType

type fontBoxType struct {
	Xmin, Ymin, Xmax, Ymax int
}

// Font flags for FontDescType.Flags as defined in the pdf specification.
const (
	// FontFlagFixedPitch is set if all glyphs have the same width (as
	// opposed to proportional or variable-pitch fonts, which have
	// different widths).
	FontFlagFixedPitch = 1 << 0
	// FontFlagSerif is set if glyphs have serifs, which are short
	// strokes drawn at an angle on the top and bottom of glyph stems.
	// (Sans serif fonts do not have serifs.)
	FontFlagSerif = 1 << 1
	// FontFlagSymbolic is set if font contains glyphs outside the
	// Adobe standard Latin character set. This flag and the
	// Nonsymbolic flag shall not both be set or both be clear.
	FontFlagSymbolic = 1 << 2
	// FontFlagScript is set if glyphs resemble cursive handwriting.
	FontFlagScript = 1 << 3
	// FontFlagNonsymbolic is set if font uses the Adobe standard
	// Latin character set or a subset of it.
	FontFlagNonsymbolic = 1 << 5
	// FontFlagItalic is set if glyphs have dominant vertical strokes
	// that are slanted.
	FontFlagItalic = 1 << 6
	// FontFlagAllCap is set if font contains no lowercase letters;
	// typically used for display purposes, such as for titles or
	// headlines.
	FontFlagAllCap = 1 << 16
	// SmallCap is set if font contains both uppercase and lowercase
	// letters. The uppercase letters are similar to those in the
	// regular version of the same typeface family. The glyphs for the
	// lowercase letters have the same shapes as the corresponding
	// uppercase letters, but they are sized and their proportions
	// adjusted so that they have the same size and stroke weight as
	// lowercase glyphs in the same typeface family.
	SmallCap = 1 << 18
	// ForceBold determines whether bold glyphs shall be painted with
	// extra pixels even at very small text sizes by a conforming
	// reader. If the ForceBold flag is set, features of bold glyphs
	// may be thickened at small text sizes.
	ForceBold = 1 << 18
)

// FontDescType (font descriptor) specifies metrics and other
// attributes of a font, as distinct from the metrics of individual
// glyphs (as defined in the pdf specification).
type FontDescType struct {
	// The maximum height above the baseline reached by glyphs in this
	// font (for example for "S"). The height of glyphs for accented
	// characters shall be excluded.
	Ascent int
	// The maximum depth below the baseline reached by glyphs in this
	// font. The value shall be a negative number.
	Descent int
	// The vertical coordinate of the top of flat capital letters,
	// measured from the baseline (for example "H").
	CapHeight int
	// A collection of flags defining various characteristics of the
	// font. (See the FontFlag* constants.)
	Flags int
	// A rectangle, expressed in the glyph coordinate system, that
	// shall specify the font bounding box. This should be the smallest
	// rectangle enclosing the shape that would result if all of the
	// glyphs of the font were placed with their origins coincident
	// and then filled.
	FontBBox fontBoxType
	// The angle, expressed in degrees counterclockwise from the
	// vertical, of the dominant vertical strokes of the font. (The
	// 9-o’clock position is 90 degrees, and the 3-o’clock position
	// is –90 degrees.) The value shall be negative for fonts that
	// slope to the right, as almost all italic fonts do.
	ItalicAngle int
	// The thickness, measured horizontally, of the dominant vertical
	// stems of glyphs in the font.
	StemV int
	// The width to use for character codes whose widths are not
	// specified in a font dictionary’s Widths array. This shall have
	// a predictable effect only if all such codes map to glyphs whose
	// actual widths are the same as the value of the MissingWidth
	// entry. (Default value: 0.)
	MissingWidth int
}

type fontDefType struct {
	Tp           string        // "Core", "TrueType", ...
	Name         string        // "Courier-Bold", ...
	Desc         FontDescType  // Font descriptor
	Up           int           // Underline position
	Ut           int           // Underline thickness
	Cw           []int         // Character width by ordinal
	Enc          string        // "cp1252", ...
	Diff         string        // Differences from reference encoding
	File         string        // "Redressed.z"
	Size1, Size2 int           // Type1 values
	OriginalSize int           // Size of uncompressed font file
	N            int           // Set by font loader
	DiffN        int           // Position of diff in app array, set by font loader
	i            string        // 1-based position in font list, set by font loader, not this program
	utf8File     *utf8FontFile // UTF-8 font
	usedRunes    map[int]int   // Array of used runes
}

// generateFontID generates a font Id from the font definition
func generateFontID(fdt fontDefType) (string, error) {
	// file can be different if generated in different instance
	fdt.File = ""
	b, err := json.Marshal(&fdt)
	return tinystring.Fmt("%x", sha1.Sum(b)), err
}

type fontInfoType struct {
	Data               []byte
	File               string
	OriginalSize       int
	FontName           string
	Bold               bool
	IsFixedPitch       bool
	UnderlineThickness int
	UnderlinePosition  int
	Widths             []int
	Size1, Size2       uint32
	Desc               FontDescType
}
