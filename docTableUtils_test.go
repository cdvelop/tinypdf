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
			},
		},
		{
			name:  "Header with right alignment and right column alignment",
			input: "Price|HRR",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Price",
				HeaderAlignment: Right,
				ColumnAlignment: Right,
				Prefix:          "",
				Suffix:          "",
			},
		},
		{
			name:  "Header with center alignment and column right alignment",
			input: "Amount|HR",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Amount",
				HeaderAlignment: Right,
				ColumnAlignment: Left,
				Prefix:          "",
				Suffix:          "",
			},
		},
		{
			name:  "Header with prefix",
			input: "Price|HRRP:$",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Price",
				HeaderAlignment: Right,
				ColumnAlignment: Right,
				Prefix:          "$",
				Suffix:          "",
			},
		},
		{
			name:  "Header with suffix",
			input: "Percentage|HCS:%",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Percentage",
				HeaderAlignment: Center,
				ColumnAlignment: Center,
				Prefix:          "",
				Suffix:          "%",
			},
		},
		{
			name:  "Header with both prefix and suffix",
			input: "Balance|HRRP:$S:USD",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Balance",
				HeaderAlignment: Right,
				ColumnAlignment: Right,
				Prefix:          "$",
				Suffix:          "USD",
			},
		},
		{
			name:  "Header with column center alignment",
			input: "Title|HLC",
			expectedResult: headerFormatOptions{
				HeaderTitle:     "Title",
				HeaderAlignment: Left,
				ColumnAlignment: Center,
				Prefix:          "",
				Suffix:          "",
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
		})
	}
}

func TestParseHeaderAlignment(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedResult int
	}{
		{"Header left", "HL", Left},
		{"Header right", "HR", Right},
		{"Header center", "H", Center},
		{"Default when not specified", "CR", 0},
		{"Multiple options including header left", "HLRP:$", Left},
		{"Multiple options including header right", "HRRP:$", Right},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseHeaderAlignment(tc.input)
			if result != tc.expectedResult {
				t.Errorf("Expected %d, got %d", tc.expectedResult, result)
			}
		})
	}
}

func TestParseColumnAlignment(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedResult int
	}{
		{"Column left", "L", Left},
		{"Column center", "C", Center},
		{"Column right", "R", Right},
		{"Default when not specified", "H", 0},
		{"Multiple options including column center", "HLC", Center},
		{"Multiple options including column right", "HRR", Right},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseColumnAlignment(tc.input)
			if result != tc.expectedResult {
				t.Errorf("Expected %d, got %d", tc.expectedResult, result)
			}
		})
	}
}

func TestParsePrefix(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedResult string
	}{
		{"Simple prefix", "P:$", "$"},
		{"Prefix in multiple options", "HRRP:$", "$"},
		{"Prefix followed by suffix", "P:$S:%", "$"},
		{"No prefix", "HRR", ""},
		{"Empty prefix", "P:", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseTableHeaderPrefix(tc.input)
			if result != tc.expectedResult {
				t.Errorf("Expected '%s', got '%s'", tc.expectedResult, result)
			}
		})
	}
}

func TestParseSuffix(t *testing.T) {
	testCases := []struct {
		name           string
		input          string
		expectedResult string
	}{
		{"Simple suffix", "S:%", "%"},
		{"Suffix in multiple options", "HRRS:%", "%"},
		{"Suffix after prefix", "P:$S:%", "%"},
		{"No suffix", "HRR", ""},
		{"Empty suffix", "S:", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseTableHeaderSuffix(tc.input)
			if result != tc.expectedResult {
				t.Errorf("Expected '%s', got '%s'", tc.expectedResult, result)
			}
		})
	}
}
