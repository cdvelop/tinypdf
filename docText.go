package tinypdf

// FontStyle defines the available font styles
const (
	FontRegular = "regular"
	FontBold    = "bold"
	FontItalic  = "italic"
)

// TextStyle defines text appearance properties
type TextStyle struct {
	Size        float64
	Color       RGBColor
	LineSpacing float64
	Alignment   int // Uses same alignment constants as CellOption (Left, Center, Right, etc)
}

// TextBuilder is a helper struct to build text cells
type TextBuilder struct {
	doc      *Document
	text     string
	opts     CellOption
	rect     *Rect
	style    TextStyle
	fontName string
}

// newTextBuilder creates a new TextBuilder with the given style
func (d *Document) newTextBuilder(text string, style TextStyle, fontName string) *TextBuilder {
	builder := &TextBuilder{
		doc:      d,
		text:     text,
		style:    style, // Store the style
		fontName: fontName,
		rect: &Rect{
			W: d.pageWidth,
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
func (d *Document) AddText(text string) *TextBuilder {
	return d.newTextBuilder(text, d.fontConfig.Normal, FontRegular)
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

// Draw render the text with the specified options
func (tb *TextBuilder) Draw() error {
	// Draw the text
	err := tb.doc.MultiCellWithOption(tb.rect, tb.text, tb.opts)
	if err != nil {
		return err
	}
	// Reset font to regular for next text (prevents style bleed)
	tb.doc.setDefaultFont()

	// Calculate appropriate spacing based on the current text style
	// Use the font size and line spacing from the text style
	spacing := tb.style.Size * tb.style.LineSpacing

	tb.doc.SetY(tb.doc.GetY() + spacing)

	return nil
}

// Br adds vertical space (in mm)
func (d *Document) Br(mm float64) {
	// Convert mm to points (1 mm = 72.0/25.4 points)
	d.SetY(d.GetY() + mm*(72.0/25.4))
}
