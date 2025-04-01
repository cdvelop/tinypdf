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
