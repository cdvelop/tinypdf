# Architecture Document: TinyWASM PDF Charts Integration

## 1. Context and Problem
In previous iterations (e.g., `docpdfOLD`), the charting capabilities were built as a completely separate library (`docpdfOLD/chart`). This approach, while modular, introduced significant disadvantages for WebAssembly (WASM) environments:

1. **Code Duplication:** Both `fpdf` and `chart` defined their own abstractions for `Color`, `Style`, `Point`, `Font`, and rendering structs.
2. **Binary Bloat:** Duplicating drawing logic, color palettes, and mathematical utilities drastically increases the final `.wasm` binary size, which goes against the `tinywasm` philosophy.
3. **Double Rendering Penalty:** The chart library effectively renders to an intermediate `canvas` abstraction, which is then translated *again* into PDF instructions by `fpdf`. 

## 2. Objective
Design a new `Chart` architecture that is **deeply integrated directly into the `fpdf` core rendering pipeline**. By leveraging `fpdf`'s native drawing primitives (lines, rectangles, bezier curves, and SVG paths) and sharing the exact same `Style` definitions, we can eliminate the intermediate abstraction layer and drastically reduce the WASM binary size.

## 3. Architecture Proposal: Deep fpdf Integration

### 3.1 Unification of Core Types
The charting module will **not** define its own colors, fonts, or base styles. It will strictly consume `pdf.Style`, `pdf.Color`, and `pdf.Font`.

*Removed from Chart:*
- `chart.ColorType` / `chart.Style` (Replaced by `pdf.Style`)
- `chart.Renderer` interface (Replaced by direct calls to `*fpdf.Fpdf` methods).

### 3.2 Drawing Directly to PDF Primitives
Instead of the `chart` library using an abstract `canvas` that `fpdf` later interprets, the Chart component will use `fpdf`'s drawing methods directly.

*Example - Drawing a Bar:*
```go
// Instead of: canvas.DrawBox() 
// We will do directly:
doc.SetFillColor(...)
doc.Rect(x, y, w, h, "F") // Native PDF rendering
```

### 3.3 Leveraging Native SVG Paths
`fpdf` already contains a robust, lightweight engine for parsing and drawing SVG paths (Lines, curves, moves). Complex charts (like Line charts with smoothing or Pie charts) should generate simple SVG path strings (`"M x y L x y"`) and feed them directly to `fpdf`'s path renderer. This avoids duplicating Bezier curve math in the `chart` package.

## 4. Proposed Chart API (Fluent & Integrated)
The Chart API will remain visually fluent (as established in `REFACTOR_DOC_API.md`), but under the hood, it's just a macro over `fpdf` drawing calls.

```go
// 1. Start a chart anchored to the current document
barChart := doc.Chart().Bar().
    Title("Sales Projection").
    Height(300).
    Width(400) // If not provided, use available width

// 2. Define Data
barChart.AddSerie("Sales", pdf.Style{FillColor: pdf.ColorRGB(50, 100, 200)}).
    AddData("Jan", 120).
    AddData("Feb", 140)

// 3. Draw Directly (Calls doc.Rect(), doc.Text(), doc.Line() internally)
barChart.Draw() 
```

## 5. Development Rules & Constraints
1. **Zero Intermediate Buffers:** Do not render the chart to an image byte slice (PNG/JPG) in memory. Charts must be drawn as vector instructions (Lines, Polygons, Text) directly into the PDF byte stream. This keeps the file size tiny and quality infinite (vectorial).
2. **No External Math/Dependency:** Do not import heavy external math or rendering libraries. Calculate coordinates using standard Go math and draw using `fpdf`.
3. **Lazy Evaluation:** The `Draw()` method is responsible for all calculations (scaling, axis mapping). Before `Draw()`, the chart simply collects data points and configs.

## 6. Implementation Guides (For Builder Agent Jules)
To isolate complexity and strictly guide the builder LLM (*Jules*) to avoid code duplication and comply with WASM binary size constraints, the development of this module is divided into three specific flow documents per chart type.

The agent responsible for implementing must follow these guides to the letter, using `fpdf` native primitives (SVG paths, polygons, lines, text boxes) strictly forbidding intermediate *bitmap* image rendering.

**Specific Implementation Documents:**
- 👉 [Bar Chart Implementation](./CHART_BAR_IMPL.md)
- 👉 [Line Chart Implementation](./CHART_LINE_IMPL.md)
- 👉 [Pie Chart Implementation](./CHART_PIE_IMPL.md)

## 7. Testing Strategy (Instruction for Jules)
Testing the exact output of vector graphics (SVG embedded in PDF) by comparing raw _bytes_ or _strings_ is **extremely fragile** and impractical for an LLM, as minimal floating-point variations in math would break the tests.

**Testing Instruction for Jules:**
1. **Avoid Exact String Tests:** Do not attempt to create a unit test that compares the final PDF document string or drawing command with an exact "expected string".
2. **Black-Box Testing (DDT):** Following the TinyWASM Diagram-Driven Testing or Black-Box standard, the main focus should be integration testing. Jules must create a test in `tests/` that generates a real PDF document containing the 3 types of charts drawn (`TestCharts_GeneratePDF()`).
3. **Visual/Manual Validation:** Test success is validated by checking in code that the `charts_tests_output.pdf` file was generated without errors in the file system (e.g., `os.WriteFile`), is not corrupt (size > 0), and in the real workflow, the human developer will open that PDF to visually confirm that the mathematical calculations (scales, positions, axes) render correctly.
