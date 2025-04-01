package tinypdf

import (
	"strings"
)

// CellOption cell option
type CellOption struct {
	Align                  int //Allows to align the text. Possible values are: Left,Center,Right,Top,Bottom,Middle
	Border                 int //Indicates if borders must be drawn around the cell. Possible values are: Left, Top, Right, Bottom, ALL
	Float                  int //Indicates where the current position should go after the call. Possible values are: Right, Bottom
	Transparency           *Transparency
	CoefUnderlinePosition  float64
	CoefLineHeight         float64
	CoefUnderlineThickness float64
	BreakOption            *BreakOption

	extGStateIndexes []int
}

// Text write text start at current x,y ( current y is the baseline of text )
func (gp *GoPdf) Text(text string) error {

	text, err := gp.curr.FontISubset.AddChars(text)
	if err != nil {
		return err
	}

	err = gp.getContent().AppendStreamText(text)
	if err != nil {
		return err
	}

	return nil
}

// CellWithOption create cell of text ( use current x,y is upper-left corner of cell)
func (gp *GoPdf) CellWithOption(rectangle *Rect, text string, opt CellOption) error {
	transparency, err := gp.getCachedTransparency(opt.Transparency)
	if err != nil {
		return err
	}

	if transparency != nil {
		opt.extGStateIndexes = append(opt.extGStateIndexes, transparency.extGStateIndex)
	}

	rectangle = rectangle.UnitsToPoints(gp.config.Unit)
	text, err = gp.curr.FontISubset.AddChars(text)
	if err != nil {
		return err
	}
	if err := gp.getContent().AppendStreamSubsetFont(rectangle, text, opt); err != nil {
		return err
	}

	return nil
}

// Cell : create cell of text ( use current x,y is upper-left corner of cell)
// Note that this has no effect on Rect.H pdf (now). Fix later :-)
func (gp *GoPdf) Cell(rectangle *Rect, text string) error {
	rectangle = rectangle.UnitsToPoints(gp.config.Unit)
	defaultopt := CellOption{
		Align:  Left | Top,
		Border: 0,
		Float:  Right,
	}

	text, err := gp.curr.FontISubset.AddChars(text)
	if err != nil {
		return err
	}
	err = gp.getContent().AppendStreamSubsetFont(rectangle, text, defaultopt)
	if err != nil {
		return err
	}

	return nil
}

// [experimental]
// PlaceHolderText Create a text placehold for fillin text later with function FillInPlaceHoldText.
func (gp *GoPdf) PlaceHolderText(placeHolderName string, placeHolderWidth float64) error {

	//placeHolderText := fmt.Sprintf("{%s}", placeHolderName)
	_, err := gp.curr.FontISubset.AddChars("")
	if err != nil {
		return err
	}

	gp.PointsToUnitsVar(&placeHolderWidth)
	err = gp.getContent().appendStreamPlaceHolderText(placeHolderWidth)
	if err != nil {
		return err
	}

	content := gp.pdfObjs[gp.indexOfContent].(*ContentObj)
	indexInContent := len(content.listCache.caches) - 1
	indexOfContent := gp.indexOfContent
	fontISubset := gp.curr.FontISubset

	gp.placeHolderTexts[placeHolderName] = append(
		gp.placeHolderTexts[placeHolderName],
		placeHolderTextInfo{
			indexOfContent:   indexOfContent,
			indexInContent:   indexInContent,
			fontISubset:      fontISubset,
			placeHolderWidth: placeHolderWidth,
			fontSize:         gp.curr.FontSize,
			charSpacing:      gp.curr.CharSpacing,
		},
	)

	return nil
}

// [experimental]
// fill in text that created by function PlaceHolderText
// align: Left,Right,Center
func (gp *GoPdf) FillInPlaceHoldText(placeHolderName string, text string, align int) error {

	infos, ok := gp.placeHolderTexts[placeHolderName]
	if !ok {
		return newErr("placeHolderName not found")
	}

	for _, info := range infos {
		content, ok := gp.pdfObjs[info.indexOfContent].(*ContentObj)
		if !ok {
			return newErr("gp.pdfObjs is not *ContentObj")
		}
		contentText, ok := content.listCache.caches[info.indexInContent].(*cacheContentText)
		if !ok {
			return newErr("listCache.caches is not *cacheContentText")
		}
		info.fontISubset.AddChars(text)
		contentText.text = text

		//Calculate position
		_, _, textWidthPdfUnit, err := createContent(gp.curr.FontISubset, text, info.fontSize, info.charSpacing, nil)
		if err != nil {
			return err
		}
		width := pointsToUnits(gp.config, textWidthPdfUnit)

		if align == Right {
			diff := info.placeHolderWidth - width
			contentText.x = contentText.x + diff
		} else if align == Center {
			diff := info.placeHolderWidth - width
			contentText.x = contentText.x + diff/2
		}
	}

	return nil
}

// MultiCell : create of text with line breaks (use current x,y is upper-left corner of cell)
func (gp *GoPdf) MultiCell(rectangle *Rect, text string) error {
	var line []rune
	x := gp.GetX()
	var totalLineHeight float64
	length := len([]rune(text))

	// get lineHeight
	text, err := gp.curr.FontISubset.AddChars(text)
	if err != nil {
		return err
	}
	_, lineHeight, _, err := createContent(gp.curr.FontISubset, text, gp.curr.FontSize, gp.curr.CharSpacing, nil)
	if err != nil {
		return err
	}
	gp.PointsToUnitsVar(&lineHeight)

	for i, v := range []rune(text) {
		if totalLineHeight+lineHeight > rectangle.H {
			break
		}
		lineWidth, _ := gp.MeasureTextWidth(string(line))
		runeWidth, _ := gp.MeasureTextWidth(string(v))

		if lineWidth+runeWidth > rectangle.W {
			gp.Cell(&Rect{W: rectangle.W, H: lineHeight}, string(line))
			gp.Br(lineHeight)
			gp.SetX(x)
			totalLineHeight = totalLineHeight + lineHeight
			line = nil
		}

		line = append(line, v)

		if i == length-1 {
			gp.Cell(&Rect{W: rectangle.W, H: lineHeight}, string(line))
			gp.Br(lineHeight)
			gp.SetX(x)
		}
	}
	return nil
}

// IsFitMultiCell : check whether the rectangle's area is big enough for the text
func (gp *GoPdf) IsFitMultiCell(rectangle *Rect, text string) (bool, float64, error) {
	var line []rune
	var totalLineHeight float64
	length := len([]rune(text))

	// get lineHeight
	text, err := gp.curr.FontISubset.AddChars(text)
	if err != nil {
		return false, totalLineHeight, err
	}
	_, lineHeight, _, err := createContent(gp.curr.FontISubset, text, gp.curr.FontSize, gp.curr.CharSpacing, nil)

	if err != nil {
		return false, totalLineHeight, err
	}
	gp.PointsToUnitsVar(&lineHeight)

	for i, v := range []rune(text) {
		if totalLineHeight+lineHeight > rectangle.H {
			return false, totalLineHeight, nil
		}
		lineWidth, _ := gp.MeasureTextWidth(string(line))
		runeWidth, _ := gp.MeasureTextWidth(string(v))

		if lineWidth+runeWidth > rectangle.W {
			totalLineHeight += lineHeight
			line = nil
		}

		line = append(line, v)

		if i == length-1 {
			totalLineHeight += lineHeight
		}
	}

	ok := true
	if totalLineHeight > rectangle.H {
		ok = false
	}

	return ok, totalLineHeight, nil
}

// IsFitMultiCellWithNewline : similar to IsFitMultiCell, but process char newline as Br
func (gp *GoPdf) IsFitMultiCellWithNewline(rectangle *Rect, text string) (bool, float64, error) {
	r := *rectangle
	strs := strings.Split(text, "\n")

	for _, s := range strs {
		ok, height, err := gp.IsFitMultiCell(&r, s)
		if err != nil || !ok {
			return false, 0, err
		}
		r.H -= height
	}

	return true, rectangle.H - r.H, nil
}

// MultiCellWithOption create of text with line breaks (use current x,y is upper-left corner of cell)
func (gp *GoPdf) MultiCellWithOption(rectangle *Rect, text string, opt CellOption) error {
	if opt.BreakOption == nil {
		opt.BreakOption = &DefaultBreakOption
	}

	// Si es justificado, aseguramos que use BreakModeIndicatorSensitive para evitar cortar palabras
	isJustify := (opt.Align & Justify) == Justify
	if isJustify {
		// Guardar las opciones originales, pero forzar modo sensible a indicadores (espacios)
		originalOpt := *opt.BreakOption
		opt.BreakOption = &BreakOption{
			Mode:           BreakModeIndicatorSensitive,
			BreakIndicator: ' ',
			Separator:      originalOpt.Separator,
		}
	}

	transparency, err := gp.getCachedTransparency(opt.Transparency)
	if err != nil {
		return err
	}

	if transparency != nil {
		opt.extGStateIndexes = append(opt.extGStateIndexes, transparency.extGStateIndex)
	}

	x := gp.GetX()

	// get lineHeight
	itext, err := gp.curr.FontISubset.AddChars(text)
	if err != nil {
		return err
	}
	_, lineHeight, _, err := createContent(gp.curr.FontISubset, itext, gp.curr.FontSize, gp.curr.CharSpacing, nil)
	if err != nil {
		return err
	}
	gp.PointsToUnitsVar(&lineHeight)

	textSplits, err := gp.SplitTextWithOption(text, rectangle.W, opt.BreakOption)
	if err != nil {
		return err
	}

	// Última línea no se justifica normalmente
	lastLineIndex := len(textSplits) - 1

	y := gp.GetY()
	for i, text := range textSplits {
		// Solo justificar si:
		// 1. Se solicita justificación
		// 2. No es la última línea (o es la única)
		// 3. Tiene contenido
		// 4. Tiene al menos un espacio para ajustar
		// 5. La línea ocupa al menos el 70% del ancho disponible (para evitar estirar demasiado pocas palabras)
		shouldJustify := isJustify &&
			i != lastLineIndex &&
			len(strings.TrimSpace(text)) > 0 &&
			strings.Contains(text, " ")

		// Comprobamos que la línea ocupe suficiente espacio horizontal para justificarla
		if shouldJustify {
			lineWidth, _ := gp.MeasureTextWidth(text)
			shouldJustify = lineWidth >= (rectangle.W * 0.7)
		}

		beforeY := gp.GetY()

		if shouldJustify {
			// Procesar para justificación
			jText, err := parseTextForJustification(gp, text, rectangle.W)
			if err != nil {
				return err
			}

			err = drawJustifiedLine(gp, jText, x, y)
			if err != nil {
				return err
			}
		} else {
			// Usar el método normal para alineación no justificada o última línea
			gp.CellWithOption(&Rect{W: rectangle.W, H: lineHeight}, string(text), opt)

			// Reset Y position to ensure consistent behavior with justified text
			// CellWithOption advances Y, so we need to undo that advancement
			gp.SetY(beforeY)
		}

		// Use consistent line spacing for both justified and non-justified text
		// Only apply Br if this isn't the last line
		if i < len(textSplits)-1 {
			gp.Br(lineHeight)
		}

		gp.SetX(x)
		y = gp.GetY()
	}

	return nil
}

// SplitText splits text into multiple lines based on width performing potential mid-word breaks.
func (gp *GoPdf) SplitText(text string, width float64) ([]string, error) {
	return gp.SplitTextWithOption(text, width, &DefaultBreakOption)
}

// SplitTextWithWordWrap behaves the same way SplitText does but performs a word-wrap considering spaces in case
// a text line split would split a word.
func (gp *GoPdf) SplitTextWithWordWrap(text string, width float64) ([]string, error) {
	return gp.SplitTextWithOption(text, width, &BreakOption{
		Mode:           BreakModeIndicatorSensitive,
		BreakIndicator: ' ',
	})
}

// SplitTextWithOption splits a text into multiple lines based on the current font size of the document.
// BreakOptions allow to define the behavior of the split (strict or sensitive). For more information see BreakOption.
func (gp *GoPdf) SplitTextWithOption(text string, width float64, opt *BreakOption) ([]string, error) {
	// fallback to default break option
	if opt == nil {
		opt = &DefaultBreakOption
	}
	var lineText []rune
	var lineTexts []string
	utf8Texts := []rune(text)
	utf8TextsLen := len(utf8Texts) // utf8 string quantity
	if utf8TextsLen == 0 {
		return lineTexts, errEmptyString
	}
	separatorWidth, err := gp.MeasureTextWidth(opt.Separator)
	if err != nil {
		return nil, err
	}
	// possible (not conflicting) position of the separator within the currently processed line
	separatorIdx := 0
	for i := 0; i < utf8TextsLen; i++ {
		lineWidth, err := gp.MeasureTextWidth(string(lineText))
		if err != nil {
			return nil, err
		}
		runeWidth, err := gp.MeasureTextWidth(string(utf8Texts[i]))
		if err != nil {
			return nil, err
		}
		// mid-word break required since the max width of the given rect is exceeded
		if lineWidth+runeWidth > width && utf8Texts[i] != '\n' {
			// forceBreak will be set to true in case an indicator sensitive break was not possible which will cause
			// strict break to not exceed the desired width
			forceBreak := false
			if opt.Mode == BreakModeIndicatorSensitive {
				forceBreak = !performIndicatorSensitiveLineBreak(&lineTexts, &lineText, &i, opt)
			}
			// BreakModeStrict breaks immediately with an optionally available separator
			if opt.Mode == BreakModeStrict || forceBreak {
				performStrictLineBreak(&lineTexts, &lineText, &i, separatorIdx, opt)
			}
			continue
		}
		// regular break due to a new line rune
		if utf8Texts[i] == '\n' {
			lineTexts = append(lineTexts, string(lineText))
			lineText = lineText[0:0]
			continue
		}
		// end of text
		if i == utf8TextsLen-1 {
			lineText = append(lineText, utf8Texts[i])
			lineTexts = append(lineTexts, string(lineText))
		}
		// store overall index when separator would still fit in the currently processed text-line
		if opt.HasSeparator() && lineWidth+runeWidth+separatorWidth <= width {
			separatorIdx = i
		}
		lineText = append(lineText, utf8Texts[i])
	}
	return lineTexts, nil
}

// performIndicatorSensitiveLineBreak - función auxiliar para SplitTextWithOption
// Intenta realizar un salto de línea sensible al indicador de ruptura (típicamente espacios en blanco)
func performIndicatorSensitiveLineBreak(lineTexts *[]string, lineText *[]rune, i *int, opt *BreakOption) bool {
	brIdx := breakIndicatorIndex(*lineText, opt.BreakIndicator)
	if brIdx > 0 {
		diff := len(*lineText) - brIdx
		*lineText = (*lineText)[0:brIdx]
		*lineTexts = append(*lineTexts, string(*lineText))
		*lineText = (*lineText)[0:0]
		*i -= diff
		return true
	}
	return false
}

// performStrictLineBreak - función auxiliar para SplitTextWithOption
// Realiza un salto de línea estricto, posiblemente agregando un separador
func performStrictLineBreak(lineTexts *[]string, lineText *[]rune, i *int, separatorIdx int, opt *BreakOption) {
	if opt.HasSeparator() && separatorIdx > -1 {
		// trim the line to the last possible index with an appended separator
		trimIdx := *i - separatorIdx
		*lineText = (*lineText)[0 : len(*lineText)-trimIdx]
		// append separator to the line
		*lineText = append(*lineText, []rune(opt.Separator)...)
		*lineTexts = append(*lineTexts, string(*lineText))
		*lineText = (*lineText)[0:0]
		*i = separatorIdx - 1
		return
	}
	*lineTexts = append(*lineTexts, string(*lineText))
	*lineText = (*lineText)[0:0]
	*i--
}

// breakIndicatorIndex - función auxiliar para SplitTextWithOption
// breakIndicatorIndex returns the index where a text line (i.e. rune slice) can be split "gracefully" by checking on
// the break indicator.
// In case no possible break can be identified -1 is returned.
func breakIndicatorIndex(text []rune, bi rune) int {
	for i := len(text) - 1; i > 0; i-- {
		if text[i] == bi {
			return i
		}
	}
	return -1
}
