package tinypdf

import (
	"testing"
)

func TestParseHeaderFormat(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedResult headerFormatOptions
	}{
		{
			name:  "Simple header without options",
			input: "Name",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Name",
				HeaderAlignment: Center,
				ColumnAlignment: Left,
				Prefix:          "",
				Suffix:          "",
				Width:           0,
				WidthType:       "auto",
			},
		},
		{
			name:  "Header with left alignment",
			input: "Name|HL",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Name",
				HeaderAlignment: Left,
				ColumnAlignment: Left,
				Prefix:          "",
				Suffix:          "",
				Width:           0,
				WidthType:       "auto",
			},
		},
		{
			name:  "Header with right alignment",
			input: "Price|HR",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Price",
				HeaderAlignment: Right,
				ColumnAlignment: Left,
				Prefix:          "",
				Suffix:          "",
				Width:           0,
				WidthType:       "auto",
			},
		},
		{
			name:  "Header with column alignment",
			input: "Amount|CR",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Amount",
				HeaderAlignment: Center,
				ColumnAlignment: Right,
				Prefix:          "",
				Suffix:          "",
				Width:           0,
				WidthType:       "auto",
			},
		},
		{
			name:  "Header with right alignment and right column alignment",
			input: "Price|HR,CR",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Price",
				HeaderAlignment: Right,
				ColumnAlignment: Right,
				Prefix:          "",
				Suffix:          "",
				Width:           0,
				WidthType:       "auto",
			},
		},
		{
			name:  "Header with prefix",
			input: "Price|HR,CR,P:$",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Price",
				HeaderAlignment: Right,
				ColumnAlignment: Right,
				Prefix:          "$",
				Suffix:          "",
				Width:           0,
				WidthType:       "auto",
			},
		},
		{
			name:  "Header with suffix",
			input: "Percentage|HC,CC,S:%",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Percentage",
				HeaderAlignment: Center,
				ColumnAlignment: Center,
				Prefix:          "",
				Suffix:          "%",
				Width:           0,
				WidthType:       "auto",
			},
		},
		{
			name:  "Header with prefix and suffix",
			input: "Balance|HR,CR,P:$,S:USD",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Balance",
				HeaderAlignment: Right,
				ColumnAlignment: Right,
				Prefix:          "$",
				Suffix:          "USD",
				Width:           0,
				WidthType:       "auto",
			},
		},
		{
			name:  "Header with fixed width",
			input: "Name|HL,CL,W:120",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Name",
				HeaderAlignment: Left,
				ColumnAlignment: Left,
				Prefix:          "",
				Suffix:          "",
				Width:           120,
				WidthType:       "fixed",
			},
		},
		{
			name:  "Header with percentage width",
			input: "Name|HL,CL,W:30%",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Name",
				HeaderAlignment: Left,
				ColumnAlignment: Left,
				Prefix:          "",
				Suffix:          "",
				Width:           30,
				WidthType:       "percent",
			},
		},
		{
			name:  "Complete example with all options",
			input: "Product|HL,CR,P:Item:,S:USD,W:40%",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Product",
				HeaderAlignment: Left,
				ColumnAlignment: Right,
				Prefix:          "Item:",
				Suffix:          "USD",
				Width:           40,
				WidthType:       "percent",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseHeaderFormat(tc.input)
			if result.HeaderTitle != tc.expectedResult.HeaderTitle {
				t.Errorf("Expected HeaderTitle to be %s, got %s", tc.expectedResult.HeaderTitle, result.HeaderTitle)
			}
			if result.HeaderAlignment != tc.expectedResult.HeaderAlignment {
				t.Errorf("Expected HeaderAlignment to be %d, got %d", tc.expectedResult.HeaderAlignment, result.HeaderAlignment)
			}
			if result.ColumnAlignment != tc.expectedResult.ColumnAlignment {
				t.Errorf("Expected ColumnAlignment to be %d, got %d", tc.expectedResult.ColumnAlignment, result.ColumnAlignment)
			}
			if result.Prefix != tc.expectedResult.Prefix {
				t.Errorf("Expected Prefix to be %s, got %s", tc.expectedResult.Prefix, result.Prefix)
			}
			if result.Suffix != tc.expectedResult.Suffix {
				t.Errorf("Expected Suffix to be %s, got %s", tc.expectedResult.Suffix, result.Suffix)
			}
			if result.Width != tc.expectedResult.Width {
				t.Errorf("Expected Width to be %f, got %f", tc.expectedResult.Width, result.Width)
			}
			if result.WidthType != tc.expectedResult.WidthType {
				t.Errorf("Expected WidthType to be %s, got %s", tc.expectedResult.WidthType, result.WidthType)
			}
		})
	}
}
