package tinypdf

// DocTable represents a table to be added to the document
type DocTable struct {
	doc          *Document
	columns      []tableColumn
	rows         [][]tableCell
	width        float64
	rowHeight    float64
	cellPadding  float64
	headerStyle  CellStyle
	cellStyle    CellStyle
	alignment    int // Left, Center, or Right alignment
	currentWidth float64
}

// tableColumn represents a column in the table
type tableColumn struct {
	header string  // Header text
	width  float64 // Width of the column
	align  int     // Text alignment within cells (Left, Center, Right)
}

// tableCell represents a cell in the table
type tableCell struct {
	content      string    // Content of the cell
	useCellStyle bool      // If true, use custom style instead of the table's default
	cellStyle    CellStyle // Custom style for this cell
}

// NewTable creates a new table with the specified headers
// It automatically determines column widths based on header text
func (doc *Document) NewTable(headers ...string) *DocTable {
	table := &DocTable{
		doc:         doc,
		rowHeight:   25, // Default row height
		cellPadding: 5,  // Default padding
		alignment:   Center,
		width:       0, // Will be calculated based on headers or set to document width
	}

	// Set default styles based on document font configuration
	table.headerStyle = CellStyle{
		BorderStyle: BorderStyle{
			Top:      true,
			Left:     true,
			Bottom:   true,
			Right:    true,
			Width:    1.0,
			RGBColor: RGBColor{R: 100, G: 100, B: 100},
		},
		FillColor: RGBColor{R: 240, G: 240, B: 240},
		TextColor: RGBColor{R: 0, G: 0, B: 0},
		Font:      FontBold,
		FontSize:  doc.fontConfig.Header3.Size,
	}

	table.cellStyle = CellStyle{
		BorderStyle: BorderStyle{
			Top:      true,
			Left:     true,
			Bottom:   true,
			Right:    true,
			Width:    0.5,
			RGBColor: RGBColor{R: 200, G: 200, B: 200},
		},
		FillColor: RGBColor{R: 255, G: 255, B: 255},
		TextColor: RGBColor{R: 0, G: 0, B: 0},
		Font:      FontRegular,
		FontSize:  doc.fontConfig.Normal.Size,
	}

	// Estimate widths for columns based on header text
	columns := make([]tableColumn, len(headers))

	// Calculate total table width
	doc.SetFont(FontBold, "", doc.fontConfig.Header3.Size)
	totalWidth := 0.0

	// First pass: Calculate minimum width for each column based on header text
	for i, header := range headers {
		// Estimate width based on header text - MEJORADO para evitar cortes de texto
		textWidthFactor := 0.85 // Usar un factor más generoso que el de measureTextWidthFactor
		estWidth := float64(len(header)) * doc.fontConfig.Header3.Size * textWidthFactor

		// Add padding to ensure header fits comfortably
		minWidth := estWidth + (table.cellPadding * 2)

		// Ensure minimum reasonable width and add extra space for encabezados
		if minWidth < 40 {
			minWidth = 40 // Aumentado de 30 a 40 para dar más espacio
		}

		columns[i] = tableColumn{
			header: header,
			width:  minWidth,
			align:  Center, // Cambiado a Center para centrar los encabezados por defecto
		}

		totalWidth += minWidth
	}

	// If total width exceeds document width, scale down proportionally
	maxWidth := doc.pageWidth - (doc.margins.Left + doc.margins.Right)
	if totalWidth > maxWidth {
		scaleFactor := maxWidth / totalWidth
		for i := range columns {
			columns[i].width *= scaleFactor
		}
		totalWidth = maxWidth
	}

	table.columns = columns
	table.width = totalWidth
	table.currentWidth = totalWidth

	return table
}

// Width sets the total width of the table
func (t *DocTable) Width(width float64) *DocTable {
	if width > 0 {
		scaleFactor := width / t.currentWidth
		for i := range t.columns {
			t.columns[i].width *= scaleFactor
		}
		t.width = width
		t.currentWidth = width
	}
	return t
}

// RowHeight sets the height of rows
func (t *DocTable) RowHeight(height float64) *DocTable {
	if height > 0 {
		t.rowHeight = height
	}
	return t
}

// SetColumnWidth sets the width of a specific column by index
func (t *DocTable) SetColumnWidth(columnIndex int, width float64) *DocTable {
	if columnIndex >= 0 && columnIndex < len(t.columns) && width > 0 {
		// Adjust total width
		t.currentWidth = t.currentWidth - t.columns[columnIndex].width + width
		t.columns[columnIndex].width = width
	}
	return t
}

// SetColumnAlignment sets the text alignment for a specific column
func (t *DocTable) SetColumnAlignment(columnIndex int, alignment int) *DocTable {
	if columnIndex >= 0 && columnIndex < len(t.columns) {
		t.columns[columnIndex].align = alignment
	}
	return t
}

// AlignLeft aligns the table to the left margin
func (t *DocTable) AlignLeft() *DocTable {
	t.alignment = Left
	return t
}

// AlignCenter centers the table horizontally (default)
func (t *DocTable) AlignCenter() *DocTable {
	t.alignment = Center
	return t
}

// AlignRight aligns the table to the right margin
func (t *DocTable) AlignRight() *DocTable {
	t.alignment = Right
	return t
}

// HeaderStyle sets the style for the header row
func (t *DocTable) HeaderStyle(style CellStyle) *DocTable {
	t.headerStyle = style
	return t
}

// CellStyle sets the default style for regular cells
func (t *DocTable) CellStyle(style CellStyle) *DocTable {
	t.cellStyle = style
	return t
}

// AddRow adds a row of data to the table
func (t *DocTable) AddRow(cells ...string) *DocTable {
	rowCells := make([]tableCell, len(cells))
	for i, content := range cells {
		rowCells[i] = tableCell{
			content:      content,
			useCellStyle: false,
		}
	}
	t.rows = append(t.rows, rowCells)
	return t
}

// AddStyledRow adds a row with individually styled cells
func (t *DocTable) AddStyledRow(cells ...StyledCell) *DocTable {
	rowCells := make([]tableCell, len(cells))
	for i, cell := range cells {
		rowCells[i] = tableCell{
			content:      cell.Content,
			useCellStyle: true,
			cellStyle:    cell.Style,
		}
	}
	t.rows = append(t.rows, rowCells)
	return t
}

// StyledCell represents a cell with custom styling
type StyledCell struct {
	Content string
	Style   CellStyle
}

// NewStyledCell creates a new cell with custom styling
func (doc *Document) NewStyledCell(content string, style CellStyle) StyledCell {
	return StyledCell{
		Content: content,
		Style:   style,
	}
}

// Draw renders the table on the document
func (t *DocTable) Draw() error {
	// Calculate total height of table
	totalHeight := t.rowHeight * float64(len(t.rows)+1) // +1 for header row

	// Check if table fits on current page
	y, _ := t.doc.ensureElementFits(totalHeight, t.doc.fontConfig.Normal.SpaceAfter)

	// Calculate starting X position based on alignment
	x := t.doc.margins.Left
	switch t.alignment {
	case Center:
		x = t.doc.margins.Left + (t.doc.pageWidth-(t.doc.margins.Left+t.doc.margins.Right)-t.width)/2
	case Right:
		x = t.doc.margins.Left + (t.doc.pageWidth - (t.doc.margins.Left + t.doc.margins.Right) - t.width)
	}

	// Colección para guardar información de los encabezados para dibujar sus bordes al final
	type headerInfo struct {
		x, y, width, height float64
		style               CellStyle
	}
	headers := []headerInfo{}

	// Draw header row (sin bordes por ahora, solo contenido y fondo)
	currentX := x
	for _, col := range t.columns {
		// Guardamos información del encabezado para dibujar los bordes después
		headers = append(headers, headerInfo{
			x:      currentX,
			y:      y,
			width:  col.width,
			height: t.rowHeight,
			style:  t.headerStyle,
		})

		// Dibujar contenido y fondo de la celda de encabezado
		t.drawCellContent(
			currentX,
			y,
			col.width,
			t.rowHeight,
			col.header,
			col.align,
			true, // isHeader
			t.headerStyle,
		)
		currentX += col.width
	}

	// Draw data rows
	for rowIndex, row := range t.rows {
		currentY := y + ((float64(rowIndex) + 1) * t.rowHeight)
		currentX = x

		// Check if this row fits on the current page
		if currentY+t.rowHeight > t.doc.curr.pageSize.H-t.doc.margins.Bottom {
			t.doc.AddPage()
			currentY = t.doc.margins.Top

			// Limpiar la lista de encabezados para la nueva página
			headers = []headerInfo{}

			// Redraw the header row on the new page (solo contenido y fondo)
			headerY := currentY
			headerX := x
			for _, col := range t.columns {
				// Guardamos información del encabezado para dibujar los bordes después
				headers = append(headers, headerInfo{
					x:      headerX,
					y:      headerY,
					width:  col.width,
					height: t.rowHeight,
					style:  t.headerStyle,
				})

				// Dibujar contenido y fondo de la celda de encabezado
				t.drawCellContent(
					headerX,
					headerY,
					col.width,
					t.rowHeight,
					col.header,
					col.align,
					true, // isHeader
					t.headerStyle,
				)
				headerX += col.width
			}

			// Adjust currentY to start below the header
			currentY += t.rowHeight
		}

		for colIndex, cell := range row {
			// Use column width, handle case where row has fewer cells than columns
			if colIndex < len(t.columns) {
				cellWidth := t.columns[colIndex].width
				cellAlign := t.columns[colIndex].align

				// Determine which style to use
				style := t.cellStyle
				if cell.useCellStyle {
					style = cell.cellStyle
				}

				// Dibujar celda completa (contenido, fondo y bordes)
				t.drawCell(
					currentX,
					currentY,
					cellWidth,
					t.rowHeight,
					cell.content,
					cellAlign,
					false, // not header
					style,
				)

				currentX += cellWidth
			}
		}
	}

	// Ahora dibujamos los bordes de los encabezados al final
	for _, h := range headers {
		t.drawCellBorder(h.x, h.y, h.width, h.height, h.style.BorderStyle)
	}

	// Update document position to after the table
	t.doc.SetY(y + totalHeight + t.doc.fontConfig.Normal.SpaceAfter)

	return nil
}

// drawCellContent dibuja solo el contenido y fondo de una celda (sin bordes)
func (t *DocTable) drawCellContent(
	x float64,
	y float64,
	width float64,
	height float64,
	content string,
	align int,
	isHeader bool,
	style CellStyle,
) {
	// Fill the cell background if a fill color is specified
	if (style.FillColor != RGBColor{}) {
		t.doc.SetFillColor(style.FillColor.R, style.FillColor.G, style.FillColor.B)
		t.doc.RectFromUpperLeftWithStyle(x, y, width, height, "F")
	}

	// Set text properties
	if style.Font != "" {
		t.doc.SetFont(style.Font, "", style.FontSize)
	}
	t.doc.SetTextColor(style.TextColor.R, style.TextColor.G, style.TextColor.B)

	// Calculate text position with padding
	textX := x + t.cellPadding
	textY := y + t.cellPadding
	textWidth := width - (2 * t.cellPadding)
	textHeight := height - (2 * t.cellPadding)

	// Create cell options
	cellOpt := CellOption{
		Align:  align | Middle, // Combine horizontal alignment with vertical middle alignment
		Border: 0,              // No borders
		BreakOption: &BreakOption{
			Mode:           BreakModeIndicatorSensitive,
			BreakIndicator: ' ',
		},
	}

	// Draw the cell content
	t.doc.SetXY(textX, textY)
	err := t.doc.MultiCellWithOption(&Rect{W: textWidth, H: textHeight}, content, cellOpt)
	if err != nil && err.Error() != "empty string" {
		t.doc.log("Error drawing table cell:", err)
	}
}

// drawCellBorder dibuja solo los bordes de una celda
func (t *DocTable) drawCellBorder(
	x float64,
	y float64,
	width float64,
	height float64,
	borderStyle BorderStyle,
) {
	if borderStyle.Width > 0 {
		t.doc.SetLineWidth(borderStyle.Width)
		t.doc.SetStrokeColor(
			borderStyle.RGBColor.R,
			borderStyle.RGBColor.G,
			borderStyle.RGBColor.B,
		)

		if borderStyle.Top {
			t.doc.Line(x, y, x+width, y)
		}
		if borderStyle.Bottom {
			t.doc.Line(x, y+height, x+width, y+height)
		}
		if borderStyle.Left {
			t.doc.Line(x, y, x, y+height)
		}
		if borderStyle.Right {
			t.doc.Line(x+width, y, x+width, y+height)
		}
	}
}

// drawCell draws a single cell of the table
func (t *DocTable) drawCell(
	x float64,
	y float64,
	width float64,
	height float64,
	content string,
	align int,
	isHeader bool,
	style CellStyle,
) {
	// Primero dibujamos el contenido y el fondo
	t.drawCellContent(x, y, width, height, content, align, isHeader, style)

	// Luego dibujamos los bordes
	t.drawCellBorder(x, y, width, height, style.BorderStyle)
}
