package tinypdf

func (c *Current) setTextColor(color ICacheColorText) {
	c.txtColor = color
}

func (c *Current) textColor() ICacheColorText {
	return c.txtColor
}

// ImageCache is metadata for caching images.
type ImageCache struct {
	Path  string //ID or Path
	Index int
	Rect  *Rect
}
