package fontManager

type fontStyle string

const (
	Regular fontStyle = "Regular"
	Bold    fontStyle = "Bold"
	Italic  fontStyle = "Italic"
	// Additional styles
	BoldItalic fontStyle = "BoldItalic"
	Light      fontStyle = "Light"
	SemiBold   fontStyle = "SemiBold"
	ExtraBold  fontStyle = "ExtraBold"
)
