package tinypdf

// FontStyle defines the available font styles
const (
	FontRegular = "regular"
	FontBold    = "bold"
	FontItalic  = "italic"
)

type elementPosition string

const (
	inlinePosition  elementPosition = "inline"  // posicionamiento en línea (inline)
	newlinePosition elementPosition = "newline" // posicionamiento por defecto (salto de línea)
	fixedPosition   elementPosition = "fixed"   //posicionamiento fijo (no se mueve con el texto)
)

// TextStyle defines text appearance properties
type TextStyle struct {
	Size        float64
	Color       RGBColor
	LineSpacing float64
	Alignment   int     // Uses same alignment constants as CellOption (Left, Center, Right, etc)
	SpaceBefore float64 // Space before paragraph (in points)
	SpaceAfter  float64 // Space after paragraph (in points)
}

// docText is a helper struct to build text cells
type docText struct {
	doc         *Document
	text        string
	opts        CellOption
	rect        *Rect
	style       TextStyle
	fontName    string
	fullWidth   bool            // Por defecto es false (solo usa el ancho necesario)
	positioning elementPosition // "inline", "newline" (por defecto newline)
}

// newTextBuilder creates a new docText with the given style
func (d *Document) newTextBuilder(text string, style TextStyle, fontName string) *docText {
	builder := &docText{
		doc:         d,
		text:        text,
		style:       style, // Store the style
		fontName:    fontName,
		fullWidth:   true,            // Por defecto usar ancho completo para mantener compatibilidad
		positioning: newlinePosition, // Por defecto es newline
		rect: &Rect{
			W: 0, // se calcula en Draw()
			H: 0,
		},
		opts: CellOption{
			Align:          style.Alignment,
			Border:         0,
			Float:          Bottom,
			CoefLineHeight: style.LineSpacing,
		},
	}

	// Apply style
	d.SetFont(fontName, "", style.Size)
	d.SetTextColor(style.Color.R, style.Color.G, style.Color.B)

	return builder
}

// AddText crea texto normal
func (d *Document) AddText(text string) *docText {
	dt := d.newTextBuilder(text, d.fontConfig.Normal, FontRegular)
	dt.fullWidth = false // Solo para texto normal, usar ancho automático
	return dt
}

// AddHeader1 crea un encabezado nivel 1
func (d *Document) AddHeader1(text string) *docText {
	return d.newTextBuilder(text, d.fontConfig.Header1, FontBold)
}

// AddHeader2 crea un encabezado nivel 2
func (d *Document) AddHeader2(text string) *docText {
	return d.newTextBuilder(text, d.fontConfig.Header2, FontBold)
}

// AddHeader3 crea un encabezado nivel 3
func (d *Document) AddHeader3(text string) *docText {
	return d.newTextBuilder(text, d.fontConfig.Header3, FontBold)
}

// AddFootnote crea una nota al pie
func (d *Document) AddFootnote(text string) *docText {
	return d.newTextBuilder(text, d.fontConfig.Footnote, FontItalic)
}

// AddJustifiedText crea texto justificado directamente
func (d *Document) AddJustifiedText(text string) *docText {
	dt := d.AddText(text)
	return dt.Justify()
}

func (dt *docText) AlignCenter() *docText {
	dt.opts.Align = Center | Top
	return dt
}

func (dt *docText) AlignRight() *docText {
	dt.opts.Align = Right | Top
	dt.fullWidth = true
	return dt
}

func (dt *docText) AlignLeft() *docText {
	dt.opts.Align = Left | Top
	return dt
}

func (dt *docText) Justify() *docText {
	dt.opts.Align = Justify | Top
	return dt
}

func (dt *docText) WithBorder() *docText {
	dt.opts.Border = AllBorders
	return dt
}

// Métodos para cambiar el estilo de fuente
func (dt *docText) Bold() *docText {
	dt.fontName = FontBold
	dt.doc.SetFont(FontBold, "", dt.style.Size)
	return dt
}

func (dt *docText) Italic() *docText {
	dt.fontName = FontItalic
	dt.doc.SetFont(FontItalic, "", dt.style.Size)
	return dt
}

func (dt *docText) Regular() *docText {
	dt.fontName = FontRegular
	dt.doc.SetFont(FontRegular, "", dt.style.Size)
	return dt
}

// SpaceBefore adds vertical space (in font spaces)
// example: SpaceBefore(2) adds 2 spaces before the text
func (d *Document) SpaceBefore(spaces ...float64) {
	space := 1.0 // Default value is 1 space if no parameter provided
	if len(spaces) > 0 && spaces[0] > 0 {
		space = spaces[0]
	}

	// Get the current font size
	fontSize := d.curr.FontSize
	if fontSize <= 0 {
		fontSize = d.fontConfig.Normal.Size // Default font size if none is set
	}

	// Add vertical space based on font size
	d.SetY(d.GetY() + fontSize*space)
}

// FullWidth hace que el texto ocupe todo el ancho disponible
func (dt *docText) FullWidth() *docText {
	dt.fullWidth = true
	return dt
}

// WidthPercent establece el ancho como porcentaje del ancho de página
func (dt *docText) WidthPercent(percent float64) *docText {
	if percent > 0 && percent <= 100 {
		dt.rect.W = dt.doc.pageWidth * (percent / 100)
	}
	return dt
}

// Inline intenta posicionar este elemento en la misma línea que el anterior
func (dt *docText) Inline() *docText {
	dt.positioning = inlinePosition
	return dt
}

// retorna el factor de ancho para una fuente específica
func (d *Document) measureTextWidthFactor(fontName string) float64 {
	// Factor de escala para el ancho de caracteres, varía según el estilo
	var widthFactor float64 = 0.6 // Factor para fuente regular

	// Ajustar factor según el estilo de fuente
	switch fontName {
	case FontBold:
		widthFactor = 0.65 // La negrita es un poco más ancha
	case FontItalic:
		widthFactor = 0.55 // La itálica suele ser ligeramente más estrecha
	}

	return widthFactor
}

// estima el ancho mínimo necesario para el texto
func (dt *docText) minimumWidthRequiredForText() {
	// Calcular ancho necesario para el texto si no se especificó un ancho fijo
	// o si se solicitó ancho completo
	if dt.fullWidth {
		// Usar ancho completo de la página
		dt.rect.W = dt.doc.pageWidth
	} else {
		// Si no se especificó un ancho, calcular el ancho mínimo necesario
		if dt.rect.W <= 0 {
			// Obtener el factor de ancho según el tipo de fuente
			widthFactor := dt.doc.measureTextWidthFactor(dt.fontName)

			// Calcular ancho considerando un factor de reducción más realista
			charWidth := dt.style.Size * widthFactor

			// Considerar longitud efectiva (algunos caracteres son más estrechos)
			// Aumentar el factor para evitar problemas con signos de puntuación
			effectiveLength := float64(len(dt.text)) * 0.85 // Reducir solo 15% en lugar de 20%

			// Add extra space for text ending with punctuation marks
			if len(dt.text) > 0 {
				lastChar := dt.text[len(dt.text)-1]
				if lastChar == ':' || lastChar == '.' || lastChar == '!' || lastChar == '?' {
					effectiveLength += 0.5 // Add extra space for punctuation marks
				}
			}

			// Calcular ancho estimado
			width := effectiveLength * charWidth

			// Añadir un pequeño margen
			width += dt.style.Size * 1.1 // Aumentar el margen un 10%

			// Si el texto es largo usar el ancho de página
			if width >= dt.doc.pageWidth {
				width = dt.doc.pageWidth
			}

			// Asegurar un ancho mínimo razonable
			minWidth := dt.style.Size
			if width < minWidth {
				width = minWidth
			}

			dt.rect.W = width
		}
	}
}

// Draw renders the text on the document to include page break handling
func (dt *docText) Draw() error {
	// Apply space before the paragraph
	if dt.style.SpaceBefore > 0 {
		dt.doc.SetY(dt.doc.GetY() + dt.style.SpaceBefore)
	}

	// Handle positioning
	if dt.positioning == inlinePosition {
		// Keep current X position for inline elements
	} else {
		// Si no es inline, siempre restauramos la posición X al margen izquierdo
		// independientemente de si el elemento anterior era inline o no
		dt.doc.SetX(dt.doc.margins.Left)
		dt.doc.inlineMode = false
	}

	dt.minimumWidthRequiredForText()

	// Calculate how many lines the text will occupy
	textSplits, err := dt.doc.SplitTextWithOption(dt.text, dt.rect.W, dt.opts.BreakOption)
	if err != nil {
		return err
	}

	// Get line height in current font and size
	_, lineHeight, _, err := createContent(dt.doc.curr.FontISubset, dt.text,
		dt.doc.curr.FontSize, dt.doc.curr.CharSpacing, nil)
	if err != nil {
		return err
	}

	dt.doc.PointsToUnitsVar(&lineHeight)

	// Calculate total height needed for all lines
	totalHeight := float64(len(textSplits)) * lineHeight

	// Set the rectangle height to accommodate all text
	dt.rect.H = totalHeight

	// Check if the text fits on current page
	newY, _ := dt.doc.ensureElementFits(totalHeight, dt.style.SpaceAfter)
	dt.doc.SetY(newY)

	// Draw the text with the properly sized rectangle
	err = dt.doc.MultiCellWithOption(dt.rect, dt.text, dt.opts)
	if err != nil {
		return err
	}

	// Update inline mode based on current element's positioning
	dt.doc.inlineMode = (dt.positioning == inlinePosition)

	// If not inline, ensure we do a proper line break
	if dt.positioning == newlinePosition {
		dt.doc.newLineBreakBasedOnDefaultFont(dt.doc.GetY())
	}

	return nil
}

func (doc *Document) newLineBreakBasedOnDefaultFont(originY float64) {
	// Reset font to regular for next text (prevents style bleed)
	doc.setDefaultFont()

	// Apply space after the paragraph
	doc.SetY(originY + doc.fontConfig.Normal.Size + doc.fontConfig.Normal.SpaceAfter)
}
