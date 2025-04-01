package tinypdf

import (
	"strconv"
	"strings"
)

// tableHeaderFormat represents the formatting options for a table header
type tableHeaderFormat struct {
	HeaderTitle     string   // The displayed title text
	HeaderAlignment position // Alignment of the header text (Left, Center, Right)
	ColumnAlignment position // Alignment of the column values (Left, Center, Right)
	Prefix          string   // Prefix to add before each value in the column
	Suffix          string   // Suffix to add after each value in the column
	Width           float64  // Width of the column (0 = auto)
	WidthType       string   // Type of width: "fixed", "percent", or "auto"
}

// parseHeaderFormat parses a header string with format options
// Format: "headerTitle|option1,option2,option3,..."
//
// Examples:
//
//	"Name" -> {HeaderTitle: "Name", HeaderAlignment: Center, ColumnAlignment: Left}
//	"Price|HR,CR" -> {HeaderTitle: "Price", HeaderAlignment: Right, ColumnAlignment: Right}
//	"Amount|HR,CR,P:$" -> {HeaderTitle: "Amount", HeaderAlignment: Right, ColumnAlignment: Right, Prefix: "$"}
//	"Percent|HC,CC,S:%" -> {HeaderTitle: "Percent", HeaderAlignment: Center, ColumnAlignment: Center, Suffix: "%"}
//	"Name|HL,CL,W:30%" -> {HeaderTitle: "Name", HeaderAlignment: Left, ColumnAlignment: Left, Width: 30, WidthType: "percent"}
func parseHeaderFormat(headerStr string) tableHeaderFormat {
	result := tableHeaderFormat{
		HeaderAlignment: Center, // Default header alignment is center
		ColumnAlignment: Left,   // Default column alignment is left
		WidthType:       "auto", // Default width type is auto
	}

	// Split by the separator character
	parts := strings.Split(headerStr, "|")
	result.HeaderTitle = parts[0]

	// If there are no format options, return with defaults
	if len(parts) < 2 || parts[1] == "" {
		return result
	}

	// Get the format options
	formatOptions := parts[1]

	// Process each option separately
	options := strings.Split(formatOptions, ",")
	for _, option := range options {
		option = strings.TrimSpace(option)

		// Process header alignment
		if strings.HasPrefix(option, "H") {
			if option == "HL" {
				result.HeaderAlignment = Left
			} else if option == "HR" {
				result.HeaderAlignment = Right
			} else if option == "HC" {
				result.HeaderAlignment = Center
			}
		}

		// Process column alignment
		if strings.HasPrefix(option, "C") {
			if option == "CL" {
				result.ColumnAlignment = Left
			} else if option == "CR" {
				result.ColumnAlignment = Right
			} else if option == "CC" {
				result.ColumnAlignment = Center
			}
		}

		// Process prefix
		if strings.HasPrefix(option, "P:") {
			result.Prefix = option[2:]
		}

		// Process suffix
		if strings.HasPrefix(option, "S:") {
			result.Suffix = option[2:]
		}

		// Process width
		if strings.HasPrefix(option, "W:") {
			widthStr := option[2:]
			if strings.HasSuffix(widthStr, "%") {
				result.WidthType = "percent"
				// Parse percentage value (remove the % sign)
				percentStr := widthStr[:len(widthStr)-1]
				if val, err := strconv.ParseFloat(percentStr, 64); err == nil {
					result.Width = val
				}
			} else {
				// Si no tiene %, se asume un ancho fijo
				result.WidthType = "fixed"
				if val, err := strconv.ParseFloat(widthStr, 64); err == nil {
					result.Width = val
				}
			}
		}
	}

	return result
}

// createDefaultTableStyles crea los estilos predeterminados para una tabla basados en la configuración del documento
// Devuelve los estilos para el encabezado y las celdas normales
func createDefaultTableStyles(doc *Document) (headerStyle CellStyle, cellStyle CellStyle) {
	// Estilo para el encabezado
	headerStyle = CellStyle{
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

	// Estilo para las celdas normales
	cellStyle = CellStyle{
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

	return headerStyle, cellStyle
}

// calculateAutoColumnWidth estima el ancho necesario para una columna basado en el texto del encabezado
func calculateAutoColumnWidth(headerTitle string, fontSize float64, cellPadding float64) float64 {
	textWidthFactor := 0.85
	estWidth := float64(len(headerTitle)) * fontSize * textWidthFactor

	// Añadir padding a ambos lados
	colWidth := estWidth + (cellPadding * 2)

	// Asegurar un ancho mínimo razonable
	if colWidth < 40 {
		colWidth = 40
	}

	return colWidth
}

// parseTableHeaders procesa los encabezados de la tabla y calcula los anchos iniciales
// Devuelve las columnas configuradas, el total de anchos fijos y si hay columnas con porcentajes
func parseTableHeaders(doc *Document, headers []string, cellPadding float64, maxWidth float64) (
	columns []tableColumn,
	fixedWidthTotal float64,
	hasPercentageWidths bool) {

	columns = make([]tableColumn, len(headers))

	// Configurar la fuente para estimaciones de ancho
	doc.SetFont(FontBold, "", doc.fontConfig.Header3.Size)

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
			// Store the percentage for now, will calculate actual width later
			colWidth = options.Width
		} else {
			// Auto width based on content (default behavior)
			colWidth = calculateAutoColumnWidth(options.HeaderTitle, doc.fontConfig.Header3.Size, cellPadding)
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

	return columns, fixedWidthTotal, hasPercentageWidths
}

// adjustPercentageWidths ajusta los anchos de las columnas que se especificaron como porcentajes
// y devuelve el ancho total final de la tabla
func adjustPercentageWidths(
	columns []tableColumn,
	headers []string,
	hasPercentageWidths bool,
	maxWidth float64,
	fixedWidthTotal float64) float64 {

	if !hasPercentageWidths {
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

		return totalWidth
	}

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
	return totalTableWidth
}
