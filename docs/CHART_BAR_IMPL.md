# Implementation: Bar Chart

## Objective
Precise instructions for implementing the **Bar Chart** component (`BarChart`) integrated directly over `fpdf` primitives, ensuring minimal binary size for WASM.

## Data Structure (No external dependencies)
The component must live within `tinywasm/pdf` and exclusively use the base types from this library (`pdf.Style`, `pdf.Color`).

```go
package pdf

type BarChart struct {
    doc    *Document // Reference to parent fpdf document
    title  string
    width  float64
    height float64
    bars   []barData
    style  Style     // General style (font, text color)
}

type barData struct {
    label string
    value float64
    color Color // Allows individual colors per bar
}
```

## Fluent Interface (Builder)
```go
func (c *BarChart) Title(t string) *BarChart { c.title = t; return c }
func (c *BarChart) Height(h float64) *BarChart { c.height = h; return c }
func (c *BarChart) Width(w float64) *BarChart { c.width = w; return c }
func (c *BarChart) AddBar(val float64, label string, col ...Color) *BarChart { ... }
```

## Rendering Logic (`Draw()`)
The `.Draw()` method will execute the scaling math and call `fpdf` primitives.
**Attention Jules:** Do not create temporary images (PNG/JPG)! Draw directly.

### Steps inside `Draw()`:
1. **Calculate Scale (Y):** 
   - Find the `max_value` within `bars`.
   - Calculate scale factor: `scaleY = (chart.height - margins) / max_value`.
2. **Calculate Width (X):**
   - Bar width = `(chart.width - margins) / len(bars)`.
3. **Draw Axes (fpdf Lines):**
   - Use `doc.fpdf.Line(x1, y1, x2, y2)` for the X and Y axis lines.
4. **Draw Bars (fpdf Rect):**
   - Iterate `bars`. For each bar:
     - Rendered height = `bar.value * scaleY`.
     - `doc.fpdf.SetFillColor(col.R, col.G, col.B)`.
     - `doc.fpdf.Rect(x, y - height, width, height, "F")` (Where "F" is Fill).
5. **Draw Texts (fpdf Text):**
   - `doc.fpdf.Text(x, y, bar.label)`.
   - `doc.fpdf.Text(x, y - height - 2, fmt.Sprintf("%.2f", bar.value))`.

## Notes for Jules
- Calculate X/Y coordinates considering that in `fpdf`, the origin (0,0) is the top-left corner of the page, so the Y-axis increases downwards.
- For all other structural constraints (zero buffers, no external dependencies), adhere strictly to Section 5 of `CHART_ARCHITECTURE.md`.

## Reference Code (From legacy `docpdfOLD/chart`)

Jules, you can use the mathematical logic from this legacy code to calculate widths and heights, but **you must translate `Draw.Box` or `r.FillStroke` directly to `fpdf` calls** (`doc.SetFillColor()`, `doc.Rect()`).

```go
func (bc BarChart) drawBars(r Renderer, canvasBox canvas.Box, yr Range) {
	xoffset := canvasBox.Left

	width, spacing, _ := bc.calculateScaledTotalWidth(canvasBox)
	bs2 := spacing >> 1

	var barBox canvas.Box
	var bxl, bxr, by int
	for index, bar := range bc.Bars {
		bxl = xoffset + bs2
		bxr = bxl + width

		by = canvasBox.Bottom - yr.Translate(bar.Value)

		if bc.UseBaseValue {
			barBox = canvas.Box{
				Top:    by,
				Left:   bxl,
				Right:  bxr,
				Bottom: canvasBox.Bottom - yr.Translate(bc.BaseValue),
			}
		} else {
			barBox = canvas.Box{
				Top:    by,
				Left:   bxl,
				Right:  bxr,
				Bottom: canvasBox.Bottom,
			}
		}

		Draw.Box(r, barBox, bar.Style.InheritFrom(bc.styleDefaultsBar(index)))

		xoffset += width + spacing
	}
}
```
