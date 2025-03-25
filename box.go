package tinypdf

// Box represents a rectangular area with explicit coordinates for all four sides.
// It is used for defining boundaries in PDF documents, such as margins, trim boxes, etc.
// The coordinates are stored in the current unit system (points by default, but can be mm, cm, inches, or pixels).
type Box struct {
	Left, Top, Right, Bottom float64
	unitOverride             defaultUnitConfig
}

// UnitsToPoints converts the box coordinates to Points.
// When this is called it is assumed the values of the box are in the specified unit system.
// The method creates a new Box instance with coordinates converted to points.
//
// Parameters:
//   - t: An integer representing the unit type to convert from (UnitPT, UnitMM, UnitCM, UnitIN, UnitPX)
//
// Returns:
//   - A new Box pointer with coordinates converted to points
func (box *Box) UnitsToPoints(t int) (b *Box) {
	if box == nil {
		return
	}

	unitCfg := defaultUnitConfig{Unit: t}
	if box.unitOverride.getUnit() != UnitUnset {
		unitCfg = box.unitOverride
	}

	b = &Box{
		Left:   box.Left,
		Top:    box.Top,
		Right:  box.Right,
		Bottom: box.Bottom,
	}
	unitsToPointsVar(unitCfg, &b.Left, &b.Top, &b.Right, &b.Bottom)
	return
}

// unitsToPoints is an internal method that converts the box coordinates to Points
// using the provided unit configuration.
// It creates a new Box instance with coordinates converted to points.
//
// Parameters:
//   - unitCfg: A unitConfigurator interface that provides unit conversion configuration
//
// Returns:
//   - A new Box pointer with coordinates converted to points
func (box *Box) unitsToPoints(unitCfg unitConfigurator) (b *Box) {
	if box == nil {
		return
	}

	if box.unitOverride.getUnit() != UnitUnset {
		unitCfg = box.unitOverride
	}

	b = &Box{
		Left:   box.Left,
		Top:    box.Top,
		Right:  box.Right,
		Bottom: box.Bottom,
	}
	unitsToPointsVar(unitCfg, &b.Left, &b.Top, &b.Right, &b.Bottom)
	return
}
