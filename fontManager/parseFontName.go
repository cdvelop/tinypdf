package fontManager

import (
	. "github.com/cdvelop/tinystring"
)

// parseFontName extracts the font family and style from a filename or path.
// It preserves the family name casing as provided in the basename and detects
// common style suffixes (case-insensitive) such as "bold" and "italic".
func parseFontName(fontPath string) (family string, style fontStyle) {
	// Extract base filename (handles paths) and strip extension case-insensitively
	base := PathBase(fontPath)
	if HasSuffix(Convert(base).ToLower().String(), ".ttf") {
		base = base[:len(base)-4]
	}

	// Split on hyphen; last part may be the style
	parts := Convert(base).Split("-")
	if len(parts) > 1 {
		rawStyle := parts[len(parts)-1]
		family = Convert(parts[:len(parts)-1]).Join("-").String()

		// Normalize style (case-insensitive)
		s := Convert(rawStyle).ToLower().String()
		switch s {
		case "bold", "b":
			style = Bold
		case "italic", "it", "i", "oblique", "ob":
			style = Italic
		case "bolditalic", "bold-italic", "italicbold", "bi":
			style = BoldItalic
		case "light", "thin", "extralight", "ultralight":
			style = Light
		case "semibold", "demibold", "demi-bold", "sb":
			style = SemiBold
		case "extrabold", "heavy", "black", "eb":
			style = ExtraBold
		default:
			// Unknown style token -> treat as Regular and include the token in family
			// e.g. "SomeFont-Extra" -> family: "SomeFont-Extra" (no style)
			family = base
			style = Regular
		}
	} else {
		family = base
		style = Regular
	}
	return family, style
}
