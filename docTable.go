package tinypdf

import "strconv"

// docTable represents a table to be added to the document
type docTable struct {
	doc          *Document
	columns      []tableColumn
	rows         [][]tableCell
	width        float64
	rowHeight    float64
	cellPadding  float64
	headerStyle  CellStyle
	cellStyle    CellStyle
	alignment    position // Left, Center, or Right alignment
	currentWidth float64
}

// tableColumn represents a column in the table
type tableColumn struct {
	header      string   // Header text
	width       float64  // Width of the column
	headerAlign position // Text alignment for header (Left, Center, Right)
	align       position // Text alignment within cells (Left, Center, Right)
	prefix      string   // Prefix to add before each value in the column
	suffix      string   // Suffix to add after each value in the column
}

// tableCell represents a cell in the table
type tableCell struct {
	content      string    // Content of the cell
	useCellStyle bool      // If true, use custom style instead of the table's default
	cellStyle    CellStyle // Custom style for this cell
}

// NewTable creates a new table with the specified headers
// Headers can include formatting options in the following format.
//
// Format: "headerTitle|option1,option2,option3,..."
//
// Available options:
//   - Header alignment: "HL" (left), "HC" (center, default), "HR" (right)
//   - Column alignment: "CL" (left, default), "CC" (center), "CR" (right)
//   - Prefix/Suffix: "P:value" (adds prefix), "S:value" (adds suffix)
//   - Width: "W:number" (fixed width), "W:number%" (percentage width)
//     Note: If width is not specified, auto width is used by default
//
// Examples:
//   - "Name" - Normal header, left-aligned column, auto width
//   - "Price|HR,CR" - Right-aligned header, right-aligned column
//   - "Price|HR,CR,P:$" - Right-aligned header, right-aligned column with "$" prefix
//   - "Percentage|HC,CC,S:%" - Center-aligned header, center-aligned column with "%" suffix
//   - "Name|HL,CL,W:30%" - Left-aligned header, left-aligned column with 30% of available width
//   - "Age|HC,CR,W:20" - Center-aligned header, right-aligned column with fixed width of 20 units
func (doc *Document) NewTable(headers ...string) *docTable {
	table := &docTable{
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

	// Parse headers and estimate widths for columns
	columns := make([]tableColumn, len(headers))

	// Calculate total table width
	doc.SetFont(FontBold, "", doc.fontConfig.Header3.Size)

	// Track columns with fixed width and percentage width
	fixedWidthTotal := 0.0
	percentageWidthTotal := 0.0
	hasPercentageWidths := false

	// Available width for the table (accounting for margins)
	maxWidth := doc.pageWidth - (doc.margins.Left + doc.margins.Right)

	// Apply a small margin to account for spacing between elements
	// This ensures the table doesn't stretch all the way to the edges
	tableMargin := 10.0 // Aumentamos el margen a cada lado para evitar que se acerque demasiado a los bordes
	maxWidth = maxWidth - (2 * tableMargin)

	// First pass: Calculate minimum width for each column based on header text
	for i, headerFormat := range headers {
		// Parse the header format
		options := parseHeaderFormat(headerFormat)

		// Determine the width based on the format options or estimate from header text
		var colWidth float64

		// Check if a specific width was provided
		if options.WidthType == "fixed" {
			// Fixed width specified in units
			colWidth = options.Width
			fixedWidthTotal += colWidth
		} else if options.WidthType == "percent" {
			// Percentage of available width
			hasPercentageWidths = true
			percentageWidthTotal += options.Width
			// Store the percentage for now, will calculate actual width later
			colWidth = options.Width
		} else {
			// Auto width based on content (default behavior)
			// Estimate width based on header text
			textWidthFactor := 0.85
			estWidth := float64(len(options.HeaderTitle)) * doc.fontConfig.Header3.Size * textWidthFactor

			// Add padding to ensure header fits comfortably
			colWidth = estWidth + (table.cellPadding * 2)

			// Ensure minimum reasonable width
			if colWidth < 40 {
				colWidth = 40
			}

			fixedWidthTotal += colWidth
		}

		columns[i] = tableColumn{
			header:      options.HeaderTitle,
			width:       colWidth, // Will be adjusted in second pass for percentage widths
			headerAlign: options.HeaderAlignment,
			align:       options.ColumnAlignment,
			prefix:      options.Prefix,
			suffix:      options.Suffix,
		}
	}

	// Second pass: Adjust widths for columns with percentage specifications
	if hasPercentageWidths {
		// Si hay anchos porcentuales, siempre usar el ancho máximo disponible
		totalTableWidth := maxWidth

		// Calcular el ancho mínimo requerido basado en los porcentajes especificados
		minimumWidthRequired := 0.0
		for i := range columns {
			options := parseHeaderFormat(headers[i])
			if options.WidthType == "percent" {
				// Aplicar el porcentaje mínimo especificado
				columns[i].width = (maxWidth * options.Width) / 100.0
				minimumWidthRequired += columns[i].width
			}
		}

		// Calcular el espacio restante después de considerar columnas de ancho fijo
		nonPercentageWidth := fixedWidthTotal
		remainingWidth := totalTableWidth - nonPercentageWidth - minimumWidthRequired

		// Si hay espacio adicional disponible, distribuirlo uniformemente entre las columnas porcentuales
		if remainingWidth > 0 {
			// Contar columnas porcentuales para distribución uniforme del espacio restante
			percentageColumnsCount := 0
			for i := range columns {
				options := parseHeaderFormat(headers[i])
				if options.WidthType == "percent" {
					percentageColumnsCount++
				}
			}

			// Distribuir el espacio restante uniformemente
			extraWidthPerColumn := remainingWidth / float64(percentageColumnsCount)
			for i := range columns {
				options := parseHeaderFormat(headers[i])
				if options.WidthType == "percent" {
					// Agregar el espacio extra a cada columna porcentual
					columns[i].width += extraWidthPerColumn
				}
			}
		} else if remainingWidth < 0 {
			// Si no hay suficiente espacio, escalar las columnas fijas proporcionalmente
			scaleFactor := (totalTableWidth - minimumWidthRequired) / nonPercentageWidth
			if scaleFactor > 0 && scaleFactor < 1 {
				for i := range columns {
					options := parseHeaderFormat(headers[i])
					if options.WidthType != "percent" {
						columns[i].width *= scaleFactor
					}
				}
			}
		}

		// Establecer el ancho total de la tabla al ancho máximo disponible
		table.width = totalTableWidth
	} else {
		// Calculate total width after all adjustments
		totalWidth := 0.0
		for _, col := range columns {
			totalWidth += col.width
		}

		// If total width exceeds document width, scale down proportionally
		if totalWidth > maxWidth {
			scaleFactor := maxWidth / totalWidth
			for i := range columns {
				columns[i].width *= scaleFactor
			}
			totalWidth = maxWidth
		}

		table.width = totalWidth
	}

	table.columns = columns
	table.currentWidth = table.width

	return table
}

// Width sets the total width of the table
func (t *docTable) Width(width float64) *docTable {
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
func (t *docTable) RowHeight(height float64) *docTable {
	if height > 0 {
		t.rowHeight = height
	}
	return t
}

// SetColumnWidth sets the width of a specific column by index
func (t *docTable) SetColumnWidth(columnIndex int, width float64) *docTable {
	if columnIndex >= 0 && columnIndex < len(t.columns) && width > 0 {
		// Adjust total width
		t.currentWidth = t.currentWidth - t.columns[columnIndex].width + width
		t.columns[columnIndex].width = width
	}
	return t
}

// SetHeaderAlignment sets the text alignment for a specific header
func (t *docTable) SetHeaderAlignment(columnIndex int, alignment position) *docTable {
	if columnIndex >= 0 && columnIndex < len(t.columns) {
		t.columns[columnIndex].headerAlign = alignment
	}
	return t
}

// SetColumnPrefix sets a prefix for all values in a column
func (t *docTable) SetColumnPrefix(columnIndex int, prefix string) *docTable {
	if columnIndex >= 0 && columnIndex < len(t.columns) {
		t.columns[columnIndex].prefix = prefix
	}
	return t
}

// SetColumnSuffix sets a suffix for all values in a column
func (t *docTable) SetColumnSuffix(columnIndex int, suffix string) *docTable {
	if columnIndex >= 0 && columnIndex < len(t.columns) {
		t.columns[columnIndex].suffix = suffix
	}
	return t
}

// AlignLeft aligns the table to the left margin
func (t *docTable) AlignLeft() *docTable {
	t.alignment = Left
	return t
}

// AlignCenter centers the table horizontally (default)
func (t *docTable) AlignCenter() *docTable {
	t.alignment = Center
	return t
}

// AlignRight aligns the table to the right margin
func (t *docTable) AlignRight() *docTable {
	t.alignment = Right
	return t
}

// HeaderStyle sets the style for the header row
func (t *docTable) HeaderStyle(style CellStyle) *docTable {
	t.headerStyle = style
	return t
}

// CellStyle sets the default style for regular cells
func (t *docTable) CellStyle(style CellStyle) *docTable {
	t.cellStyle = style
	return t
}

// anyToString converts any value to a string without using fmt
// Uses TinyGo-compatible approach to convert numbers to strings
func (t *docTable) anyToString(v any) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.FormatInt(int64(val), 10)
	case int8:
		return strconv.FormatInt(int64(val), 10)
	case int16:
		return strconv.FormatInt(int64(val), 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case uint8:
		return strconv.FormatUint(uint64(val), 10)
	case uint16:
		return strconv.FormatUint(uint64(val), 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		// For any other type, return empty string
		// In a full implementation, you might want to handle more types
		t.doc.log("docTable anyToString error Unsupported type for conversion to string:", val)
		return ""
	}
}

// AddRow adds a row of data to the table
// Accepts any value type which will be converted to strings
func (t *docTable) AddRow(cells ...any) *docTable {
	rowCells := make([]tableCell, len(cells))
	for i, content := range cells {
		formattedContent := t.anyToString(content)

		// Apply prefix and suffix if column exists
		if i < len(t.columns) {
			if t.columns[i].prefix != "" {
				formattedContent = t.columns[i].prefix + formattedContent
			}
			if t.columns[i].suffix != "" {
				formattedContent = formattedContent + t.columns[i].suffix
			}
		}

		rowCells[i] = tableCell{
			content:      formattedContent,
			useCellStyle: false,
		}
	}
	t.rows = append(t.rows, rowCells)
	return t
}

// AddStyledRow adds a row with individually styled cells
func (t *docTable) AddStyledRow(cells ...StyledCell) *docTable {
	rowCells := make([]tableCell, len(cells))
	for i, cell := range cells {
		formattedContent := cell.Content

		// Apply prefix and suffix if column exists
		if i < len(t.columns) {
			if t.columns[i].prefix != "" {
				formattedContent = t.columns[i].prefix + formattedContent
			}
			if t.columns[i].suffix != "" {
				formattedContent = formattedContent + t.columns[i].suffix
			}
		}

		rowCells[i] = tableCell{
			content:      formattedContent,
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
func (t *docTable) Draw() error {
	// Calculate total height of table
	totalHeight := t.rowHeight * float64(len(t.rows)+1) // +1 for header row

	// Check if table fits on current page
	y, _ := t.doc.ensureElementFits(totalHeight, t.doc.fontConfig.Normal.SpaceAfter)

	// Verificar que la tabla no exceda el ancho disponible
	availableWidth := t.doc.pageWidth - (t.doc.margins.Left + t.doc.margins.Right)

	// Aplicar un padding lateral para que la tabla no quede pegada a los márgenes
	tablePadding := 10.0 // Padding a cada lado
	availableWidth -= (2 * tablePadding)

	// Si el ancho de la tabla excede el disponible, ajustar proporcionalmente
	if t.width > availableWidth {
		scaleFactor := availableWidth / t.width
		for i := range t.columns {
			t.columns[i].width *= scaleFactor
		}
		t.width = availableWidth
		t.currentWidth = availableWidth
	}

	// Calculate starting X position using the new positioning method
	x := t.calculatePosition()

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

		// Dibujar contenido y fondo de la celda de encabezado usando la alineación especificada
		t.drawCellContent(
			currentX,
			y,
			col.width,
			t.rowHeight,
			col.header,
			col.headerAlign, // Use header-specific alignment
			true,            // isHeader
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

				// Dibujar contenido y fondo de la celda de encabezado usando la alineación especificada
				t.drawCellContent(
					headerX,
					headerY,
					col.width,
					t.rowHeight,
					col.header,
					col.headerAlign, // Use header-specific alignment
					true,            // isHeader
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
func (t *docTable) drawCellContent(
	x float64,
	y float64,
	width float64,
	height float64,
	content string,
	align position,
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
func (t *docTable) drawCellBorder(
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
func (t *docTable) drawCell(
	x float64,
	y float64,
	width float64,
	height float64,
	content string,
	align position,
	isHeader bool,
	style CellStyle,
) {
	// Primero dibujamos el contenido y el fondo
	t.drawCellContent(x, y, width, height, content, align, isHeader, style)

	// Luego dibujamos los bordes
	t.drawCellBorder(x, y, width, height, style.BorderStyle)
}

// calculatePosition determina donde colocar la tabla
func (t *docTable) calculatePosition() float64 {
	x := t.doc.margins.Left

	// Aplicar alineación de manera consistente con docImage.go
	// Usar todo el ancho de página disponible para el cálculo
	availableWidth := t.doc.pageWidth
	switch t.alignment {
	case Center:
		x = t.doc.margins.Left + (availableWidth-t.width)/2
	case Right:
		x = t.doc.margins.Left + availableWidth - t.width
	}

	return x
}
