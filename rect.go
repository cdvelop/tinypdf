package tinypdf

// Rect defines a rectangle by its width and height.
// This is used for defining page sizes, content areas, and other rectangular regions in PDF documents.
// The dimensions are stored in the current unit system (points by default, but can be mm, cm, inches, or pixels).
type Rect struct {
	W            float64 // Width of the rectangle
	H            float64 // Height of the rectangle
	unitOverride defaultUnitConfig
}

// PointsToUnits converts the rectangle's width and height from points to the specified unit system.
// When this is called it is assumed the values of the rectangle are in points.
// The method creates a new Rect instance with dimensions converted to the specified units.
//
// Parameters:
//   - t: An integer representing the unit type to convert to (UnitPT, UnitMM, UnitCM, UnitIN, UnitPX)
//
// Returns:
//   - A new Rect pointer with dimensions converted to the specified units
func (rect *Rect) PointsToUnits(t int) (r *Rect) {
	if rect == nil {
		return
	}

	unitCfg := defaultUnitConfig{Unit: t}
	if rect.unitOverride.getUnit() != UnitUnset {
		unitCfg = rect.unitOverride
	}

	r = &Rect{W: rect.W, H: rect.H}
	pointsToUnitsVar(unitCfg, &r.W, &r.H)
	return
}

// UnitsToPoints converts the rectangle's width and height from the specified unit system to points.
// When this is called it is assumed the values of the rectangle are in the specified units.
// The method creates a new Rect instance with dimensions converted to points.
//
// Parameters:
//   - t: An integer representing the unit type to convert from (UnitPT, UnitMM, UnitCM, UnitIN, UnitPX)
//
// Returns:
//   - A new Rect pointer with dimensions converted to points
func (rect *Rect) UnitsToPoints(t int) (r *Rect) {
	if rect == nil {
		return
	}

	unitCfg := defaultUnitConfig{Unit: t}
	if rect.unitOverride.getUnit() != UnitUnset {
		unitCfg = rect.unitOverride
	}

	r = &Rect{W: rect.W, H: rect.H}
	unitsToPointsVar(unitCfg, &r.W, &r.H)
	return
}

// unitsToPoints is an internal method that converts the rectangle's dimensions to points
// using the provided unit configuration.
// It creates a new Rect instance with dimensions converted to points.
//
// Parameters:
//   - unitCfg: A unitConfigurator interface that provides unit conversion configuration
//
// Returns:
//   - A new Rect pointer with dimensions converted to points
func (rect *Rect) unitsToPoints(unitCfg unitConfigurator) (r *Rect) {
	if rect == nil {
		return
	}
	if rect.unitOverride.getUnit() != UnitUnset {
		unitCfg = rect.unitOverride
	}
	r = &Rect{W: rect.W, H: rect.H}
	unitsToPointsVar(unitCfg, &r.W, &r.H)
	return
}
