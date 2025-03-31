package tinypdf

import "strconv"

type Document struct {
	*GoPdf
	fontConfig      FontConfig
	pageWidth       float64
	inlineMode      bool    // Add this field to track inline element state
	lastInlineWidth float64 // Track the width of the last inline element
	log             func(a ...any)
}

// NewDocument creates a new PDF document with configurable settings
// Accepts optional configurations including custom margins in millimeters:
//   - FontConfig: Custom font configuration
//   - Font: Custom font family
//   - Margins: Custom margins in millimeters (more intuitive than points)
//
// Example: NewDocument(fmt.Println, Margins{Left: 15, Top: 10, Right: 10, Bottom: 10})
func NewDocument(logPrint func(a ...any), configs ...any) *Document {

	doc := &Document{
		GoPdf: &GoPdf{
			log: logPrint,
		},
		fontConfig:      DefaultFontConfig(),
		inlineMode:      false,
		lastInlineWidth: 0,
		log:             logPrint,
	}

	// Default margins: 1.5 cm left, 1 cm on other sides
	leftMargin := 42.52   // 1.5 cm in points
	otherMargins := 28.35 // 1 cm in points

	// Start with default page configuration
	doc.Start(Config{
		PageSize: *PageSizeLetter,
	})

	// Set default margins explicitly
	doc.SetMargins(leftMargin, otherMargins, otherMargins, otherMargins)

	// Process all configurations in one place
	for _, v := range configs {
		switch v := v.(type) {
		case FontConfig:
			doc.fontConfig = v
		case Font:
			doc.fontConfig.Family = v
		case Margins:
			// Convert millimeters to points (1 mm = 72.0/25.4 points)
			doc.SetMargins(
				v.Left*(72.0/25.4),
				v.Top*(72.0/25.4),
				v.Right*(72.0/25.4),
				v.Bottom*(72.0/25.4),
			)
		}
	}

	err := doc.loadFonts()
	if err != nil {
		doc.log("Error loading fonts: ", err)
	}

	doc.pageWidth = doc.config.PageSize.W - (doc.margins.Left + doc.margins.Right)

	doc.AddPage()
	doc.setDefaultFont()

	return doc
}

// GetLineHeight returns the current line height based on the active font and size
func (doc *Document) GetLineHeight() float64 {
	// Get current font size and add some padding
	fontSize := doc.curr.FontSize
	if fontSize <= 0 {
		fontSize = doc.fontConfig.Normal.Size // Default font size as fallback
	}

	// Line height is typically 1.2 to 1.5 times the font size
	// Using 1.2 as a conservative multiplier
	return fontSize * 1.2
}

// AddPage : add new page
func (gp *GoPdf) AddPage() {
	emptyOpt := PageOption{}
	gp.AddPageWithOption(emptyOpt)
}

// AddPageWithOption  : add new page with option
func (gp *GoPdf) AddPageWithOption(opt PageOption) {
	opt.TrimBox = opt.TrimBox.UnitsToPoints(gp.config.Unit)
	opt.PageSize = opt.PageSize.UnitsToPoints(gp.config.Unit)

	page := new(PageObj)
	page.init(func() *GoPdf {
		return gp
	})

	if !opt.isEmpty() { //use page option
		page.setOption(opt)
		gp.curr.pageSize = opt.PageSize

		if opt.isTrimBoxSet() {
			gp.curr.trimBox = opt.TrimBox
		}
	} else { //use default
		gp.curr.pageSize = &gp.config.PageSize
		gp.curr.trimBox = &gp.config.TrimBox
	}

	page.ResourcesRelate = strconv.Itoa(gp.indexOfProcSet+1) + " 0 R"
	index := gp.addObj(page)
	if gp.indexOfFirstPageObj == -1 {
		gp.indexOfFirstPageObj = index
	}
	gp.curr.IndexOfPageObj = index

	gp.numOfPagesObj++

	//reset
	gp.indexOfContent = -1
	gp.resetCurrXY()

	if gp.headerFunc != nil {
		gp.headerFunc()
		gp.resetCurrXY()
	}

	if gp.footerFunc != nil {
		gp.footerFunc()
		gp.resetCurrXY()
	}
}

// SetPage set current page
func (gp *GoPdf) SetPage(pageno int) error {
	var pageIndex int
	for i := 0; i < len(gp.pdfObjs); i++ {
		switch gp.pdfObjs[i].(type) {
		case *ContentObj:
			pageIndex += 1
			if pageIndex == pageno {
				gp.indexOfContent = i
				return nil
			}
		}
	}

	return newErr("invalid page number")
}
