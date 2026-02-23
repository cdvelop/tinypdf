# Master Prompt & Execution Plan: TinyWASM PDF Refactor

## 🎯 Global Context (For Jules)
**Project:** `github.com/tinywasm/pdf`
**Objective:** Build a fluent, WASM-compatible, and highly optimized PDF generation API on top of an internal `fpdf` fork.
**Core Constraints:**
- **Zero bloat:** The WASM binary size is critical. Do not import heavy external libraries.
- **TinyGo Compatibility:** Avoid `fmt`, `reflect`, `encoding/json`, etc. Use `github.com/tinywasm/fmt` (`tinystring`) instead, as detailed in `TINYGO_PORT.md`.
- **DDT Testing:** All tests for the new API must be Diagram-Driven Testing / Black-box and placed in `tests/` or `chart/tests/`. Compare generated PDFs visually/by file presence, NOT by exact string matching.

---

## ⚠️ Known Inconsistencies & Architectural Directives
Before starting the steps, Jules must be aware of the following architectural resolutions to avoid getting stuck:

1. **Inconsistency: Asynchronous Loading vs `fpdf` Initialization**
   - *Problem:* `fpdf` traditionally loads `.ttf` fonts synchronously from disk using file paths. Our new API uses an isomorphic `Load(callback)` method that fetches bytes asynchronously in WASM.
   - *Resolution:* Before implementing the fluent API, Jules **MUST** ensure `fpdf` supports loading fonts from `[]byte`. This requires implementing `TtfParseBytes` within the `fpdf` internal package, as detailed in `TTF_PARSE_BYTES.md`. 

2. **Inconsistency: `fpdf` Refactoring vs New API Creation**
   - *Problem:* There are pending refactors for `fpdf` (e.g., `TINYGO_PORT.md` to remove `fmt`). Mixing this massive internal refactor with the creation of the new Fluent API (`document.go`) will exceed context limits and cause errors.
   - *Resolution:* Internal `fpdf` updates must be treated as a strict **prerequisite step (Step 1)** before writing the new high-level wrapper.

3. **Inconsistency: Chart Drawing Primitives (Pie Chart)**
   - *Problem:* `CHART_PIE_IMPL.md` asks to draw pie slices, but native `fpdf` might not have a `Sector` function exposed, or doing arcs manually might be complex.
   - *Resolution:* Jules must inspect `fpdf`'s drawing primitives. If `Sector` or `ArcTo` is missing, use the **Point Polygon** strategy (calculating perimeter points with `math.Cos`/`math.Sin` and drawing a `Polygon`). 

4. **Inconsistency: Subpackage vs Root Package for Charts**
   - *Problem:* If charts are in a `chart` subfolder but use the main `Document` pointer from the parent `pdf` package, it creates circular dependencies.
   - *Resolution:* Keep the chart implementation files (`bar_chart.go`, `line_chart.go`, `pie_chart.go`) in the **root `pdf` package** alongside `document.go`, or ensure `chart` only receives the internal `*fpdf.Fpdf` instance to avoid depending on the `Document` wrapper. For simplicity, put them in the root package but group their tests in a dedicated `tests/chart/` folder to maintain order.

---

## 📋 Execution Order (Step-by-Step)
Jules MUST execute these steps sequentially. Do not move to the next step until the current one compiles and its tests pass.

### **Step 1: Internal `fpdf` Preparation & TinyGo Port**
- Read `TINYGO_PORT.md` and `TTF_PARSE_BYTES.md`.
- Refactor internal `fpdf` package to remove standard library dependencies (`fmt`, `strconv`) and replace them with `github.com/tinywasm/fmt`.
- Implement `TtfParseBytes([]byte)` inside `fpdf` to allow font injection from memory.
- *Check:* Run `gotest` on `fpdf/` to ensure the core still works.

### **Step 2: Isomorphic IO & Resource Loading**
- Create `env.back.go` (using `os.ReadFile` and `os.WriteFile`).
- Create `env.front.go` (using JS `fetch` via `github.com/tinywasm/fetch` and `localStorage` / Browser Download triggers).
- Both should fulfill the internal `readFile` and `writeFile` interfaces needed by the API.

### **Step 3: Base Fluent API (`Document`)**
- Read `REFACTOR_DOC_API.md` (Sections 1, 2, 4, and 5).
- Create `document.go` containing the `Document` struct (wrapping `fpdf.Fpdf`).
- Implement `RegisterFont`, `RegisterImage`, and the `Load(callback)` pattern.
- Implement base components: `AddText`, `AddHeader1`, `SetPageHeader`, `SpaceBefore`, `AddPage`.
- *Check:* Create `tests/api_test.go` that generates a basic PDF without tables or charts.

### **Step 4: Fluent Table Builder**
- Read `REFACTOR_DOC_API.md` (Section 3).
- Create `table.go` with the fluent builder pattern (`AddColumn`, `AddRow`, `Draw`).
- *Check:* Create `tests/table_test.go` that generates a PDF containing a formatted table.

### **Step 5: Chart Architecture Foundation**
- Read `CHART_ARCHITECTURE.md`.
- Implement the base `Chart()` factory on the `Document` struct.
- Define how `pdf.Style` and `pdf.Color` will be passed down to the chart generators.

### **Step 6: Individual Chart Implementations**
- Read `CHART_BAR_IMPL.md` -> Create `bar_chart.go`.
- Read `CHART_LINE_IMPL.md` -> Create `line_chart.go`.
- Read `CHART_PIE_IMPL.md` -> Create `pie_chart.go`.
- *Check:* Create `tests/chart_test.go` that generates a comprehensive PDF containing all three charts. Visually verify the output.

---
**End of Prompt.** 
*(Jules: Start by confirming your understanding of Step 1 and proceed to execute it.)*
