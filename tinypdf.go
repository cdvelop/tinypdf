package tinypdf

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"time"

	"github.com/cdvelop/tinypdf/fontManager"
	. "github.com/cdvelop/tinystring"
)

var gl struct {
	catalogSort  bool
	noCompress   bool // Initial zero value indicates compression
	creationDate time.Time
	modDate      time.Time
}

type fmtBuffer struct {
	bytes.Buffer
}

func (b *fmtBuffer) printf(fmtStr string, args ...any) {
	b.Buffer.WriteString(Fmt(fmtStr, args...))
}

func New(fontsPath []string, logger func(...any)) (f *TinyPDF) {
	f = new(TinyPDF)
	f.log = logger
	f.fm = fontManager.New(&fontManager.Config{
		Log:                 logger,
		FontsPath:           fontsPath,
		ConversionRatio:     f.ConversionRatio,
		CurrentObjectNumber: f.CurrentObjectNumber,
		SetFontCB: func(family, style string, size float64) {
			// proxy back to TinyPDF's SetFont implementation
			f.SetFont(family, style, size)
		},
		FontFamilyEscape: f.fontFamilyEscape,
	})

	var size = PageSize{0, 0, false}
	var initType *InitType

	// Set default values
	f.defOrientation = Portrait
	f.rootDirectory = "."
	f.unitType = MM

	if initType != nil {
		f.defOrientation = initType.OrientationStr
		if f.defOrientation == "" {
			f.defOrientation = Portrait
		}
		if initType.UnitType != "" {
			f.unitType = initType.UnitType
		}

		// Note: page size conversion happens later after scale factor is set
		f.rootDirectory = initType.RootDirectory
	}

	f.page = 0
	f.n = 2
	f.pages = make([]*bytes.Buffer, 0, 8)
	f.pages = append(f.pages, bytes.NewBufferString("")) // pages[0] is unused (1-based)
	f.pageSizes = make(map[int]PageSize)
	f.pageBoxes = make(map[int]map[string]PageBox)
	f.defPageBoxes = make(map[string]PageBox)
	f.state = 0
	// Template/imported-template support (minimal for gofpdi)
	f.importedObjs = make(map[string][]byte)
	f.importedObjPos = make(map[string]map[int]string)
	f.importedTplObjs = make(map[string]string)
	f.importedTplIDs = make(map[string]int)
	f.images = make(map[string]*ImageInfoType)
	f.pageLinks = make([][]linkType, 0, 8)
	f.pageLinks = append(f.pageLinks, make([]linkType, 0)) // pageLinks[0] is unused (1-based)
	f.links = make([]intLinkType, 0, 8)
	f.links = append(f.links, intLinkType{}) // links[0] is unused (1-based)
	f.pageAttachments = make([][]annotationAttach, 0, 8)
	f.pageAttachments = append(f.pageAttachments, []annotationAttach{}) //
	f.aliasMap = make(map[string]string)
	f.inHeader = false
	f.inFooter = false
	f.lasth = 0
	f.SetFontSize(12)
	f.underline = false
	f.strikeout = false
	f.setDrawColor(0, 0, 0)
	f.setFillColor(0, 0, 0)
	f.setTextColor(0, 0, 0)
	f.colorFlag = false
	f.ws = 0

	// Scale factor
	switch f.unitType {
	case POINT:
		f.k = 1.0
	case MM:
		f.k = 72.0 / 25.4
	case CM:
		f.k = 72.0 / 2.54
	case IN:
		f.k = 72.0
	default:
		f.err = Err(D.Invalid, D.Format)
		return
	}
	f.stdPageSizes = make(map[string]PageSize)
	f.stdPageSizes["a3"] = A3
	f.stdPageSizes["a4"] = A4
	f.stdPageSizes["a5"] = A5
	f.stdPageSizes["a6"] = A6
	f.stdPageSizes["a7"] = A7
	f.stdPageSizes["a2"] = A2
	f.stdPageSizes["a1"] = A1
	f.stdPageSizes["letter"] = Letter
	f.stdPageSizes["legal"] = Legal
	f.stdPageSizes["tabloid"] = Tabloid

	// Set default page size
	if initType != nil && initType.Size.Wd > 0 && initType.Size.Ht > 0 {
		// Convert user-specified page size to points for internal use
		f.defPageSize = PageSize{
			Wd:     initType.Size.Wd * f.k,
			Ht:     initType.Size.Ht * f.k,
			AutoHt: initType.Size.AutoHt,
		}
	} else if size.Wd > 0 && size.Ht > 0 {
		// Convert deprecated size parameter to points for internal use
		f.defPageSize = PageSize{
			Wd:     size.Wd * f.k,
			Ht:     size.Ht * f.k,
			AutoHt: size.AutoHt,
		}
	} else {
		// Use default A4 size in points
		f.defPageSize = A4
	}
	f.curPageSize = f.defPageSize // Page orientation
	switch f.defOrientation {
	case Portrait:
		f.w = f.defPageSize.Wd / f.k
		f.h = f.defPageSize.Ht / f.k
		// dbg("Assign h: %8.2f", f.h)
	case Landscape:
		f.w = f.defPageSize.Ht / f.k
		f.h = f.defPageSize.Wd / f.k
	default:
		f.err = Err(D.Invalid, D.Format)
		return
	}
	f.curOrientation = f.defOrientation
	f.wPt = f.w * f.k
	f.hPt = f.h * f.k
	// Page margins (1 cm)
	margin := 28.35 / f.k
	f.SetMargins(margin, margin, margin)
	// Interior cell margin (1 mm)
	f.cMargin = margin / 10
	// Line width (0.2 mm)
	f.lineWidth = 0.567 / f.k
	// 	Automatic page break
	f.SetAutoPageBreak(true, 2*margin)
	// Default display mode
	f.SetDisplayMode("default", "default")
	if f.err != nil {
		return
	}
	f.acceptPageBreak = func() bool {
		return f.autoPageBreak
	}
	// Enable compression
	f.SetCompression(!gl.noCompress)
	f.spotColorMap = make(map[string]spotColorType)
	f.blendList = make([]blendModeType, 0, 8)
	f.blendList = append(f.blendList, blendModeType{}) // blendList[0] is unused (1-based)
	f.blendMap = make(map[string]int)
	f.blendMode = "Normal"
	f.alpha = 1
	f.gradientList = make([]gradientType, 0, 8)
	f.gradientList = append(f.gradientList, gradientType{}) // gradientList[0] is unused
	// Set default PDF version number
	f.pdfVersion = pdfVers1_3
	f.SetProducer("FPDF "+cnFpdfVersion, true)
	f.layerInit()
	f.catalogSort = gl.catalogSort
	f.creationDate = gl.creationDate
	f.modDate = gl.modDate
	f.userUnderlineThickness = 1

	// create a large enough buffer for formatting float64s.
	// math.MaxInt64  needs 19.
	// math.MaxUint64 needs 20.
	f.fmt.buf = make([]byte, 24)
	return
}

// Ok returns true if no processing errors have occurred.
func (f *TinyPDF) Ok() bool {
	return f.err == nil
}

// Err returns true if a processing error has occurred.
func (f *TinyPDF) Err() bool {
	return f.err != nil
}

// ClearError unsets the internal TinyPDF error. This method should be used with
// care, as an internal error condition usually indicates an unrecoverable
// problem with the generation of a document. It is intended to deal with cases
// in which an error is used to select an alternate form of the document.
func (f *TinyPDF) ClearError() {
	f.err = nil
}

// SetErrorf sets the internal TinyPDF error with formatted text to halt PDF
// generation; this may facilitate error handling by application. If an error
// condition is already set, this call is ignored.
//
// See the documentation for printing in the standard fmt package for details
// about fmtStr and args.
func (f *TinyPDF) SetErrorf(fmtStr string, args ...interface{}) {
	if f.err == nil {
		f.err = Errf(fmtStr, args...)
	}
}

// String satisfies the fmt.Stringer interface and summarizes the TinyPDF
// instance.
func (f *TinyPDF) String() string {
	return "TinyPDF " + cnFpdfVersion
}

// SetError sets an error to halt PDF generation. This may facilitate error
// handling by application. See also Ok(), Err() and Error().
func (f *TinyPDF) SetError(err error) {
	if f.err == nil && err != nil {
		f.err = err
	}
}

// Error returns the internal TinyPDF error; this will be nil if no error has occurred.
func (f *TinyPDF) Error() error {
	return f.err
}

// GetCellMargin returns the cell margin. This is the amount of space before
// and after the text within a cell that's left blank, and is in units passed
// to New(). It defaults to 1mm.
func (f *TinyPDF) GetCellMargin() float64 {
	return f.cMargin
}

// SetCellMargin sets the cell margin. This is the amount of space before and
// after the text within a cell that's left blank, and is in units passed to
// New().
func (f *TinyPDF) SetCellMargin(margin float64) {
	f.cMargin = margin
}

// SetDefaultCompression controls the default setting of the internal
// compression flag. See SetCompression() for more details. Compression is on
// by default.
func SetDefaultCompression(compress bool) {
	gl.noCompress = !compress
}

// GetCompression returns whether page compression is enabled.
func (f *TinyPDF) GetCompression() bool {
	return f.compress
}

// SetCompression activates or deactivates page compression with zlib. When
// activated, the internal representation of each page is compressed, which
// leads to a compression ratio of about 2 for the resulting document.
// Compression is on by default.
func (f *TinyPDF) SetCompression(compress bool) {
	f.compress = compress
}

// GetProducer returns the producer of the document as ISO-8859-1 or UTF-16BE.
func (f *TinyPDF) GetProducer() string {
	return f.producer
}

// SetProducer defines the producer of the document. isUTF8 indicates if the string
// is encoded in ISO-8859-1 (false) or UTF-8 (true).
func (f *TinyPDF) SetProducer(producerStr string, isUTF8 bool) {
	if isUTF8 {
		producerStr = utf8toutf16(producerStr)
	}
	f.producer = producerStr
}

// GetTitle returns the title of the document as ISO-8859-1 or UTF-16BE.
func (f *TinyPDF) GetTitle() string {
	return f.title
}

// SetTitle defines the title of the document. isUTF8 indicates if the string
// is encoded in ISO-8859-1 (false) or UTF-8 (true).
func (f *TinyPDF) SetTitle(titleStr string, isUTF8 bool) {
	if isUTF8 {
		titleStr = utf8toutf16(titleStr)
	}
	f.title = titleStr
}

// GetSubject returns the subject of the document as ISO-8859-1 or UTF-16BE.
func (f *TinyPDF) GetSubject() string {
	return f.subject
}

// SetSubject defines the subject of the document. isUTF8 indicates if the
// string is encoded in ISO-8859-1 (false) or UTF-8 (true).
func (f *TinyPDF) SetSubject(subjectStr string, isUTF8 bool) {
	if isUTF8 {
		subjectStr = utf8toutf16(subjectStr)
	}
	f.subject = subjectStr
}

// GetAuthor returns the author of the document as ISO-8859-1 or UTF-16BE.
func (f *TinyPDF) GetAuthor() string {
	return f.author
}

// SetAuthor defines the author of the document. isUTF8 indicates if the string
// is encoded in ISO-8859-1 (false) or UTF-8 (true).
func (f *TinyPDF) SetAuthor(authorStr string, isUTF8 bool) {
	if isUTF8 {
		authorStr = utf8toutf16(authorStr)
	}
	f.author = authorStr
}

// GetLang returns the natural language of the document (e.g. "de-CH").
func (f *TinyPDF) GetLang() string {
	return f.lang
}

// SetLang defines the natural language of the document (e.g. "de-CH").
func (f *TinyPDF) SetLang(lang string) {
	f.lang = lang
}

// GetKeywords returns the keywords of the document as ISO-8859-1 or UTF-16BE.
func (f *TinyPDF) GetKeywords() string {
	return f.keywords
}

// SetKeywords defines the keywords of the document. keywordStr is a
// space-delimited string, for example "invoice August". isUTF8 indicates if
// the string is encoded
func (f *TinyPDF) SetKeywords(keywordsStr string, isUTF8 bool) {
	if isUTF8 {
		keywordsStr = utf8toutf16(keywordsStr)
	}
	f.keywords = keywordsStr
}

// GetCreator returns the creator of the document as ISO-8859-1 or UTF-16BE.
func (f *TinyPDF) GetCreator() string {
	return f.creator
}

// SetCreator defines the creator of the document. isUTF8 indicates if the
// string is encoded in ISO-8859-1 (false) or UTF-8 (true).
func (f *TinyPDF) SetCreator(creatorStr string, isUTF8 bool) {
	if isUTF8 {
		creatorStr = utf8toutf16(creatorStr)
	}
	f.creator = creatorStr
}

// GetXmpMetadata returns the XMP metadata that will be embedded with the document.
func (f *TinyPDF) GetXmpMetadata() []byte {
	return []byte(string(f.xmp))
}

// SetXmpMetadata defines XMP metadata that will be embedded with the document.
func (f *TinyPDF) SetXmpMetadata(xmpStream []byte) {
	f.xmp = xmpStream
}

// AddOutputIntent adds an output intent with ICC color profile
func (f *TinyPDF) AddOutputIntent(outputIntent OutputIntentType) {
	f.outputIntents = append(f.outputIntents, outputIntent)
	if f.pdfVersion < pdfVers1_4 {
		f.pdfVersion = pdfVers1_4
	}
}

// AliasNbPages defines an alias for the total number of pages. It will be
// substituted as the document is closed. An empty string is replaced with the
// string "{nb}".
//
// See the example for AddPage() for a demonstration of this method.
func (f *TinyPDF) AliasNbPages(aliasStr string) {
	if aliasStr == "" {
		aliasStr = "{nb}"
	}
	f.aliasNbPagesStr = aliasStr
}

// RTL enables right-to-left mode
func (f *TinyPDF) RTL() {
	f.isRTL = true
}

// LTR disables right-to-left mode
func (f *TinyPDF) LTR() {
	f.isRTL = false
}

// open begins a document
func (f *TinyPDF) open() {
	f.state = 1
}

// Close terminates the PDF document. It is not necessary to call this method
// explicitly because Output(), OutputAndClose() and OutputFileAndClose() do it
// automatically. If the document contains no page, AddPage() is called to
// prevent the generation of an invalid document.
func (f *TinyPDF) Close() {
	if f.err == nil {
		if f.clipNest > 0 {
			f.err = Errf("clip procedure must be explicitly ended")
		} else if f.transformNest > 0 {
			f.err = Errf("transformation procedure must be explicitly ended")
		}
	}
	if f.err != nil {
		return
	}
	if f.state == 3 {
		return
	}
	if f.page == 0 {
		f.AddPage()
		if f.err != nil {
			return
		}
	}
	// Page footer
	f.inFooter = true
	if f.footerFnc != nil {
		f.footerFnc()
	} else if f.footerFncLpi != nil {
		f.footerFncLpi(true)
	}
	f.inFooter = false

	// Close page
	f.endpage()
	// Close document
	f.enddoc()
}

func colorComp(v int) (int, float64) {
	if v < 0 {
		v = 0
	} else if v > 255 {
		v = 255
	}
	return v, float64(v) / 255.0
}

func (f *TinyPDF) rgbColorValue(r, g, b int, grayStr, fullStr string) (clr colorType) {
	clr.ir, clr.r = colorComp(r)
	clr.ig, clr.g = colorComp(g)
	clr.ib, clr.b = colorComp(b)
	clr.mode = colorModeRGB
	clr.gray = clr.ir == clr.ig && clr.r == clr.b
	const prec = 3
	if len(grayStr) > 0 {
		if clr.gray {
			// clr.str = Fmt("%.3f %s", clr.r, grayStr)
			f.fmt.col.Reset()
			f.fmt.col.WriteString(f.fmtF64(clr.r, prec))
			f.fmt.col.WriteString(" ")
			f.fmt.col.WriteString(grayStr)
			clr.str = f.fmt.col.String()
		} else {
			// clr.str = Fmt("%.3f %.3f %.3f %s", clr.r, clr.g, clr.b, fullStr)
			f.fmt.col.Reset()
			f.fmt.col.WriteString(f.fmtF64(clr.r, prec))
			f.fmt.col.WriteString(" ")
			f.fmt.col.WriteString(f.fmtF64(clr.g, prec))
			f.fmt.col.WriteString(" ")
			f.fmt.col.WriteString(f.fmtF64(clr.b, prec))
			f.fmt.col.WriteString(" ")
			f.fmt.col.WriteString(fullStr)
			clr.str = f.fmt.col.String()
		}
	} else {
		// clr.str = Fmt("%.3f %.3f %.3f", clr.r, clr.g, clr.b)
		f.fmt.col.Reset()
		f.fmt.col.WriteString(f.fmtF64(clr.r, prec))
		f.fmt.col.WriteString(" ")
		f.fmt.col.WriteString(f.fmtF64(clr.g, prec))
		f.fmt.col.WriteString(" ")
		f.fmt.col.WriteString(f.fmtF64(clr.b, prec))
		clr.str = f.fmt.col.String()
	}
	return
}

// AddLink creates a new internal link and returns its identifier. An internal
// link is a clickable area which directs to another place within the document.
// The identifier can then be passed to Cell(), Write(), Image() or Link(). The
// destination is defined with SetLink().
func (f *TinyPDF) AddLink() int {
	f.links = append(f.links, intLinkType{})
	return len(f.links) - 1
}

// SetLink defines the page and position a link points to. See AddLink().
func (f *TinyPDF) SetLink(link int, y float64, page int) {
	if y == -1 {
		y = f.y
	}
	if page == -1 {
		page = f.page
	}
	f.links[link] = intLinkType{page, y}
}

// newLink adds a new clickable link on current page
func (f *TinyPDF) newLink(x, y, w, h float64, link int, linkStr string) {
	// linkList, ok := f.pageLinks[f.page]
	// if !ok {
	// linkList = make([]linkType, 0, 8)
	// f.pageLinks[f.page] = linkList
	// }
	f.pageLinks[f.page] = append(f.pageLinks[f.page],
		linkType{x * f.k, f.hPt - y*f.k, w * f.k, h * f.k, link, linkStr})
}

// Link puts a link on a rectangular area of the page. Text or image links are
// generally put via Cell(), Write() or Image(), but this method can be useful
// for instance to define a clickable area inside an image. link is the value
// returned by AddLink().
func (f *TinyPDF) Link(x, y, w, h float64, link int) {
	f.newLink(x, y, w, h, link, "")
}

// LinkString puts a link on a rectangular area of the page. Text or image
// links are generally put via Cell(), Write() or Image(), but this method can
// be useful for instance to define a clickable area inside an image. linkStr
// is the target URL.
func (f *TinyPDF) LinkString(x, y, w, h float64, linkStr string) {
	f.newLink(x, y, w, h, 0, linkStr)
}

// Bookmark sets a bookmark that will be displayed in a sidebar outline. txtStr
// is the title of the bookmark. level specifies the level of the bookmark in
// the outline; 0 is the top level, 1 is just below, and so on. y specifies the
// vertical position of the bookmark destination in the current page; -1
// indicates the current position.
func (f *TinyPDF) Bookmark(txtStr string, level int, y float64) {
	if y == -1 {
		y = f.y
	}
	if f.isCurrentUTF8 {
		txtStr = utf8toutf16(txtStr)
	}
	f.outlines = append(f.outlines, outlineType{text: txtStr, level: level, y: y, p: f.PageNo(), prev: -1, last: -1, next: -1, first: -1})
}

// GetWordSpacing returns the spacing between words of following text.
func (f *TinyPDF) GetWordSpacing() float64 {
	return f.ws
}

// SetWordSpacing sets spacing between words of following text. See the
// WriteAligned() example for a demonstration of its use.
func (f *TinyPDF) SetWordSpacing(space float64) {
	f.ws = space
	f.out(Fmt("%.5f Tw", space*f.k))
}

// SetTextRenderingMode sets the rendering mode of following text.
// The mode can be as follows:
// 0: Fill text
// 1: Stroke text
// 2: Fill, then stroke text
// 3: Neither fill nor stroke text (invisible)
// 4: Fill text and add to path for clipping
// 5: Stroke text and add to path for clipping
// 6: Fills then stroke text and add to path for clipping
// 7: Add text to path for clipping
// This method is demonstrated in the SetTextRenderingMode example.
func (f *TinyPDF) SetTextRenderingMode(mode int) {
	if mode >= 0 && mode <= 7 {
		f.out(Fmt("%d Tr", mode))
	}
}

// Revert string to use in RTL languages
func reverseText(text string) string {
	oldText := []rune(text)
	newText := make([]rune, len(oldText))
	length := len(oldText) - 1
	for i, r := range oldText {
		newText[length-i] = r
	}
	return string(newText)
}

// Cell is a simpler version of CellFormat with no fill, border, links or
// special alignment. The Cell_strikeout() example demonstrates this method.
func (f *TinyPDF) Cell(w, h float64, txtStr string) {
	f.CellFormat(w, h, txtStr, "", 0, "L", false, 0, "")
}

// Cellf is a simpler printf-style version of CellFormat with no fill, border,
// links or special alignment. See documentation for the fmt package for
// details on fmtStr and args.
func (f *TinyPDF) Cellf(w, h float64, fmtStr string, args ...interface{}) {
	f.CellFormat(w, h, Fmt(fmtStr, args...), "", 0, "L", false, 0, "")
}

// SplitLines splits text into several lines using the current font. Each line
// has its length limited to a maximum width given by w. This function can be
// used to determine the total height of wrapped text for vertical placement
// purposes.
//
// This method is useful for codepage-based fonts only. For UTF-8 encoded text,
// use SplitText().
//
// You can use MultiCell if you want to print a text on several lines in a
// simple way.
func (f *TinyPDF) SplitLines(txt []byte, w float64) [][]byte {
	// Function contributed by Bruno Michel
	lines := [][]byte{}
	cw := f.currentFont.Cw
	wmax := int(math.Ceil((w - 2*f.cMargin) * 1000 / f.fontSize))
	s := []byte(Convert(string(txt)).Replace("\r", "").String())
	nb := len(s)
	for nb > 0 && s[nb-1] == '\n' {
		nb--
	}
	s = s[0:nb]
	sep := -1
	i := 0
	j := 0
	l := 0
	for i < nb {
		c := s[i]
		l += cw[c]
		if c == ' ' || c == '\t' || c == '\n' {
			sep = i
		}
		if c == '\n' || l > wmax {
			if sep == -1 {
				if i == j {
					i++
				}
				sep = i
			} else {
				i = sep + 1
			}
			lines = append(lines, s[j:sep])
			sep = -1
			j = i
			l = 0
		} else {
			i++
		}
	}
	if i != j {
		lines = append(lines, s[j:i])
	}
	return lines
}

// write outputs text in flowing mode
func (f *TinyPDF) write(h float64, txtStr string, link int, linkStr string) {
	// dbg("Write")
	cw := f.currentFont.Cw
	w := f.w - f.rMargin - f.x
	wmax := (w - 2*f.cMargin) * 1000 / f.fontSize
	s := Convert(txtStr).Replace("\r", "").String()
	var nb int
	if f.isCurrentUTF8 {
		nb = len([]rune(s))
		if nb == 1 && s == " " {
			f.x += f.GetStringWidth(s)
			return
		}
	} else {
		nb = len(s)
	}
	sep := -1
	i := 0
	j := 0
	l := 0.0
	nl := 1
	for i < nb {
		// Get next character
		var c rune
		if f.isCurrentUTF8 {
			c = []rune(s)[i]
		} else {
			c = rune(byte(s[i]))
		}
		if c == '\n' {
			// Explicit line break
			if f.isCurrentUTF8 {
				f.CellFormat(w, h, string([]rune(s)[j:i]), "", 2, "", false, link, linkStr)
			} else {
				f.CellFormat(w, h, s[j:i], "", 2, "", false, link, linkStr)
			}
			i++
			sep = -1
			j = i
			l = 0.0
			if nl == 1 {
				f.x = f.lMargin
				w = f.w - f.rMargin - f.x
				wmax = (w - 2*f.cMargin) * 1000 / f.fontSize
			}
			nl++
			continue
		}
		if c == ' ' {
			sep = i
		}
		l += float64(cw[int(c)])
		if l > wmax {
			// Automatic line break
			if sep == -1 {
				if f.x > f.lMargin {
					// Move to next line
					f.x = f.lMargin
					f.y += h
					w = f.w - f.rMargin - f.x
					wmax = (w - 2*f.cMargin) * 1000 / f.fontSize
					i++
					nl++
					continue
				}
				if i == j {
					i++
				}
				if f.isCurrentUTF8 {
					f.CellFormat(w, h, string([]rune(s)[j:i]), "", 2, "", false, link, linkStr)
				} else {
					f.CellFormat(w, h, s[j:i], "", 2, "", false, link, linkStr)
				}
			} else {
				if f.isCurrentUTF8 {
					f.CellFormat(w, h, string([]rune(s)[j:sep]), "", 2, "", false, link, linkStr)
				} else {
					f.CellFormat(w, h, s[j:sep], "", 2, "", false, link, linkStr)
				}
				i = sep + 1
			}
			sep = -1
			j = i
			l = 0.0
			if nl == 1 {
				f.x = f.lMargin
				w = f.w - f.rMargin - f.x
				wmax = (w - 2*f.cMargin) * 1000 / f.fontSize
			}
			nl++
		} else {
			i++
		}
	}
	// Last chunk
	if i != j {
		if f.isCurrentUTF8 {
			f.CellFormat(l/1000*f.fontSize, h, string([]rune(s)[j:]), "", 0, "", false, link, linkStr)
		} else {
			f.CellFormat(l/1000*f.fontSize, h, s[j:], "", 0, "", false, link, linkStr)
		}
	}
}

// Write prints text from the current position. When the right margin is
// reached (or the \n character is met) a line break occurs and text continues
// from the left margin. Upon method exit, the current position is left just at
// the end of the text.
//
// It is possible to put a link on the text.
//
// h indicates the line height in the Unit of measure specified in New().
func (f *TinyPDF) Write(h float64, txtStr string) {
	f.write(h, txtStr, 0, "")
}

// Writef is like Write but uses printf-style formatting. See the documentation
// for package fmt for more details on fmtStr and args.
func (f *TinyPDF) Writef(h float64, fmtStr string, args ...interface{}) {
	f.write(h, Fmt(fmtStr, args...), 0, "")
}

// WriteLinkString writes text that when clicked launches an external URL. See
// Write() for argument details.
func (f *TinyPDF) WriteLinkString(h float64, displayStr, targetStr string) {
	f.write(h, displayStr, 0, targetStr)
}

// WriteLinkID writes text that when clicked jumps to another location in the
// PDF. linkID is an identifier returned by AddLink(). See Write() for argument
// details.
func (f *TinyPDF) WriteLinkID(h float64, displayStr string, linkID int) {
	f.write(h, displayStr, linkID, "")
}

// WriteAligned is an implementation of Write that makes it possible to align
// text.
//
// width indicates the width of the box the text will be drawn in. This is in
// the Unit of measure specified in New(). If it is set to 0, the bounding box
// of the page will be taken (pageWidth - leftMargin - rightMargin).
//
// lineHeight indicates the line height in the Unit of measure specified in
// New().
//
// alignStr sees to horizontal alignment of the given textStr. The options are
// "L", "C" and "R" (Left, Center, Right). The default is "L".
func (f *TinyPDF) WriteAligned(width, lineHeight float64, textStr, alignStr string) {
	lMargin, _, rMargin, _ := f.GetMargins()

	pageWidth, _ := f.GetPageSize()
	if width == 0 {
		width = pageWidth - (lMargin + rMargin)
	}

	var lines []string

	if f.isCurrentUTF8 {
		lines = f.SplitText(textStr, width)
	} else {
		for _, line := range f.SplitLines([]byte(textStr), width) {
			lines = append(lines, string(line))
		}
	}

	for _, lineBt := range lines {
		lineStr := string(lineBt)
		lineWidth := f.GetStringWidth(lineStr)

		switch alignStr {
		case "C":
			f.SetLeftMargin(lMargin + ((width - lineWidth) / 2))
			f.Write(lineHeight, lineStr)
			f.SetLeftMargin(lMargin)
		case "R":
			f.SetLeftMargin(lMargin + (width - lineWidth) - 2.01*f.cMargin)
			f.Write(lineHeight, lineStr)
			f.SetLeftMargin(lMargin)
		default:
			f.SetRightMargin(pageWidth - lMargin - width)
			f.Write(lineHeight, lineStr)
			f.SetRightMargin(rMargin)
		}
	}
}

// Ln performs a line break. The current abscissa goes back to the left margin
// and the ordinate increases by the amount passed in parameter. A negative
// value of h indicates the height of the last printed cell.
//
// This method is demonstrated in the example for MultiCell.
func (f *TinyPDF) Ln(h float64) {
	f.x = f.lMargin
	if h < 0 {
		f.y += f.lasth
	} else {
		f.y += h
	}
}

// ImageTypeFromMime returns the image type used in various image-related
// functions (for example, Image()) that is associated with the specified MIME
// type. For example, "jpg" is returned if mimeStr is "image/jpeg". An error is
// set if the specified MIME type is not supported.
func (f *TinyPDF) ImageTypeFromMime(mimeStr string) (tp string) {
	switch mimeStr {
	case "image/png":
		tp = "png"
	case "image/jpg":
		tp = "jpg"
	case "image/jpeg":
		tp = "jpg"
	case "image/gif":
		tp = "gif"
	default:
		f.SetErrorf("unsupported image type: %s", mimeStr)
	}
	return
}

func (f *TinyPDF) imageOut(info *ImageInfoType, x, y, w, h float64, allowNegativeX, flow bool, link int, linkStr string) {
	// Automatic width and height calculation if needed
	if w == 0 && h == 0 {
		// Put image at 96 dpi
		w = -96
		h = -96
	}
	if w == -1 {
		// Set image width to whatever value for dpi we read
		// from the image or that was set manually
		w = -info.dpi
	}
	if h == -1 {
		// Set image height to whatever value for dpi we read
		// from the image or that was set manually
		h = -info.dpi
	}
	if w < 0 {
		w = -info.w * 72.0 / w / f.k
	}
	if h < 0 {
		h = -info.h * 72.0 / h / f.k
	}
	if w == 0 {
		w = h * info.w / info.h
	}
	if h == 0 {
		h = w * info.h / info.w
	}
	// Flowing mode
	if flow {
		if f.y+h > f.pageBreakTrigger && !f.inHeader && !f.inFooter && f.acceptPageBreak() {
			// Automatic page break
			x2 := f.x
			f.AddPageFormat(f.curOrientation, f.curPageSize)
			if f.err != nil {
				return
			}
			f.x = x2
		}
		y = f.y
		f.y += h
	}
	if !allowNegativeX {
		if x < 0 {
			x = f.x
		}
	}
	// dbg("h %.2f", h)
	// q 85.04 0 0 NaN 28.35 NaN cm /I2 Do Q
	// f.outf("q %.5f 0 0 %.5f %.5f %.5f cm /I%s Do Q", w*f.k, h*f.k, x*f.k, (f.h-(y+h))*f.k, info.i)
	const prec = 5
	f.put("q ")
	f.putF64(w*f.k, prec)
	f.put(" 0 0 ")
	f.putF64(h*f.k, prec)
	f.put(" ")
	f.putF64(x*f.k, prec)
	f.put(" ")
	f.putF64((f.h-(y+h))*f.k, prec)
	f.put(" cm /I" + info.i + " Do Q\n")
	if link > 0 || len(linkStr) > 0 {
		f.newLink(x, y, w, h, link, linkStr)
	}
}

// Image puts a JPEG, PNG or GIF image in the current page.
//
// Deprecated in favor of ImageOptions -- see that function for
// details on the behavior of arguments
func (f *TinyPDF) Image(imageNameStr string, x, y, w, h float64, flow bool, tp string, link int, linkStr string) {
	options := ImageOptions{
		ReadDpi:   false,
		ImageType: tp,
	}
	f.ImageOptions(imageNameStr, x, y, w, h, flow, options, link, linkStr)
}

// ImageOptions puts a JPEG, PNG or GIF image in the current page. The size it
// will take on the page can be specified in different ways. If both w and h
// are 0, the image is rendered at 96 dpi. If either w or h is zero, it will be
// calculated from the other dimension so that the aspect ratio is maintained.
// If w and/or h are -1, the dpi for that dimension will be read from the
// ImageInfoType object. PNG files can contain dpi information, and if present,
// this information will be populated in the ImageInfoType object and used in
// Width, Height, and Extent calculations. Otherwise, the SetDpi function can
// be used to change the dpi from the default of 72.
//
// If w and h are any other negative value, their absolute values
// indicate their dpi extents.
//
// Supported JPEG formats are 24 bit, 32 bit and gray scale. Supported PNG
// formats are 24 bit, indexed color, and 8 bit indexed gray scale. If a GIF
// image is animated, only the first frame is rendered. Transparency is
// supported. It is possible to put a link on the image.
//
// imageNameStr may be the name of an image as registered with a call to either
// RegisterImageReader() or RegisterImage(). In the first case, the image is
// loaded using an io.Reader. This is generally useful when the image is
// obtained from some other means than as a disk-based file. In the second
// case, the image is loaded as a file. Alternatively, imageNameStr may
// directly specify a sufficiently qualified filename.
//
// However the image is loaded, if it is used more than once only one copy is
// embedded in the file.
//
// If x is negative, the current abscissa is used.
//
// If flow is true, the current y value is advanced after placing the image and
// a page break may be made if necessary.
//
// If link refers to an internal page anchor (that is, it is non-zero; see
// AddLink()), the image will be a clickable internal link. Otherwise, if
// linkStr specifies a URL, the image will be a clickable external link.
func (f *TinyPDF) ImageOptions(imageNameStr string, x, y, w, h float64, flow bool, options ImageOptions, link int, linkStr string) {
	if f.err != nil {
		return
	}
	info := f.RegisterImageOptions(imageNameStr, options)
	if f.err != nil {
		return
	}
	f.imageOut(info, x, y, w, h, options.AllowNegativePosition, flow, link, linkStr)
}

// RegisterImageReader registers an image, reading it from Reader r, adding it
// to the PDF file but not adding it to the page.
//
// This function is now deprecated in favor of RegisterImageOptionsReader
func (f *TinyPDF) RegisterImageReader(imgName, tp string, r io.Reader) (info *ImageInfoType) {
	options := ImageOptions{
		ReadDpi:   false,
		ImageType: tp,
	}
	return f.RegisterImageOptionsReader(imgName, options, r)
}

// ImageOptions provides a place to hang any options we want to use while
// parsing an image.
//
// ImageType's possible values are (case insensitive):
// "JPG", "JPEG", "PNG" and "GIF". If empty, the type is inferred from
// the file extension.
//
// ReadDpi defines whether to attempt to automatically read the image
// dpi information from the image file. Normally, this should be set
// to true (understanding that not all images will have this info
// available). However, for backwards compatibility with previous
// versions of the API, it defaults to false.
//
// AllowNegativePosition can be set to true in order to prevent the default
// coercion of negative x values to the current x position.
type ImageOptions struct {
	ImageType             string
	ReadDpi               bool
	AllowNegativePosition bool
}

// RegisterImageOptionsReader registers an image, reading it from Reader r, adding it
// to the PDF file but not adding it to the page. Use Image() with the same
// name to add the image to the page. Note that tp should be specified in this
// case.
//
// See Image() for restrictions on the image and the options parameters.
func (f *TinyPDF) RegisterImageOptionsReader(imgName string, options ImageOptions, r io.Reader) (info *ImageInfoType) {
	// Thanks, Ivan Daniluk, for generalizing this code to use the Reader interface.
	if f.err != nil {
		return
	}
	info, ok := f.images[imgName]
	if ok {
		return
	}

	// First use of this image, get info
	if options.ImageType == "" {
		f.err = Errf("image type should be specified if reading from custom reader")
		return
	}
	options.ImageType = Convert(options.ImageType).ToLower().String()
	if options.ImageType == "jpeg" {
		options.ImageType = "jpg"
	}
	switch options.ImageType {
	case "jpg":
		info = f.ParseJPG(r)
	case "png":
		info = f.ParsePNG(r, options.ReadDpi)
	case "gif":
		info = f.ParseGIF(r)
	default:
		f.err = Errf("unsupported image type: %s", options.ImageType)
	}
	if f.err != nil {
		return
	}

	if info.i, f.err = generateImageID(info); f.err != nil {
		return
	}
	f.images[imgName] = info

	return
}

// RegisterImage registers an image, adding it to the PDF file but not adding
// it to the page. Use Image() with the same filename to add the image to the
// page. Note that Image() calls this function, so this function is only
// necessary if you need information about the image before placing it.
//
// This function is now deprecated in favor of RegisterImageOptions.
// See Image() for restrictions on the image and the "tp" parameters.
func (f *TinyPDF) RegisterImage(fileStr, tp string) (info *ImageInfoType) {
	options := ImageOptions{
		ReadDpi:   false,
		ImageType: tp,
	}
	return f.RegisterImageOptions(fileStr, options)
}

// RegisterImageOptions registers an image, adding it to the PDF file but not
// adding it to the page. Use Image() with the same filename to add the image
// to the page. Note that Image() calls this function, so this function is only
// necessary if you need information about the image before placing it. See
// Image() for restrictions on the image and the "tp" parameters.
func (f *TinyPDF) RegisterImageOptions(fileStr string, options ImageOptions) (info *ImageInfoType) {
	info, ok := f.images[fileStr]
	if ok {
		return
	}

	file, err := os.Open(fileStr)
	if err != nil {
		f.err = err
		return
	}
	defer file.Close()

	// First use of this image, get info
	if options.ImageType == "" {
		pos := LastIndex(fileStr, ".")
		if pos < 0 {
			f.err = Errf("image file has no extension and no type was specified: %s", fileStr)
			return
		}
		options.ImageType = fileStr[pos+1:]
	}

	return f.RegisterImageOptionsReader(fileStr, options, file)
}

// GetImageInfo returns information about the registered image specified by
// imageStr. If the image has not been registered, nil is returned. The
// internal error is not modified by this method.
func (f *TinyPDF) GetImageInfo(imageStr string) (info *ImageInfoType) {
	return f.images[imageStr]
}

// ImportObjects imports objects from gofpdi into current document
func (f *TinyPDF) ImportObjects(objs map[string][]byte) {
	// imported objects support removed along with templates
}

// ImportObjPos imports object hash positions from gofpdi
func (f *TinyPDF) ImportObjPos(objPos map[string]map[int]string) {
	// imported objects support removed along with templates
}

// putImportedTemplates writes the imported template objects to the PDF

// GetConversionRatio returns the conversion ratio based on the Unit given when
// creating the PDF.
func (f *TinyPDF) GetConversionRatio() float64 {
	return f.k
}

// SetHomeXY is a convenience method that sets the current position to the left
// and top margins.
func (f *TinyPDF) SetHomeXY() {
	f.SetY(f.tMargin)
	f.SetX(f.lMargin)
}

// Escape special characters in strings
func (f *TinyPDF) escape(s string) string {
	// Usar tinystring para reemplazos encadenados
	s = Convert(s).Replace("\\", "\\\\").Replace("(", "\\(").Replace(")", "\\)").Replace("\r", "\\r").String()
	return s
}

// textstring formats a text string
func (f *TinyPDF) textstring(s string) string {
	if f.protect.encrypted {
		b := []byte(s)
		f.protect.rc4(uint32(f.n), &b)
		s = string(b)
	}
	return "(" + f.escape(s) + ")"
}

// UseImportedTemplate draws an imported PDF page (registered via ImportTemplates)
// onto the current page. tplName should be the resource name provided by the
// importer (for example "/TPL1").
func (f *TinyPDF) UseImportedTemplate(tplName string, scaleX float64, scaleY float64, tX float64, tY float64) {
	// Draw the imported XObject directly. Scaling and translation are applied
	// in PDF user space. Keep the same operator sequence used previously.
	f.outf("q 0 J 1 w 0 j 0 G 0 g q %.4F 0 0 %.4F %.4F %.4F cm %s Do Q Q\n", scaleX*f.k, scaleY*f.k, tX*f.k, (tY+f.h)*f.k, tplName)
}

// ImportTemplates registers a mapping of template resource names to imported
// template identifiers coming from a PDF importer (gofpdi). The importer will
// later call UseImportedTemplate to draw the pages.
func (f *TinyPDF) ImportTemplates(tpls map[string]string) {
	if f.importedTplObjs == nil {
		f.importedTplObjs = make(map[string]string)
	}
	for tplName, tplID := range tpls {
		f.importedTplObjs[tplName] = tplID
	}
}

func blankCount(str string) (count int) {
	l := len(str)
	for j := 0; j < l; j++ {
		if byte(' ') == str[j] {
			count++
		}
	}
	return
}

// GetUnderlineThickness returns the current text underline thickness multiplier.
func (f *TinyPDF) GetUnderlineThickness() float64 {
	return f.userUnderlineThickness
}

// SetUnderlineThickness accepts a multiplier for adjusting the text underline
// thickness, defaulting to 1. See SetUnderlineThickness example.
func (f *TinyPDF) SetUnderlineThickness(thickness float64) {
	f.userUnderlineThickness = thickness
}

// Underline text
func (f *TinyPDF) dounderline(x, y float64, txt string) string {
	up := float64(f.currentFont.Up)
	ut := float64(f.currentFont.Ut) * f.userUnderlineThickness
	w := f.GetStringWidth(txt) + f.ws*float64(blankCount(txt))
	return Fmt("%.2f %.2f %.2f %.2f re f", x*f.k,
		(f.h-(y-up/1000*f.fontSize))*f.k, w*f.k, -ut/1000*f.GetFontSizePt())
}

func (f *TinyPDF) dostrikeout(x, y float64, txt string) string {
	up := float64(f.currentFont.Up)
	ut := float64(f.currentFont.Ut)
	w := f.GetStringWidth(txt) + f.ws*float64(blankCount(txt))
	return Fmt("%.2f %.2f %.2f %.2f re f", x*f.k,
		(f.h-(y+4*up/1000*f.fontSize))*f.k, w*f.k, -ut/1000*f.GetFontSizePt())
}

func (f *TinyPDF) newImageInfo() *ImageInfoType {
	// default dpi to 72 unless told otherwise
	return &ImageInfoType{scale: f.k, dpi: 72}
}

// putTemplates is a no-op since template support was removed. It remains to
// provide compatibility with code paths that call it during resource emission.
func (f *TinyPDF) putTemplates() {
	// intentionally empty
}

// ParseJPG extracts info from io.Reader with JPEG data
// Thank you, Bruno Michel, for providing this code.
func (f *TinyPDF) ParseJPG(r io.Reader) (info *ImageInfoType) {
	info = f.newImageInfo()
	var (
		data bytes.Buffer
		err  error
	)
	_, err = data.ReadFrom(r)
	if err != nil {
		f.err = err
		return
	}
	info.data = data.Bytes()

	config, err := jpeg.DecodeConfig(bytes.NewReader(info.data))
	if err != nil {
		f.err = err
		return
	}
	info.w = float64(config.Width)
	info.h = float64(config.Height)
	info.f = "DCTDecode"
	info.bpc = 8
	switch config.ColorModel {
	case color.GrayModel:
		info.cs = "DeviceGray"
	case color.YCbCrModel:
		info.cs = "DeviceRGB"
	case color.CMYKModel:
		info.cs = "DeviceCMYK"
	default:
		f.err = Errf("image JPEG buffer has unsupported color space (%v)", config.ColorModel)
		return
	}
	return
}

// ParsePNG extracts info from a PNG data
func (f *TinyPDF) ParsePNG(r io.Reader, readdpi bool) (info *ImageInfoType) {
	buf, err := newRBuffer(r)
	if err != nil {
		f.err = err
		return
	}
	return f.parsepngstream(buf, readdpi)
}

// ParseGIF extracts info from a GIF data (via PNG conversion)
func (f *TinyPDF) ParseGIF(r io.Reader) (info *ImageInfoType) {
	data, err := newRBuffer(r)
	if err != nil {
		f.err = err
		return
	}
	var img image.Image
	img, err = gif.Decode(data)
	if err != nil {
		f.err = err
		return
	}
	pngBuf := new(bytes.Buffer)
	err = png.Encode(pngBuf, img)
	if err != nil {
		f.err = err
		return
	}
	return f.parsepngstream(&rbuffer{p: pngBuf.Bytes()}, false)
}

// newobj begins a new object
func (f *TinyPDF) newobj() {
	// dbg("newobj")
	f.n++
	for j := len(f.offsets); j <= f.n; j++ {
		f.offsets = append(f.offsets, 0)
	}
	f.offsets[f.n] = f.buffer.Len()
	f.outf("%d 0 obj", f.n)
}

func (f *TinyPDF) putstream(b []byte) {
	// dbg("putstream")
	if f.protect.encrypted {
		f.protect.rc4(uint32(f.n), &b)
	}
	f.out("stream")
	f.out(string(b))
	f.out("endstream")
}

// outf adds a formatted line to the document
func (f *TinyPDF) outf(fmtStr string, args ...any) {
	f.out(Fmt(fmtStr, args...))
}

// out; Add a line to the document
func (f *TinyPDF) out(s string) {
	if f.state == 2 {
		must(f.pages[f.page].WriteString(s))
		must(f.pages[f.page].WriteString("\n"))
	} else {
		must(f.buffer.WriteString(s))
		must(f.buffer.WriteString("\n"))
	}
}

func (f *TinyPDF) put(s string) {
	if f.state == 2 {
		f.pages[f.page].WriteString(s)
	} else {
		f.buffer.WriteString(s)
	}
}

// outbuf adds a buffered line to the document
func (f *TinyPDF) outbuf(r io.Reader) {
	if f.state == 2 {
		must64(f.pages[f.page].ReadFrom(r))
		must(f.pages[f.page].WriteString("\n"))
	} else {
		must64(f.buffer.ReadFrom(r))
		must(f.buffer.WriteString("\n"))
	}
}

// RawWriteStr writes a string directly to the PDF generation buffer. This is a
// low-level function that is not required for normal PDF construction. An
// understanding of the PDF specification is needed to use this method
// correctly.
func (f *TinyPDF) RawWriteStr(str string) {
	f.out(str)
}

// RawWriteBuf writes the contents of the specified buffer directly to the PDF
// generation buffer. This is a low-level function that is not required for
// normal PDF construction. An understanding of the PDF specification is needed
// to use this method correctly.
func (f *TinyPDF) RawWriteBuf(r io.Reader) {
	f.outbuf(r)
}

func (f *TinyPDF) putF64(v float64, prec int) {
	f.put(f.fmtF64(v, prec))
}

// fmtF64 converts the floating-point number f to a string with precision prec.
func (f *TinyPDF) fmtF64(v float64, prec int) string {
	// Usar tinystring para formatear float con precisión
	return Convert(v).Round(prec).String()
}

func (f *TinyPDF) putInt(v int) {
	f.put(f.fmtInt(v))
}

func (f *TinyPDF) fmtInt(v int) string {
	// Usar tinystring para convertir int a string
	return Convert(v).String()
}

// SetDefaultCatalogSort sets the default value of the catalog sort flag that
// will be used when initializing a new TinyPDF instance. See SetCatalogSort() for
// more details.
func SetDefaultCatalogSort(flag bool) {
	gl.catalogSort = flag
}

// GetCatalogSort returns the document's internal catalog sort flag.
func (f *TinyPDF) GetCatalogSort() bool {
	return f.catalogSort
}

// SetCatalogSort sets a flag that will be used, if true, to consistently order
// the document's internal resource catalogs. This method is typically only
// used for test purposes to facilitate PDF comparison.
func (f *TinyPDF) SetCatalogSort(flag bool) {
	f.catalogSort = flag
}

// RegisterAlias adds an (alias, replacement) pair to the document so we can
// replace all occurrences of that alias after writing but before the document
// is closed. Functions ExampleFpdf_RegisterAlias() and
// ExampleFpdf_RegisterAlias_utf8() in fpdf_test.go demonstrate this method.
func (f *TinyPDF) RegisterAlias(alias, replacement string) {
	// Note: map[string]string assignments embed literal escape ("\00") sequences
	// into utf16 key and value  Consequently, subsequent search/replace
	// operations will fail unexpectedly if utf8toutf16() conversions take place
	// here. Instead, conversions are deferred until the actual search/replace
	// operation takes place when the PDF output is generated.
	f.aliasMap[alias] = replacement
}

func (f *TinyPDF) replaceAliases() {
	for mode := 0; mode < 2; mode++ {
		for alias, replacement := range f.aliasMap {
			if mode == 1 {
				alias = utf8toutf16(alias, false)
				replacement = utf8toutf16(replacement, false)
			}
			for n := 1; n <= f.page; n++ {
				s := f.pages[n].String()
				if Contains(s, alias) {
					s = Convert(s).Replace(alias, replacement).String()
					f.pages[n].Truncate(0)
					f.pages[n].WriteString(s)
				}
			}
		}
	}
}
