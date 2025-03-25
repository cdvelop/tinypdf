package tinypdf

type Document struct {
	*GoPdf
	fontConfig FontConfig
	pageWidth  float64
	log        func(a ...any)
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
		fontConfig: DefaultFontConfig(),
		log:        logPrint,
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
