# Implementation: Line Chart

## Objective
Precise instructions for implementing the **Line Chart** component (`LineChart`), focused on using the `fpdf` **Path/SVG** system to achieve continuous (and optionally smoothed) paths without inflating the WASM binary.

## Data Structure
```go
package pdf

type LineChart struct {
    doc    *Document
    title  string
    width  float64
    height float64
    series []lineSeries
}

type lineSeries struct {
    name  string
    data  []float64 // We assume equidistant X distribution for simplicity
    color Color
    width float64
}
```

## Rendering Logic (`Draw()`)
To draw the lines, we will not iterate blindly calling `doc.fpdf.Line()`, as that doesn't allow filling areas under the curve if desired in the future. We will use Path instructions (native PDF drawing paths).

### Path Construction Steps:
1. **Calculate X and Y Scales.**
2. **Draw Axes** (Same as in Bar Chart).
3. **Build the Path for each Series:**
   PDF works like SVG.
   - Initial position: `MoveTo(x0, y0)`
   - Draw line: `LineTo(x1, y1)`, `LineTo(x2, y2)`
   
### How to program the Path in fpdf:
In pure `fpdf`, this is done internally by building rendering commands (`fmtBuffer` or direct methods if the `gofpdf`/`fpdf` version has path commands).
If the `fpdf` wrapper doesn't explicitly expose `MoveTo`/`LineTo`, but you know `fpdf` interprets HTML or SVG, you can inject the drawing. However, the standard supported way in `fpdf` for polygons is `fpdf.Polygon()` or individual lines:

*Recommended Strategy for Jules:*
- Iterate through `data[i]` and `data[i+1]` points.
- Use `doc.fpdf.SetDrawColor(col.R, col.G, col.B)`.
- Use `doc.fpdf.SetLineWidth(series.width)`.
- Call `doc.fpdf.Line(x1, y1, x2, y2)` to connect points sequentially.
- Optional: Draw circles at vertices using `doc.fpdf.Circle(x, y, radius, "F")`.

## Notes for Jules
- **Bézier Curves (Smoothing):** Only implement if indispensable using `doc.fpdf.Curve(x0,y0, cx1,cy1, cx2,cy2, x1,y1, "D")`. If smoothing math is too heavy, keep it as straight lines.
- For all other structural constraints (zero buffers, no external dependencies), adhere strictly to Section 5 of `CHART_ARCHITECTURE.md`.

## Reference Code (From legacy `docpdfOLD/chart`)

Jules, you can use the mathematical looping from this legacy code, but **you must translate the custom `Renderer` calls (`r.MoveTo`, `r.LineTo`) directly into `fpdf` method calls**.

```go
func (d draw) LineSeries(r Renderer, canvasBox canvas.Box, xrange, yrange Range, style Style, vs ValuesProvider) {
	if vs.Len() == 0 { return }

	cb := canvasBox.Bottom
	cl := canvasBox.Left

	v0x, v0y := vs.GetValues(0)
	x0 := cl + xrange.Translate(v0x)
	y0 := cb - yrange.Translate(v0y)

	var vx, vy float64
	var x, y int

	if style.ShouldDrawStroke() {
		style.GetStrokeOptions().WriteDrawingOptionsToRenderer(r)
		r.MoveTo(x0, y0)
		for i := 1; i < vs.Len(); i++ {
			vx, vy = vs.GetValues(i)
			x = cl + xrange.Translate(vx)
			y = cb - yrange.Translate(vy)
			r.LineTo(x, y)
		}
		r.Stroke()
	}
}
```
