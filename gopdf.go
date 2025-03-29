package tinypdf

import (
	// NO tinygo supported
	"bytes"
	"compress/zlib" // for constants tinygo OK

	// "fmt" // NO tinygo supported

	"io"

	"math"    // tinygo OK
	"os"      // tinygo OK
	"strconv" // tinygo OK
	"time"    // NO tinygo supported
)

const subsetFont = "SubsetFont"

// the default margin if no margins are set
const defaultMargin = 10.0 //for backward compatible

type DrawableRectOptions struct {
	Rect
	X            float64
	Y            float64
	PaintStyle   PaintStyle
	Transparency *Transparency

	extGStateIndexes []int
}

type CropOptions struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

type ImageOptions struct {
	DegreeAngle    float64
	VerticalFlip   bool
	HorizontalFlip bool
	X              float64
	Y              float64
	Rect           *Rect
	Mask           *MaskOptions
	Crop           *CropOptions
	Transparency   *Transparency

	extGStateIndexes []int
}
type ImageFromOption struct {
	Format string //jpeg,png
	X      float64
	Y      float64
	Rect   *Rect
}

type MaskOptions struct {
	ImageOptions
	BBox   *[4]float64
	Holder ImageHolder
}

type lineOptions struct {
	extGStateIndexes []int
}

type polygonOptions struct {
	extGStateIndexes []int
}

// SetLineWidth : set line width
func (gp *GoPdf) SetLineWidth(width float64) {
	gp.curr.lineWidth = gp.UnitsToPoints(width)
	gp.getContent().AppendStreamSetLineWidth(gp.UnitsToPoints(width))
}

// SetCompressLevel : set compress Level for content streams
// Possible values for level:
//
//	-2 HuffmanOnly, -1 DefaultCompression (which is level 6)
//	 0 No compression,
//	 1 fastest compression, but not very good ratio
//	 9 best compression, but slowest
func (gp *GoPdf) SetCompressLevel(level int) {
	if level < -2 { //-2 = zlib.HuffmanOnly
		io.WriteString(os.Stderr, "compress level too small, using DefaultCompression instead\n")
		level = zlib.DefaultCompression
	} else if level > zlib.BestCompression {
		io.WriteString(os.Stderr, "compress level too big, using BestCompression instead\n")
		level = zlib.BestCompression
		return
	}
	// sanity check complete
	gp.compressLevel = level
}

// SetNoCompression : compressLevel = 0
func (gp *GoPdf) SetNoCompression() {
	gp.compressLevel = zlib.NoCompression
}

// SetLineType : set line type  ("dashed" ,"dotted")
//
//	Usage:
//	pdf.SetLineType("dashed")
//	pdf.Line(50, 200, 550, 200)
//	pdf.SetLineType("dotted")
//	pdf.Line(50, 400, 550, 400)
func (gp *GoPdf) SetLineType(linetype string) {
	gp.getContent().AppendStreamSetLineType(linetype)
}

// SetCustomLineType : set custom line type
//
//	Usage:
//	pdf.SetCustomLineType([]float64{0.8, 0.8}, 0)
//	pdf.Line(50, 200, 550, 200)
func (gp *GoPdf) SetCustomLineType(dashArray []float64, dashPhase float64) {
	for i := range dashArray {
		gp.UnitsToPointsVar(&dashArray[i])
	}
	gp.UnitsToPointsVar(&dashPhase)
	gp.getContent().AppendStreamSetCustomLineType(dashArray, dashPhase)
}

// Line : draw line
//
//	Usage:
//	pdf.SetTransparency(tinypdf.Transparency{Alpha: 0.5,BlendModeType: tinypdf.ColorBurn})
//	pdf.SetLineType("dotted")
//	pdf.SetStrokeColor(255, 0, 0)
//	pdf.SetLineWidth(2)
//	pdf.Line(10, 30, 585, 30)
//	pdf.ClearTransparency()
func (gp *GoPdf) Line(x1 float64, y1 float64, x2 float64, y2 float64) {
	gp.UnitsToPointsVar(&x1, &y1, &x2, &y2)
	transparency, err := gp.getCachedTransparency(nil)
	if err != nil {
		transparency = nil
	}
	var opts = lineOptions{}
	if transparency != nil {
		opts.extGStateIndexes = append(opts.extGStateIndexes, transparency.extGStateIndex)
	}
	gp.getContent().AppendStreamLine(x1, y1, x2, y2, opts)
}

// RectFromLowerLeft : draw rectangle from lower-left corner (x, y)
func (gp *GoPdf) RectFromLowerLeft(x float64, y float64, wdth float64, hght float64) {
	gp.UnitsToPointsVar(&x, &y, &wdth, &hght)

	opts := DrawableRectOptions{
		X:          x,
		Y:          y,
		PaintStyle: DrawPaintStyle,
		Rect:       Rect{W: wdth, H: hght},
	}

	gp.getContent().AppendStreamRectangle(opts)
}

// RectFromUpperLeft : draw rectangle from upper-left corner (x, y)
func (gp *GoPdf) RectFromUpperLeft(x float64, y float64, wdth float64, hght float64) {
	gp.UnitsToPointsVar(&x, &y, &wdth, &hght)

	opts := DrawableRectOptions{
		X:          x,
		Y:          y + hght,
		PaintStyle: DrawPaintStyle,
		Rect:       Rect{W: wdth, H: hght},
	}

	gp.getContent().AppendStreamRectangle(opts)
}

// RectFromLowerLeftWithStyle : draw rectangle from lower-left corner (x, y)
//   - style: Style of rectangule (draw and/or fill: D, F, DF, FD)
//     D or empty string: draw. This is the default value.
//     F: fill
//     DF or FD: draw and fill
func (gp *GoPdf) RectFromLowerLeftWithStyle(x float64, y float64, wdth float64, hght float64, style string) {
	opts := DrawableRectOptions{
		X: x,
		Y: y,
		Rect: Rect{
			H: hght,
			W: wdth,
		},
		PaintStyle: parseStyle(style),
	}
	gp.RectFromLowerLeftWithOpts(opts)
}

func (gp *GoPdf) RectFromLowerLeftWithOpts(opts DrawableRectOptions) error {
	gp.UnitsToPointsVar(&opts.X, &opts.Y, &opts.W, &opts.H)

	imageTransparency, err := gp.getCachedTransparency(opts.Transparency)
	if err != nil {
		return err
	}

	if imageTransparency != nil {
		opts.extGStateIndexes = append(opts.extGStateIndexes, imageTransparency.extGStateIndex)
	}

	gp.getContent().AppendStreamRectangle(opts)

	return nil
}

// RectFromUpperLeftWithStyle : draw rectangle from upper-left corner (x, y)
//   - style: Style of rectangule (draw and/or fill: D, F, DF, FD)
//     D or empty string: draw. This is the default value.
//     F: fill
//     DF or FD: draw and fill
func (gp *GoPdf) RectFromUpperLeftWithStyle(x float64, y float64, wdth float64, hght float64, style string) {
	opts := DrawableRectOptions{
		X: x,
		Y: y,
		Rect: Rect{
			H: hght,
			W: wdth,
		},
		PaintStyle: parseStyle(style),
	}
	gp.RectFromUpperLeftWithOpts(opts)
}

func (gp *GoPdf) RectFromUpperLeftWithOpts(opts DrawableRectOptions) error {
	gp.UnitsToPointsVar(&opts.X, &opts.Y, &opts.W, &opts.H)

	opts.Y += opts.H

	imageTransparency, err := gp.getCachedTransparency(opts.Transparency)
	if err != nil {
		return err
	}

	if imageTransparency != nil {
		opts.extGStateIndexes = append(opts.extGStateIndexes, imageTransparency.extGStateIndex)
	}

	gp.getContent().AppendStreamRectangle(opts)

	return nil
}

// Oval : draw oval
func (gp *GoPdf) Oval(x1 float64, y1 float64, x2 float64, y2 float64) {
	gp.UnitsToPointsVar(&x1, &y1, &x2, &y2)
	gp.getContent().AppendStreamOval(x1, y1, x2, y2)
}

// Br : new line
func (gp *GoPdf) Br(h float64) {
	gp.UnitsToPointsVar(&h)
	gp.curr.Y += h
	gp.curr.X = gp.margins.Left
}

// SetGrayFill set the grayscale for the fill, takes a float64 between 0.0 and 1.0
func (gp *GoPdf) SetGrayFill(grayScale float64) {
	gp.curr.txtColorMode = "gray"
	gp.curr.grayFill = grayScale
	gp.getContent().AppendStreamSetGrayFill(grayScale)
}

// SetGrayStroke set the grayscale for the stroke, takes a float64 between 0.0 and 1.0
func (gp *GoPdf) SetGrayStroke(grayScale float64) {
	gp.curr.grayStroke = grayScale
	gp.getContent().AppendStreamSetGrayStroke(grayScale)
}

func (gp *GoPdf) AddOutline(title string) {
	gp.outlines.AddOutline(gp.curr.IndexOfPageObj+1, title)
}

// AddOutlineWithPosition add an outline with position
func (gp *GoPdf) AddOutlineWithPosition(title string) *OutlineObj {
	return gp.outlines.AddOutlinesWithPosition(gp.curr.IndexOfPageObj+1, title, gp.config.PageSize.H-gp.curr.Y+20)
}

// Start : init gopdf
func (gp *GoPdf) Start(config Config) {

	gp.start(config)

}

func (gp *GoPdf) StartWithImporter(config Config, importer *importer) {

	gp.start(config, importer)

}

func (gp *GoPdf) start(config Config, importer ...*importer) {

	gp.config = config
	gp.init(importer...)
	//init all basic obj
	catalog := new(CatalogObj)
	catalog.init(func() *GoPdf {
		return gp
	})
	pages := new(PagesObj)
	pages.init(func() *GoPdf {
		return gp
	})
	gp.outlines = new(OutlinesObj)
	gp.outlines.init(func() *GoPdf {
		return gp
	})
	gp.indexOfCatalogObj = gp.addObj(catalog)
	gp.indexOfPagesObj = gp.addObj(pages)
	gp.indexOfOutlinesObj = gp.addObj(gp.outlines)
	gp.outlines.SetIndexObjOutlines(gp.indexOfOutlinesObj)

	//indexOfProcSet
	procset := new(ProcSetObj)
	procset.init(func() *GoPdf {
		return gp
	})
	gp.indexOfProcSet = gp.addObj(procset)

	if gp.isUseProtection() {
		gp.pdfProtection = gp.createProtection()
	}

	gp.placeHolderTexts = make(map[string][]placeHolderTextInfo)

}

// ImportPage imports a page and return template id.
// gofpdi code
func (gp *GoPdf) ImportPage(sourceFile string, pageno int, box string) int {
	// Set source file for fpdi
	gp.fpdi.SetSourceFile(sourceFile)

	// gofpdi needs to know where to start the object id at.
	// By default, it starts at 1, but gopdf adds a few objects initially.
	startObjID := gp.GetNextObjectID()

	// Set gofpdi next object ID to  whatever the value of startObjID is
	gp.fpdi.SetNextObjectID(startObjID)

	// Import page
	tpl := gp.fpdi.ImportPage(pageno, box)

	// Import objects into current pdf document
	tplObjIDs := gp.fpdi.PutFormXobjects()

	// Set template names and ids in gopdf
	gp.ImportTemplates(tplObjIDs)

	// Get a map[int]string of the imported objects.
	// The map keys will be the ID of each object.
	imported := gp.fpdi.GetImportedObjects()

	// Import gofpdi objects into gopdf, starting at whatever the value of startObjID is
	gp.ImportObjects(imported, startObjID)

	// Return template ID
	return tpl
}

// ImportPageStream imports page using a stream.
// Return template id after importing.
// gofpdi code
func (gp *GoPdf) ImportPageStream(sourceStream *io.ReadSeeker, pageno int, box string) int {
	// Set source file for fpdi
	gp.fpdi.SetSourceStream(sourceStream)

	// gofpdi needs to know where to start the object id at.
	// By default, it starts at 1, but gopdf adds a few objects initially.
	startObjID := gp.GetNextObjectID()

	// Set gofpdi next object ID to  whatever the value of startObjID is
	gp.fpdi.SetNextObjectID(startObjID)

	// Import page
	tpl := gp.fpdi.ImportPage(pageno, box)

	// Import objects into current pdf document
	tplObjIDs := gp.fpdi.PutFormXobjects()

	// Set template names and ids in gopdf
	gp.ImportTemplates(tplObjIDs)

	// Get a map[int]string of the imported objects.
	// The map keys will be the ID of each object.
	imported := gp.fpdi.GetImportedObjects()

	// Import gofpdi objects into gopdf, starting at whatever the value of startObjID is
	gp.ImportObjects(imported, startObjID)

	// Return template ID
	return tpl
}

// UseImportedTemplate draws an imported PDF page.
func (gp *GoPdf) UseImportedTemplate(tplid int, x float64, y float64, w float64, h float64) {
	gp.UnitsToPointsVar(&x, &y, &w, &h)
	// Get template values to draw
	tplName, scaleX, scaleY, tX, tY := gp.fpdi.UseTemplate(tplid, x, y, w, h)
	gp.getContent().AppendStreamImportedTemplate(tplName, scaleX, scaleY, tX, tY)
}

// ImportPagesFromSource imports pages from a source pdf.
// The source can be a file path, byte slice, or (*)io.ReadSeeker.
func (gp *GoPdf) ImportPagesFromSource(source interface{}, box string) error {
	switch v := source.(type) {
	case string:
		// Set source file for fpdi
		gp.fpdi.SetSourceFile(v)
	case []byte:
		// Set source stream for fpdi
		rs := io.ReadSeeker(bytes.NewReader(v))
		gp.fpdi.SetSourceStream(&rs)
	case io.ReadSeeker:
		// Set source stream for fpdi
		gp.fpdi.SetSourceStream(&v)
	case *io.ReadSeeker:
		// Set source stream for fpdi
		gp.fpdi.SetSourceStream(v)
	default:
		return newErr("source type not supported")
	}

	// Get number of pages from source file
	pages := gp.fpdi.GetNumPages()

	// Get page sizes from source file
	sizes := gp.fpdi.GetPageSizes()

	for i := 0; i < pages; i++ {
		pageno := i + 1

		// Get the size of the page
		size, ok := sizes[pageno][box]
		if !ok {
			return newErr("can not get page size")
		}

		// Add a new page to the document
		gp.AddPage()

		// gofpdi needs to know where to start the object id at.
		// By default, it starts at 1, but gopdf adds a few objects initially.
		startObjID := gp.GetNextObjectID()

		// Set gofpdi next object ID to  whatever the value of startObjID is
		gp.fpdi.SetNextObjectID(startObjID)

		// Import page
		tpl := gp.fpdi.ImportPage(pageno, box)

		// Import objects into current pdf document
		tplObjIDs := gp.fpdi.PutFormXobjects()

		// Set template names and ids in gopdf
		gp.ImportTemplates(tplObjIDs)

		// Get a map[int]string of the imported objects.
		// The map keys will be the ID of each object.
		imported := gp.fpdi.GetImportedObjects()

		// Import gofpdi objects into gopdf, starting at whatever the value of startObjID is
		gp.ImportObjects(imported, startObjID)

		// Draws the imported template on the current page
		gp.UseImportedTemplate(tpl, 0, 0, size["w"], size["h"])
	}

	return nil
}

// GetNextObjectID gets the next object ID so that gofpdi knows where to start the object IDs.
func (gp *GoPdf) GetNextObjectID() int {
	return len(gp.pdfObjs) + 1
}

// ImportObjects imports objects from gofpdi into current document.
func (gp *GoPdf) ImportObjects(objs map[int]string, startObjID int) {
	for i := startObjID; i < len(objs)+startObjID; i++ {
		if objs[i] != "" {
			gp.addObj(&ImportedObj{Data: objs[i]})
		}
	}
}

// ImportTemplates names into procset dictionary.
func (gp *GoPdf) ImportTemplates(tpls map[string]int) {
	procset := gp.pdfObjs[gp.indexOfProcSet].(*ProcSetObj)
	for tplName, tplID := range tpls {
		procset.ImportedTemplateIds[tplName] = tplID
	}
}

// AddExternalLink adds a new external link.
func (gp *GoPdf) AddExternalLink(url string, x, y, w, h float64) {
	gp.UnitsToPointsVar(&x, &y, &w, &h)

	linkOpt := linkOption{x, gp.config.PageSize.H - y, w, h, url, ""}
	gp.addLink(linkOpt)
}

// AddInternalLink adds a new internal link.
func (gp *GoPdf) AddInternalLink(anchor string, x, y, w, h float64) {
	gp.UnitsToPointsVar(&x, &y, &w, &h)

	linkOpt := linkOption{x, gp.config.PageSize.H - y, w, h, "", anchor}
	gp.addLink(linkOpt)
}

func (gp *GoPdf) addLink(option linkOption) {
	page := gp.pdfObjs[gp.curr.IndexOfPageObj].(*PageObj)
	linkObj := gp.addObj(annotObj{option, func() *GoPdf {
		return gp
	}})
	page.LinkObjIds = append(page.LinkObjIds, linkObj+1)
}

// SetAnchor creates a new anchor.
func (gp *GoPdf) SetAnchor(name string) {
	y := gp.config.PageSize.H - gp.curr.Y + float64(gp.curr.FontSize)
	gp.anchors[name] = anchorOption{gp.curr.IndexOfPageObj, y}
}

// AddTTFFontByReader adds font data by reader.
func (gp *GoPdf) AddTTFFontData(family string, fontData []byte) error {
	return gp.AddTTFFontDataWithOption(family, fontData, defaultTtfFontOption())
}

// AddTTFFontDataWithOption adds font data with option.
func (gp *GoPdf) AddTTFFontDataWithOption(family string, fontData []byte, option TtfOption) error {
	subsetFont := new(SubsetFontObj)
	subsetFont.init(func() *GoPdf {
		return gp
	})
	subsetFont.SetTtfFontOption(option)
	subsetFont.SetFamily(family)
	err := subsetFont.SetTTFData(fontData)
	if err != nil {
		return err
	}

	return gp.setSubsetFontObject(subsetFont, family, option)
}

// AddTTFFontByReader adds font file by reader.
func (gp *GoPdf) AddTTFFontByReader(family string, rd io.Reader) error {
	return gp.AddTTFFontByReaderWithOption(family, rd, defaultTtfFontOption())
}

// AddTTFFontByReaderWithOption adds font file by reader with option.
func (gp *GoPdf) AddTTFFontByReaderWithOption(family string, rd io.Reader, option TtfOption) error {
	subsetFont := new(SubsetFontObj)
	subsetFont.init(func() *GoPdf {
		return gp
	})
	subsetFont.SetTtfFontOption(option)
	subsetFont.SetFamily(family)
	err := subsetFont.SetTTFByReader(rd)
	if err != nil {
		return err
	}

	return gp.setSubsetFontObject(subsetFont, family, option)
}

// setSubsetFontObject sets SubsetFontObj.
// The given SubsetFontObj is expected to be configured in advance.
func (gp *GoPdf) setSubsetFontObject(subsetFont *SubsetFontObj, family string, option TtfOption) error {
	unicodemap := new(UnicodeMap)
	unicodemap.init(func() *GoPdf {
		return gp
	})
	unicodemap.setProtection(gp.protection())
	unicodemap.SetPtrToSubsetFontObj(subsetFont)
	unicodeindex := gp.addObj(unicodemap)

	pdfdic := new(PdfDictionaryObj)
	pdfdic.init(func() *GoPdf {
		return gp
	})
	pdfdic.setProtection(gp.protection())
	pdfdic.SetPtrToSubsetFontObj(subsetFont)
	pdfdicindex := gp.addObj(pdfdic)

	subfontdesc := new(SubfontDescriptorObj)
	subfontdesc.init(func() *GoPdf {
		return gp
	})
	subfontdesc.SetPtrToSubsetFontObj(subsetFont)
	subfontdesc.SetIndexObjPdfDictionary(pdfdicindex)
	subfontdescindex := gp.addObj(subfontdesc)

	cidfont := new(CIDFontObj)
	cidfont.init(func() *GoPdf {
		return gp
	})
	cidfont.SetPtrToSubsetFontObj(subsetFont)
	cidfont.SetIndexObjSubfontDescriptor(subfontdescindex)
	cidindex := gp.addObj(cidfont)

	subsetFont.SetIndexObjCIDFont(cidindex)
	subsetFont.SetIndexObjUnicodeMap(unicodeindex)
	index := gp.addObj(subsetFont) //add หลังสุด

	if gp.indexOfProcSet != -1 {
		procset := gp.pdfObjs[gp.indexOfProcSet].(*ProcSetObj)
		if !procset.Relates.IsContainsFamilyAndStyle(family, option.Style&^Underline) {
			procset.Relates = append(procset.Relates, RelateFont{Family: family, IndexOfObj: index, CountOfFont: gp.curr.CountOfFont, Style: option.Style &^ Underline})
			subsetFont.CountOfFont = gp.curr.CountOfFont
			gp.curr.CountOfFont++
		}
	}
	return nil
}

// AddTTFFontWithOption : add font file
func (gp *GoPdf) AddTTFFontWithOption(family string, ttfpath string, option TtfOption) error {

	if _, err := os.Stat(ttfpath); os.IsNotExist(err) {
		return err
	}
	data, err := os.ReadFile(ttfpath)
	if err != nil {
		return err
	}
	rd := bytes.NewReader(data)
	return gp.AddTTFFontByReaderWithOption(family, rd, option)
}

// AddTTFFont : add font file
func (gp *GoPdf) AddTTFFont(family string, ttfpath string) error {
	return gp.AddTTFFontWithOption(family, ttfpath, defaultTtfFontOption())
}

// KernOverride override kern value
func (gp *GoPdf) KernOverride(family string, fn FuncKernOverride) error {
	i := 0
	max := len(gp.pdfObjs)
	for i < max {
		if gp.pdfObjs[i].getType() == subsetFont {
			obj := gp.pdfObjs[i]
			sub, ok := obj.(*SubsetFontObj)
			if ok {
				if sub.GetFamily() == family {
					sub.funcKernOverride = fn
					return nil
				}
			}
		}
		i++
	}
	return errMissingFontFamily
}

// SetTextColor :  function sets the text color
func (gp *GoPdf) SetTextColor(r uint8, g uint8, b uint8) {
	gp.curr.txtColorMode = "color"
	rgb := cacheContentTextColorRGB{
		r: r,
		g: g,
		b: b,
	}
	gp.curr.setTextColor(rgb)
}

func (gp *GoPdf) SetTextColorCMYK(c, m, y, k uint8) {
	gp.curr.txtColorMode = "color"
	cmyk := cacheContentTextColorCMYK{
		c: c,
		m: m,
		y: y,
		k: k,
	}
	gp.curr.setTextColor(cmyk)
}

// SetStrokeColor set the color for the stroke
func (gp *GoPdf) SetStrokeColor(r uint8, g uint8, b uint8) {
	gp.getContent().AppendStreamSetColorStroke(r, g, b)
}

// SetFillColor set the color for the stroke
func (gp *GoPdf) SetFillColor(r uint8, g uint8, b uint8) {
	gp.getContent().AppendStreamSetColorFill(r, g, b)
}

// SetStrokeColorCMYK set the color for the stroke in CMYK color mode
func (gp *GoPdf) SetStrokeColorCMYK(c, m, y, k uint8) {
	gp.getContent().AppendStreamSetColorStrokeCMYK(c, m, y, k)
}

// SetFillColorCMYK set the color for the fill in CMYK color mode
func (gp *GoPdf) SetFillColorCMYK(c, m, y, k uint8) {
	gp.getContent().AppendStreamSetColorFillCMYK(c, m, y, k)
}

// MeasureTextWidth : measure Width of text (use current font)
func (gp *GoPdf) MeasureTextWidth(text string) (float64, error) {

	text, err := gp.curr.FontISubset.AddChars(text) //AddChars for create CharacterToGlyphIndex
	if err != nil {
		return 0, err
	}

	_, _, textWidthPdfUnit, err := createContent(gp.curr.FontISubset, text, gp.curr.FontSize, gp.curr.CharSpacing, nil)
	if err != nil {
		return 0, err
	}
	return pointsToUnits(gp.config, textWidthPdfUnit), nil
}

// MeasureCellHeightByText : measure Height of cell by text (use current font)
func (gp *GoPdf) MeasureCellHeightByText(text string) (float64, error) {

	text, err := gp.curr.FontISubset.AddChars(text) //AddChars for create CharacterToGlyphIndex
	if err != nil {
		return 0, err
	}

	_, cellHeightPdfUnit, _, err := createContent(gp.curr.FontISubset, text, gp.curr.FontSize, gp.curr.CharSpacing, nil)
	if err != nil {
		return 0, err
	}
	return pointsToUnits(gp.config, cellHeightPdfUnit), nil
}

// Curve Draws a Bézier curve (the Bézier curve is tangent to the line between the control points at either end of the curve)
// Parameters:
// - x0, y0: Start point
// - x1, y1: Control point 1
// - x2, y2: Control point 2
// - x3, y3: End point
// - style: Style of rectangule (draw and/or fill: D, F, DF, FD)
func (gp *GoPdf) Curve(x0 float64, y0 float64, x1 float64, y1 float64, x2 float64, y2 float64, x3 float64, y3 float64, style string) {
	gp.UnitsToPointsVar(&x0, &y0, &x1, &y1, &x2, &y2, &x3, &y3)
	gp.getContent().AppendStreamCurve(x0, y0, x1, y1, x2, y2, x3, y3, style)
}

/*
//SetProtection set permissions as well as user and owner passwords
func (gp *GoPdf) SetProtection(permissions int, userPass []byte, ownerPass []byte) {
	gp.pdfProtection = new(PDFProtection)
	gp.pdfProtection.setProtection(permissions, userPass, ownerPass)
}*/

// SetInfo set Document Information Dictionary
func (gp *GoPdf) SetInfo(info PdfInfo) {
	gp.info = &info
	gp.isUseInfo = true
}

// GetInfo get Document Information Dictionary
func (gp *GoPdf) GetInfo() PdfInfo {
	return *gp.info
}

// Rotate rotate text or image
// angle is angle in degrees.
// x, y is rotation center
func (gp *GoPdf) Rotate(angle, x, y float64) {
	gp.UnitsToPointsVar(&x, &y)
	gp.getContent().appendRotate(angle, x, y)
}

// RotateReset reset rotate
func (gp *GoPdf) RotateReset() {
	gp.getContent().appendRotateReset()
}

// Polygon : draw polygon
//   - style: Style of polygon (draw and/or fill: D, F, DF, FD)
//     D or empty string: draw. This is the default value.
//     F: fill
//     DF or FD: draw and fill
//
// Usage:
//
//	 pdf.SetStrokeColor(255, 0, 0)
//		pdf.SetLineWidth(2)
//		pdf.SetFillColor(0, 255, 0)
//		pdf.Polygon([]tinypdf.Point{{X: 10, Y: 30}, {X: 585, Y: 200}, {X: 585, Y: 250}}, "DF")
func (gp *GoPdf) Polygon(points []Point, style string) {

	transparency, err := gp.getCachedTransparency(nil)
	if err != nil {
		transparency = nil
	}

	var opts = polygonOptions{}
	if transparency != nil {
		opts.extGStateIndexes = append(opts.extGStateIndexes, transparency.extGStateIndex)
	}

	var pointReals []Point
	for _, p := range points {
		x := p.X
		y := p.Y
		gp.UnitsToPointsVar(&x, &y)
		pointReals = append(pointReals, Point{X: x, Y: y})
	}
	gp.getContent().AppendStreamPolygon(pointReals, style, opts)
}

// Rectangle : draw rectangle, and add radius input to make a round corner, it helps to calculate the round corner coordinates and use Polygon functions to draw rectangle
//   - style: Style of Rectangle (draw and/or fill: D, F, DF, FD)
//     D or empty string: draw. This is the default value.
//     F: fill
//     DF or FD: draw and fill
//
// Usage:
//
//	 pdf.SetStrokeColor(255, 0, 0)
//		pdf.SetLineWidth(2)
//		pdf.SetFillColor(0, 255, 0)
//		pdf.Rectangle(196.6, 336.8, 398.3, 379.3, "DF", 3, 10)
func (gp *GoPdf) Rectangle(x0 float64, y0 float64, x1 float64, y1 float64, style string, radius float64, radiusPointNum int) error {
	if x1 <= x0 || y1 <= y0 {
		return errInvalidRectangleCoordinates
	}
	if radiusPointNum <= 0 || radius <= 0 {
		//draw rectangle without round corner
		points := []Point{}
		points = append(points, Point{X: x0, Y: y0})
		points = append(points, Point{X: x1, Y: y0})
		points = append(points, Point{X: x1, Y: y1})
		points = append(points, Point{X: x0, Y: y1})
		gp.Polygon(points, style)

	} else {

		if radius > (x1-x0) || radius > (y1-y0) {
			return errInvalidRectangleCoordinates
		}

		degrees := []float64{}
		angle := float64(90) / float64(radiusPointNum+1)
		accAngle := angle
		for accAngle < float64(90) {
			degrees = append(degrees, accAngle)
			accAngle += angle
		}

		radians := []float64{}
		for _, v := range degrees {
			radians = append(radians, v*math.Pi/180)
		}

		points := []Point{}
		points = append(points, Point{X: x0, Y: (y0 + radius)})
		for _, v := range radians {
			offsetX := radius * math.Cos(v)
			offsetY := radius * math.Sin(v)
			x := x0 + radius - offsetX
			y := y0 + radius - offsetY
			points = append(points, Point{X: x, Y: y})
		}
		points = append(points, Point{X: (x0 + radius), Y: y0})

		points = append(points, Point{X: (x1 - radius), Y: y0})
		for i := range radians {
			v := radians[len(radians)-1-i]
			offsetX := radius * math.Cos(v)
			offsetY := radius * math.Sin(v)
			x := x1 - radius + offsetX
			y := y0 + radius - offsetY
			points = append(points, Point{X: x, Y: y})
		}
		points = append(points, Point{X: x1, Y: (y0 + radius)})

		points = append(points, Point{X: x1, Y: (y1 - radius)})
		for _, v := range radians {
			offsetX := radius * math.Cos(v)
			offsetY := radius * math.Sin(v)
			x := x1 - radius + offsetX
			y := y1 - radius + offsetY
			points = append(points, Point{X: x, Y: y})
		}
		points = append(points, Point{X: (x1 - radius), Y: y1})

		points = append(points, Point{X: (x0 + radius), Y: y1})
		for i := range radians {
			v := radians[len(radians)-1-i]
			offsetX := radius * math.Cos(v)
			offsetY := radius * math.Sin(v)
			x := x0 + radius - offsetX
			y := y1 - radius + offsetY
			points = append(points, Point{X: x, Y: y})
		}
		points = append(points, Point{X: x0, Y: y1 - radius})

		gp.Polygon(points, style)
	}
	return nil
}

/*---private---*/

// init
func (gp *GoPdf) init(importer ...*importer) {
	gp.pdfObjs = []IObj{}
	gp.buf = bytes.Buffer{}
	gp.indexEncodingObjFonts = []int{}
	gp.pdfProtection = nil
	gp.encryptionObjID = 0
	gp.isUseInfo = false
	gp.info = nil

	//default
	gp.margins = Margins{
		Left:   defaultMargin,
		Top:    defaultMargin,
		Right:  defaultMargin,
		Bottom: defaultMargin,
	}

	//init curr
	gp.resetCurrXY()
	gp.curr = Current{}
	gp.curr.IndexOfPageObj = -1
	gp.curr.CountOfFont = 0
	gp.curr.CountOfL = 0
	gp.curr.CountOfImg = 0                       //img
	gp.curr.ImgCaches = make(map[int]ImageCache) //= *new([]ImageCache)
	gp.curr.sMasksMap = NewSMaskMap()
	gp.curr.extGStatesMap = NewExtGStatesMap()
	gp.curr.transparencyMap = NewTransparencyMap()
	gp.anchors = make(map[string]anchorOption)
	gp.curr.txtColorMode = "gray"

	//init index
	gp.indexOfPagesObj = -1
	gp.indexOfFirstPageObj = -1
	gp.indexOfContent = -1

	//No underline
	//gp.IsUnderline = false
	gp.curr.lineWidth = 1

	// default to zlib.DefaultCompression
	gp.compressLevel = zlib.DefaultCompression

	// change the unit type
	gp.config.PageSize = *gp.config.PageSize.unitsToPoints(gp.config)
	gp.config.TrimBox = *gp.config.TrimBox.unitsToPoints(gp.config)

	// init gofpdi free pdf document importer
	gp.fpdi = gp.importerOrDefault(importer...)

}

func (gp *GoPdf) importerOrDefault(importer ...*importer) *importer {
	if len(importer) != 0 {
		return importer[len(importer)-1]
	}

	return newImporter(gp.log)
}

func (gp *GoPdf) resetCurrXY() {
	gp.curr.X = gp.margins.Left
	gp.curr.Y = gp.margins.Top
}

// UnitsToPoints converts the units to the documents unit type
func (gp *GoPdf) UnitsToPoints(u float64) float64 {
	return unitsToPoints(gp.config, u)
}

// UnitsToPointsVar converts the units to the documents unit type for all variables passed in
func (gp *GoPdf) UnitsToPointsVar(u ...*float64) {
	unitsToPointsVar(gp.config, u...)
}

// PointsToUnits converts the points to the documents unit type
func (gp *GoPdf) PointsToUnits(u float64) float64 {
	return pointsToUnits(gp.config, u)
}

// PointsToUnitsVar converts the points to the documents unit type for all variables passed in
func (gp *GoPdf) PointsToUnitsVar(u ...*float64) {
	pointsToUnitsVar(gp.config, u...)
}

func (gp *GoPdf) isUseProtection() bool {
	return gp.config.Protection.UseProtection
}

func (gp *GoPdf) createProtection() *PDFProtection {
	var prot PDFProtection
	prot.setProtection(
		gp.config.Protection.Permissions,
		gp.config.Protection.UserPass,
		gp.config.Protection.OwnerPass,
	)
	return &prot
}

func (gp *GoPdf) protection() *PDFProtection {
	return gp.pdfProtection
}

func (gp *GoPdf) prepare() {

	if gp.isUseProtection() {
		encObj := gp.pdfProtection.encryptionObj()
		gp.addObj(encObj)
	}

	if gp.outlines.Count() > 0 {
		catalogObj := gp.pdfObjs[gp.indexOfCatalogObj].(*CatalogObj)
		catalogObj.SetIndexObjOutlines(gp.indexOfOutlinesObj)
	}

	if gp.indexOfPagesObj != -1 {
		indexCurrPage := -1
		pagesObj := gp.pdfObjs[gp.indexOfPagesObj].(*PagesObj)
		i := 0 //gp.indexOfFirstPageObj
		max := len(gp.pdfObjs)
		for i < max {
			objtype := gp.pdfObjs[i].getType()
			switch objtype {
			case "Page":
				pagesObj.Kids = pagesObj.Kids + strconv.Itoa(i+1) + " 0 R "
				pagesObj.PageCount++
				indexCurrPage = i
			case "Content":
				if indexCurrPage != -1 {
					gp.pdfObjs[indexCurrPage].(*PageObj).Contents = gp.pdfObjs[indexCurrPage].(*PageObj).Contents + strconv.Itoa(i+1) + " 0 R "
				}
			case "Font":
				tmpfont := gp.pdfObjs[i].(*FontObj)
				j := 0
				jmax := len(gp.indexEncodingObjFonts)
				for j < jmax {
					tmpencoding := gp.pdfObjs[gp.indexEncodingObjFonts[j]].(*EncodingObj).GetFont()
					if tmpfont.Family == tmpencoding.GetFamily() { //ใส่ ข้อมูลของ embed font
						tmpfont.IsEmbedFont = true
						tmpfont.SetIndexObjEncoding(gp.indexEncodingObjFonts[j] + 1)
						tmpfont.SetIndexObjWidth(gp.indexEncodingObjFonts[j] + 2)
						tmpfont.SetIndexObjFontDescriptor(gp.indexEncodingObjFonts[j] + 3)
						break
					}
					j++
				}
			case "Encryption":
				gp.encryptionObjID = i + 1
			}
			i++
		}
	}
}

func (gp *GoPdf) xref(w io.Writer, xrefbyteoffset int64, linelens []int64, i int) error {
	io.WriteString(w, "xref\n")
	io.WriteString(w, "0 "+strconv.Itoa(i+1)+"\n")
	io.WriteString(w, "0000000000 65535 f \n")
	j := 0
	max := len(linelens)
	for j < max {
		linelen := linelens[j]
		io.WriteString(w, gp.formatXrefline(linelen))
		io.WriteString(w, " 00000 n \n")
		j++
	}
	io.WriteString(w, "trailer\n")
	io.WriteString(w, "<<\n")
	io.WriteString(w, "/Size "+strconv.Itoa(max+1)+"\n")
	io.WriteString(w, "/Root 1 0 R\n")
	if gp.isUseProtection() {
		io.WriteString(w, "/Encrypt "+strconv.Itoa(gp.encryptionObjID)+" 0 R\n")
		io.WriteString(w, "/ID [()()]\n")
	}
	if gp.isUseInfo {
		gp.writeInfo(w)
	}
	io.WriteString(w, ">>\n")
	io.WriteString(w, "startxref\n")
	io.WriteString(w, strconv.FormatInt(xrefbyteoffset, 10))
	io.WriteString(w, "\n%%EOF\n")

	return nil
}

func (gp *GoPdf) writeInfo(w io.Writer) {
	var zerotime time.Time
	io.WriteString(w, "/Info <<\n")

	if gp.info.Author != "" {
		io.WriteString(w, "/Author <FEFF"+encodeUtf8(gp.info.Author)+">\n")
	}

	if gp.info.Title != "" {
		io.WriteString(w, "/Title <FEFF"+encodeUtf8(gp.info.Title)+">\n")
	}

	if gp.info.Subject != "" {
		io.WriteString(w, "/Subject <FEFF"+encodeUtf8(gp.info.Subject)+">\n")
	}

	if gp.info.Creator != "" {
		io.WriteString(w, "/Creator <FEFF"+encodeUtf8(gp.info.Creator)+">\n")
	}

	if gp.info.Producer != "" {
		io.WriteString(w, "/Producer <FEFF"+encodeUtf8(gp.info.Producer)+">\n")
	}

	if !zerotime.Equal(gp.info.CreationDate) {
		io.WriteString(w, "/CreationDate(D:"+infodate(gp.info.CreationDate)+")\n")
	}

	io.WriteString(w, " >>\n")
}

// ปรับ xref ให้เป็น 10 หลัก
func (gp *GoPdf) formatXrefline(n int64) string {
	str := strconv.FormatInt(n, 10)
	for len(str) < 10 {
		str = "0" + str
	}
	return str
}

func (gp *GoPdf) addObj(iobj IObj) int {
	index := len(gp.pdfObjs)
	gp.pdfObjs = append(gp.pdfObjs, iobj)
	return index
}

func (gp *GoPdf) getContent() *ContentObj {
	var content *ContentObj
	if gp.indexOfContent <= -1 {
		content = new(ContentObj)
		content.init(func() *GoPdf {
			return gp
		})
		gp.indexOfContent = gp.addObj(content)
	} else {
		content = gp.pdfObjs[gp.indexOfContent].(*ContentObj)
	}
	return content
}

func encodeUtf8(str string) string {
	var buff bytes.Buffer
	for _, r := range str {
		// Convertir runa a hexadecimal usando strconv en lugar de fmt
		hex := strconv.FormatInt(int64(r), 16)

		// Asegurar que tenga 4 caracteres (rellenando con ceros)
		for len(hex) < 4 {
			hex = "0" + hex
		}

		// Convertir a mayúsculas manualmente y añadir al buffer
		for i := 0; i < len(hex); i++ {
			c := hex[i]
			if c >= 'a' && c <= 'f' {
				buff.WriteByte(c - 32) // 'a'-'A' = 32 en ASCII
			} else {
				buff.WriteByte(c)
			}
		}
	}
	return buff.String()
}

func infodate(t time.Time) string {
	ft := t.Format("20060102150405-07'00'")
	return ft
}

// SetTransparency sets transparency.
// alpha: 		value from 0 (transparent) to 1 (opaque)
// blendMode:   blend mode, one of the following:
//
//	Normal, Multiply, Screen, Overlay, Darken, Lighten, ColorDodge, ColorBurn,
//	HardLight, SoftLight, Difference, Exclusion, Hue, Saturation, Color, Luminosity
func (gp *GoPdf) SetTransparency(transparency Transparency) error {
	t, err := gp.saveTransparency(&transparency)
	if err != nil {
		return err
	}

	gp.curr.transparency = t

	return nil
}

func (gp *GoPdf) ClearTransparency() {
	gp.curr.transparency = nil
}

func (gp *GoPdf) getCachedTransparency(transparency *Transparency) (*Transparency, error) {
	if transparency == nil {
		transparency = gp.curr.transparency
	} else {
		cached, err := gp.saveTransparency(transparency)
		if err != nil {
			return nil, err
		}

		transparency = cached
	}

	return transparency, nil
}

func (gp *GoPdf) saveTransparency(transparency *Transparency) (*Transparency, error) {
	cached, ok := gp.curr.transparencyMap.Find(*transparency)
	if ok {
		return &cached, nil
	} else if transparency.Alpha != DefaultAplhaValue {
		bm := transparency.BlendModeType
		opts := ExtGStateOptions{
			BlendMode:     &bm,
			StrokingCA:    &transparency.Alpha,
			NonStrokingCa: &transparency.Alpha,
		}

		extGState, err := GetCachedExtGState(opts, gp)
		if err != nil {
			return nil, err
		}

		transparency.extGStateIndex = extGState.Index + 1

		gp.curr.transparencyMap.Save(*transparency)

		return transparency, nil
	}

	return nil, nil
}

// IsCurrFontContainGlyph defines is current font contains to a glyph
// r:           any rune
func (gp *GoPdf) IsCurrFontContainGlyph(r rune) (bool, error) {
	fontISubset := gp.curr.FontISubset
	if fontISubset == nil {
		return false, nil
	}

	glyphIndex, err := fontISubset.CharCodeToGlyphIndex(r)
	if err == ErrGlyphNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	if glyphIndex == 0 {
		return false, nil
	}

	return true, nil
}

//tool for validate pdf https://www.pdf-online.com/osa/validate.aspx
