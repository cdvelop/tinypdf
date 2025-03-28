package tinypdf

// FontStyle defines the available font styles
const (
	FontRegular = "regular"
	FontBold    = "bold"
	FontItalic  = "italic"
)

type elementPosition string

const (
	inlinePosition  elementPosition = "inline"
	newlinePosition elementPosition = "newline"
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

// TextBuilder is a helper struct to build text cells
type TextBuilder struct {
	doc         *Document
	text        string
	opts        CellOption
	rect        *Rect
	style       TextStyle
	fontName    string
	fullWidth   bool            // Por defecto es false (solo usa el ancho necesario)
	positioning elementPosition // "inline", "newline" (por defecto newline)
}

// newTextBuilder creates a new TextBuilder with the given style
func (d *Document) newTextBuilder(text string, style TextStyle, fontName string) *TextBuilder {
	builder := &TextBuilder{
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
func (d *Document) AddTextOLD(text string) *TextBuilder {
	return d.newTextBuilder(text, d.fontConfig.Normal, FontRegular)
}

// AddText crea texto normal
func (d *Document) AddText(text string) *TextBuilder {
	tb := d.newTextBuilder(text, d.fontConfig.Normal, FontRegular)
	tb.fullWidth = false // Solo para texto normal, usar ancho automático
	return tb
}

// AddHeader1 crea un encabezado nivel 1
func (d *Document) AddHeader1(text string) *TextBuilder {
	return d.newTextBuilder(text, d.fontConfig.Header1, FontBold)
}

// AddHeader2 crea un encabezado nivel 2
func (d *Document) AddHeader2(text string) *TextBuilder {
	return d.newTextBuilder(text, d.fontConfig.Header2, FontBold)
}

// AddHeader3 crea un encabezado nivel 3
func (d *Document) AddHeader3(text string) *TextBuilder {
	return d.newTextBuilder(text, d.fontConfig.Header3, FontBold)
}

// AddFooter crea un pie de página
func (d *Document) AddFooter(text string) *TextBuilder {
	return d.newTextBuilder(text, d.fontConfig.Footer, FontRegular)
}

// AddFootnote crea una nota al pie
func (d *Document) AddFootnote(text string) *TextBuilder {
	return d.newTextBuilder(text, d.fontConfig.Footnote, FontItalic)
}

// AddJustifiedText crea texto justificado directamente
func (d *Document) AddJustifiedText(text string) *TextBuilder {
	tb := d.AddText(text)
	return tb.Justify()
}

func (tb *TextBuilder) AlignCenter() *TextBuilder {
	tb.opts.Align = Center | Top
	return tb
}

func (tb *TextBuilder) AlignRight() *TextBuilder {
	tb.opts.Align = Right | Top
	return tb
}

func (tb *TextBuilder) AlignLeft() *TextBuilder {
	tb.opts.Align = Left | Top
	return tb
}

func (tb *TextBuilder) Justify() *TextBuilder {
	tb.opts.Align = Justify | Top
	return tb
}

func (tb *TextBuilder) WithBorder() *TextBuilder {
	tb.opts.Border = AllBorders
	return tb
}

// Métodos para cambiar el estilo de fuente
func (tb *TextBuilder) Bold() *TextBuilder {
	tb.fontName = FontBold
	tb.doc.SetFont(FontBold, "", tb.style.Size)
	return tb
}

func (tb *TextBuilder) Italic() *TextBuilder {
	tb.fontName = FontItalic
	tb.doc.SetFont(FontItalic, "", tb.style.Size)
	return tb
}

func (tb *TextBuilder) Regular() *TextBuilder {
	tb.fontName = FontRegular
	tb.doc.SetFont(FontRegular, "", tb.style.Size)
	return tb
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
func (tb *TextBuilder) FullWidth() *TextBuilder {
	tb.fullWidth = true
	return tb
}

// WidthPercent establece el ancho como porcentaje del ancho de página
func (tb *TextBuilder) WidthPercent(percent float64) *TextBuilder {
	if percent > 0 && percent <= 100 {
		tb.rect.W = tb.doc.pageWidth * (percent / 100)
	}
	return tb
}

// Inline intenta posicionar este elemento en la misma línea que el anterior
func (tb *TextBuilder) Inline() *TextBuilder {
	tb.positioning = inlinePosition
	return tb
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
func (tb *TextBuilder) minimumWidthRequiredForText() {
	// Calcular ancho necesario para el texto si no se especificó un ancho fijo
	// o si se solicitó ancho completo
	if tb.fullWidth {
		// Usar ancho completo de la página
		tb.rect.W = tb.doc.pageWidth
	} else {
		// Si no se especificó un ancho, calcular el ancho mínimo necesario
		if tb.rect.W <= 0 {
			// Obtener el factor de ancho según el tipo de fuente
			widthFactor := tb.doc.measureTextWidthFactor(tb.fontName)

			// Calcular ancho considerando un factor de reducción más realista
			charWidth := tb.style.Size * widthFactor

			// Considerar longitud efectiva (algunos caracteres son más estrechos)
			effectiveLength := float64(len(tb.text)) * 0.8 // Reducir un % por espacios y caracteres estrechos
			// effectiveLength := float64(len(tb.text)) * 0.9 // Reducir un 10% por espacios y caracteres estrechos

			// Calcular ancho estimado
			width := effectiveLength * charWidth

			// Añadir un pequeño margen
			width += tb.style.Size // Añadir un margen completo del tamaño de la fuente

			// Si el texto es largo usar el ancho de página
			if width >= tb.doc.pageWidth {
				width = tb.doc.pageWidth
			}

			// Asegurar un ancho mínimo razonable
			minWidth := tb.style.Size
			if width < minWidth {
				width = minWidth
			}

			tb.rect.W = width
		}
	}
}

// Draw renders the text on the document
func (tb *TextBuilder) Draw() error {
	// Apply space before the paragraph
	if tb.style.SpaceBefore > 0 {
		tb.doc.SetY(tb.doc.GetY() + tb.style.SpaceBefore)
	}

	// Posicionamiento inline si se solicitó
	if tb.positioning == inlinePosition {
		// Mantener posición X actual
		// tb.opts.Float = Right
	}

	tb.minimumWidthRequiredForText()

	// Calculate how many lines the text will occupy
	textSplits, err := tb.doc.SplitTextWithOption(tb.text, tb.rect.W, tb.opts.BreakOption)
	if err != nil {
		return err
	}

	// Get line height in current font and size
	_, lineHeight, _, err := createContent(tb.doc.curr.FontISubset, tb.text,
		tb.doc.curr.FontSize, tb.doc.curr.CharSpacing, nil)
	if err != nil {
		return err
	}

	tb.doc.PointsToUnitsVar(&lineHeight)

	// Calculate total height needed for all lines
	totalHeight := float64(len(textSplits)) * lineHeight

	// Set the rectangle height to accommodate all text
	tb.rect.H = totalHeight

	// Draw the text with the properly sized rectangle
	err = tb.doc.MultiCellWithOption(tb.rect, tb.text, tb.opts)
	if err != nil {
		return err
	}

	// Reset font to regular for next text (prevents style bleed)
	tb.doc.setDefaultFont()

	// Apply space after the paragraph
	tb.doc.SetY(tb.doc.GetY() + tb.style.Size + tb.style.SpaceAfter)

	return nil
}
