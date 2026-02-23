# Implementation: Pie Chart

## Objective
Precise instructions for implementing the **Pie Chart** in `tinywasm/pdf` using directly `fpdf` geometric primitives to maintain WASM binary lightness, thus replacing the old separate abstract library.

## Data Structure
```go
package pdf

type PieChart struct {
    doc    *Document
    title  string
    width  float64
    height float64
    slices []pieSlice
}

type pieSlice struct {
    label string
    value float64
    color Color
}
```

## Rendering Logic (`Draw()`)

### 1. Data Preparation (Angles)
1. Iterate through all `slices` to find the `total_sum`.
2. For each slice, calculate its angle in the circle (proportion): 
   `slice_angle = (slice.value / total_sum) * 360.0`.
   *(Note: in computer graphics, radians are often used instead of degrees)*.

### 2. Rendering Pie Slices
Unlike simple lines or rectangles, drawing a pie slice requires arc primitives or Bézier curves. **Jules must verify the specific capabilities of the `fpdf` branch**. 

**Strategy 1 (Point Polygon - Recommended for simplicity):**
If `fpdf` doesn't have a native `Slice` or `ArcTo` function:
- Define the chart center `(cx, cy)` and a radius `r`.
- For each `slice`, use a loop that divides the `slice_angle` into small steps (e.g., 1 degree).
- Calculate edge points using basic Go trigonometry: 
  - `x = cx + r * math.Cos(angle)`
  - `y = cy + r * math.Sin(angle)`
- Build an array of points that includes `(cx, cy)`, then the perimeter arc points, and close the polygon.
- Call `doc.fpdf.Polygon(points, "F")` after setting the color with `SetFillColor(col.R,col.G,col.B)`.

**Strategy 2 (If `Sector` exists in `fpdf`):**
Certain versions/forks of `fpdf` have the command `Sector(xc, yc, r, a, b, style)`. If this is the case in our `tinywasm/fpdf` version:
- Call `doc.fpdf.Sector(...)` directly.

### 3. Labels / Legends
Jules must implement label placement in two ways (chosen by the Builder):
**Mode A (Radial):** 
Print text near the center of the arc (`middle_angle`):
`tx = cx + (r * 0.7) * math.Cos(middle_angle)`
`ty = cy + (r * 0.7) * math.Sin(middle_angle)`
`doc.fpdf.Text(tx, ty, label)`

**Mode B (Side Legend):**
Draw color squares and text to the right of the chart.

## Notes for Jules
- For all structural constraints (zero buffers, no external math), adhere strictly to Section 5 of `CHART_ARCHITECTURE.md`.

## Reference Code (From legacy `docpdfOLD/chart`)

Jules, review how the coordinates (`cx`, `cy`), radius, and angles (`mathutils.PercentToRadians`) were computed. Adapt this to **pure `fpdf` paths**. Since `fpdf` might not have an `ArcTo` method, remember to use Point Polygons calculation via sine/cosine if `Sector` does not exist in our internal package.

```go
func (pc PieChart) drawSlices(r Renderer, canvasBox canvas.Box, values []Value) {
	cx, cy := canvasBox.Center()
	diameter := mathutils.MinInt(canvasBox.Width(), canvasBox.Height())
	radius := float64(diameter >> 1)
	labelRadius := (radius * 2.0) / 3.0

	var rads, delta, delta2, total float64
	var lx, ly int

	for index, v := range values {
		v.Style.InheritFrom(pc.stylePieChartValue(index)).WriteToRenderer(r)

		r.MoveTo(cx, cy)
		rads = mathutils.PercentToRadians(total) // Scale custom percentage to radians
		delta = mathutils.PercentToRadians(v.Value)

		r.ArcTo(cx, cy, radius, radius, rads, delta) // Note: Replace this with Polygon logic if fpdf lacks ArcTo

		r.LineTo(cx, cy)
		r.Close()
		r.FillStroke()
		total = total + v.Value
	}

	// draw the labels
	total = 0
	for index, v := range values {
		if len(v.Label) > 0 {
			delta2 = mathutils.PercentToRadians(total + (v.Value / 2.0))
			delta2 = mathutils.RadianAdd(delta2, mathutils.TwoPi)
			lx, ly = mathutils.CirclePoint(cx, cy, labelRadius, delta2)

			tb := r.MeasureText(v.Label)
			lx = lx - (tb.Width() >> 1)
			ly = ly + (tb.Height() >> 1)

			r.Text(v.Label, lx, ly)
		}
		total = total + v.Value
	}
}
```
