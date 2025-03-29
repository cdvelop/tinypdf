package tinypdf

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
