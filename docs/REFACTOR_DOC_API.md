# Refactoring Document: TinyWASM PDF API

## 1. Goal
Design a fluent, intuitive, and easy-to-use API for creating PDF documents that is fully compatible with WebAssembly (WASM) through the `tinywasm` ecosystem. The API must be a significant improvement over previous iterations (like `docpdfOLD`), specifically addressing the unintuitive table creation process.

## 2. API Design Principles
- **Fluent & Chainable:** Methods should return an interface or pointer to allow chaining where it makes sense (e.g., `doc.AddText("...").Bold().AlignRight().Draw()`).
- **Type Safety & IDE Support:** Eliminate "magic strings" and use strong typing (structs or builder methods) so that autocomplete helps the developer avoid errors.
- **WASM Compatibility:** Maximize compatibility with `tinygo`. Avoid `fmt`, `encoding/json`, `reflect` when possible, instead relying on `tinywasm/fmt`, `tinywasm/json`, etc.
- **Dependency Injection / IO Isolation:** Continue using the `env.back.go` / `env.front.go` abstractions for file system access (`readFile`, `writeFile`), ensuring the core API logic remains completely isomorphic and independent.

## 3. Table API: The Fluent Builder Approach (Selected)
For table creation, we will use the Fluent Builder pattern. This approach allows defining the table structure step-by-step programmatically, taking advantage of IDE autocomplete and eliminating string-based configurations.

### Usage Example:
```go
table := doc.AddTable().
    AddColumn("Code").Width(10).AlignCenter().
    AddColumn("Product").Width(40).AlignLeft().
    AddColumn("Quantity").Width(15).AlignRight().Suffix(" units").
    AddColumn("Price").Width(15).AlignRight().Prefix("$").
    AddColumn("Total").Width(20).AlignRight().Prefix("$")

// Global table styles
table.HeaderStyle(pdf.Style{
    FillColor: pdf.ColorRGB(220, 230, 255),
    TextColor: pdf.ColorRGB(20, 20, 100),
    Font:      pdf.FontBold,
    FontSize:  12,
})

// Add data
for _, item := range invoice.Items {
    table.AddRow(item.Code, item.Name, item.Qty, item.Price, item.Total)
}

table.Draw()
```

## 4. Resource Management (Fonts & Images)
Handling external resources (fonts and images) must be predictable across both server (Backend) and browser (WASM Frontend).

### Default Fonts (Core Fonts)
If no font is registered, the PDF library utilizes the built-in "Core Fonts" (`helvetica`, `courier`, `times`, etc.). These allow creating PDFs **100% synchronously** without needing downloads, as the glyphs are embedded in the PDF standard.

```go
// Synchronous Usage (No external resources)
doc := pdf.NewDocument()
doc.AddText("Text with Core Font").SetFont("helvetica", 12).Draw()
doc.WritePdf("doc.pdf") 
```

### External Resources: Async Loading (WASM) and Sync (Backend)
If the user requires custom `.ttf` or `.png` assets, the API will require pre-loading them before generating the document to transparently handle web environment asynchrony (JS `fetch`) and local Go disk reading (`os.ReadFile`).

```go
doc := pdf.NewDocument()

// 1. Declare mandatory dependencies
doc.RegisterFont("Roboto", "fonts/Roboto-Regular.ttf")
doc.RegisterImage("Logo", "img/logo.png")

// 2. Load dependencies in a platform-agnostic way
doc.Load(func(err error) {
    if err != nil { return } // Handle error

    // 3. Create document using memory-cached dependencies
    doc.AddText("Hello").SetFont("Roboto", 12).Draw()
    doc.AddImage("Logo").Width(50).Draw()
    doc.WritePdf("doc.pdf")
})
```

## 5. Base Structural Components
To maintain simplicity and consistency, all structural elements added to the PDF will use a Build -> Modify -> Draw (`.Draw()`) pattern. This `.Draw()` verb tells the engine to insert the block into the current flow.

### 5.1 Text & Paragraphs
```go
// Simple paragraphs
doc.AddText("Normal paragraph flowing with page margins.").Draw()

// Text with stacked styles
doc.AddText("Red text aligned to the right.").
    Bold().
    AlignRight().
    SetColor(255, 0, 0).
    Draw()

// Justified text for long paragraphs
doc.AddText("A very long text that must fill the width.").Justify().Draw()
```

### 5.2 Document Structure (Headers)
Predefined header levels that automatically control size, font (bold), and margins (spacing before/after).
```go
doc.AddHeader1("Main Document Title").Draw()
doc.AddHeader2("Section 1").Draw()
doc.AddHeader3("Subsection 1.1").Draw()
```

### 5.3 Global Page Header/Footer
Configured once and automatically invoked whenever the engine starts a new page.
```go
doc.SetPageHeader().
    SetLeftText("Invoice N° 1234").
    SetRightText("Customer: Acme Inc.")

doc.SetPageFooter().
    SetCenterText("Confidential").
    WithPageTotal(pdf.AlignRight) // Renders "1 / 4" on the right
```

### 5.4 Structural Flow Control
```go
doc.SpaceBefore(5)      // Add 5 units of empty space.
doc.AddPage()           // Force a manual page break.
doc.AddSeparator()      // Draw an optional horizontal divisor line (HR).
```

### 5.5 Charts & Images
Fluent usage that allows referencing the logical ID of a pre-loaded Image (or if it's Core, it's inserted directly).
```go
// Images (The originID "Logo" must have been pre-registered if external)
doc.AddImage("Logo").Height(35).AlignCenter().Draw()
```

> [!NOTE]
> **Chart Integration**
> To keep the WASM binary size minimal, charts will no longer be an abstract independent library; they will be deeply integrated, consuming `fpdf` native vector primitives. The technical design to achieve this without code duplication is detailed in the document:
> **[CHART_ARCHITECTURE.md](./CHART_ARCHITECTURE.md)**

## 6. Next Steps (Implementation)
1. **Project Scaffolding (`tinywasm/pdf`):** 
   - Clean and prepare the base on `fpdf`.
   - Implement the main `Document` scaffolding and its isomorphic initialization (WASM and OS).
2. **Implement Load/Fetch Recursion Flow:** `env.back.go` and `env.front.go` with `tinywasm/fetch`.
3. **Develop Base Components:** Implement the Text interface, Modular Headers, and Flow Control.
4. **Develop Table Component:** Implement the Fluent Builder detailed in **Section 3**.
5. **Generate Dual Test Suite:** Frontend/Backend using the `tinywasm` infrastructure.

## 7. Project Structure & Testing
To maintain internal order (since `fpdf` has its own mixed tests), the new API and charts will have strict organization rules.

### 7.1 Test Folders (`tests/`)
Any integration or e2e test for the new API, or black-box (DDT) tests for generating PDFs, must go in isolated `tests/` directories, _not_ in the library root.

### 7.2 Expected Structure
```text
tinywasm/pdf/
├── env.back.go         # OS Implementation (Sync file reading)
├── env.front.go        # WASM Implementation (Async Fetch)
├── fpdf/               # (Internal fpdf refactor/fork, keeps legacy unit tests)
├── chart/              # Chart organizational logic (delegating drawing to pdf)
│   ├── bar.go
│   ├── line.go
│   ├── pie.go
│   └── tests/          # Specific tests for generating Chart PDFs (DDT)
│       ├── bar_test.go
│       └── ...
├── tests/              # Tests for the new main API (Flow, Tables, Text)
│   ├── api_test.go     # e2e or DDT tests for the new fluent API
│   └── table_test.go   
├── document.go         # New Fluent API Core
├── components.go       # Text, Headers
├── table.go            # Fluent Table Builder
└── ...
```
