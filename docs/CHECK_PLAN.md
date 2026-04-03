# PLAN: Reduce WASM binary size of tinywasm/pdf

## Current Status: 738 KB (compiled with TinyGo)

## Goal: ~150 KB (-80%)

## WASM Functional Scope
- Generate PDF with text, styles, tables
- Charts (bar, line, pie) rendered as PDF paths
- Simple images (PNG, JPEG)
- **NO**: modify existing PDFs, add sheets, PDF protection, GIF

---

## Size breakdown by dependency (twiggy analysis)

| Dependency          | Bytes  | KB    | %     | Necessary for WASM? |
|---------------------|--------|-------|-------|-----------------|
| encoding/json       | 93,474 | 91 KB | 12.4% | NO - only parses fonts |
| time (stdlib)       | 77,045 | 75 KB | 10.2% | NO - only date metadata |
| web/ui.setupUI      | 71,945 | 70 KB | 9.5%  | YES - it's the demo, not the lib |
| runtime             | 62,674 | 61 KB | 8.3%  | YES - irreducible |
| image/* total       | 58,071 | 57 KB | 7.7%  | PARTIAL |
| ├─ image/jpeg       | 26,838 | 26 KB |       | YES if used |
| ├─ image/png        | 17,424 | 17 KB |       | YES if used |
| └─ image/gif        | 11,914 | 12 KB |       | NO |
| fpdf core           | 53,482 | 52 KB | 7.1%  | YES |
| fmt (stdlib)        | 46,204 | 45 KB | 6.1%  | NO - indirectly dragged in |
| compress/*          | 38,711 | 38 KB | 5.1%  | NO in WASM (no compression) |
| crypto/*            | 30,719 | 30 KB | 4.1%  | NO - only PDF protection + sha1 |
| tinywasm/fmt        | 22,948 | 22 KB | 3.0%  | YES |
| rodata segments     | ~80,000| 78 KB | 10.6% | PARTIAL - reduces with code |
| pdf pkg             | 12,866 | 13 KB | 1.7%  | YES |
| others              | ~7,000 | 7 KB  | 0.9%  | various |

### Estimated reduction potential

| Action                                | Estimated Savings |
|----------------------------------------|-----------------|
| Remove encoding/json → tinywasm/json | ~90 KB          |
| Replace time → tinywasm/time        | ~75 KB          |
| Remove fmt stdlib (indirect)        | ~45 KB          |
| Remove crypto/* (protection + sha1)  | ~30 KB          |
| Remove image/gif                     | ~12 KB          |
| Deactivate compress/* in WASM          | ~38 KB          |
| Associated rodata reduction              | ~30 KB          |
| **Estimated Total**                     | **~320 KB (43%)**|

---

## STAGES

### Stage 1: Remove crypto/* (PDF protection) — Savings ~30 KB
**Risk**: low | **Complexity**: low

**Context**: `crypto/md5`, `crypto/rc4` are used only in `fpdf/protect.go` for PDF encryption.
`crypto/sha1` is used in `fpdf/def.go:667-671` for `generateFontID()`.

**What to do**:
1. Add `//go:build !wasm` to `fpdf/protect.go`
2. Create `fpdf/protect_wasm.go` (`//go:build wasm`) with stubs for all public functions in `protect.go` that return no-op or error.
3. Refactor `generateFontID()` to remove `crypto/sha1`:
   - Currently: `json.Marshal(&fdt)` → `sha1.Sum(b)` → `fmt.Sprintf("%x", hash)`
   - New: use `tinywasm/unixid` to generate a unique ID.
   - `generateFontID` is called in `fonts.go:109` and `fonts.go:952` when loading fonts.
   - The ID only needs to be unique and deterministic within the session.

**Current `generateFontID` detail** (`fpdf/def.go:667-672`):
```go
func generateFontID(fdt fontDefType) (string, error) {
    fdt.File = ""
    b, err := json.Marshal(&fdt)
    return fmt.Sprintf("%x", sha1.Sum(b)), err
}
```
**New** (`fpdf/def.go`, wasm build tag):
```go
func generateFontID(fdt fontDefType) (string, error) {
    // Deterministic ID based on font name and type
    return fdt.Tp + "_" + fdt.Name, nil
}
```
**Note**: on the backend (`!wasm`), the original SHA1 is maintained. In WASM, the ID only needs to be unique as an internal map key — `Tp_Name` is sufficient and deterministic.

**Files to modify**:
- `fpdf/protect.go` → add `//go:build !wasm`
- `fpdf/protect_wasm.go` → new, empty stubs
- `fpdf/def.go` → split `generateFontID` by build tag: `fpdf/fontid_back.go` (!wasm) and `fpdf/fontid_wasm.go` (wasm)

**Signatures to stub in protect_wasm.go** (check protect.go for complete list):
- Copy all functions `func (f *Fpdf) SetProtection(...)` etc. with an empty body or `f.err = errors.New("protection not supported in WASM")`

**Validation**:
1. `wasmbuild` compiles without errors.
2. `twiggy top ... | grep crypto` → 0 results.
3. `go test ./...` backend tests still pass.
4. Generating PDF in WASM works (without protection).

---

### Stage 2: Remove image/gif — Savings ~12 KB
**Risk**: low | **Complexity**: low

**Context**: `image/gif` is imported in `fpdf/fpdf.go` and drags in the full decoder + `compress/lzw`.

**What to do**:
1. Search for the `"image/gif"` import in `fpdf/fpdf.go` and the code that uses it.
2. Move that code to `fpdf/gif_back.go` with `//go:build !wasm`.
3. Create `fpdf/gif_wasm.go` with a stub that returns a "GIF not supported in WASM" error.

**Search pattern**: search for `gif.` in `fpdf/fpdf.go` to find the exact call sites.
The GIF code is likely in an image registration/decoding function.

**Files**:
- `fpdf/fpdf.go` → extract GIF code to a separate file.
- `fpdf/gif_back.go` → (!wasm) original GIF code.
- `fpdf/gif_wasm.go` → (wasm) stub

**Validation**:
1. `wasmbuild` compiles.
2. `twiggy top ... | grep "image/gif"` → 0 results.
3. Test: registering PNG and JPEG images works in WASM.
4. Backend: GIF still works.

---

### Stage 3: Deactivate compress/* in WASM — Savings ~38 KB
**Risk**: low | **Complexity**: low

**Context**: `compress/zlib` is used in `fpdf/xcompr.go` to compress internal PDF streams.
In WASM, the PDF is generated for local viewing/printing, not transmitted over the network.
Uncompressed PDF is ~2-3x larger but works the same in any viewer (it's part of the PDF spec).

**What to do**:
1. Add `//go:build !wasm` to `fpdf/xcompr.go`.
2. Create `fpdf/xcompr_wasm.go` (`//go:build wasm`) with no-op functions:
   - The compression function simply writes the uncompressed bytes.
   - Maintain the same signature/interface.

**Functions in `fpdf/xcompr.go`** to review (read the file to identify exact signatures):
- Probably `compress()` or similar wrapping `zlib.Writer`.

**Files**:
- `fpdf/xcompr.go` → add `//go:build !wasm`.
- `fpdf/xcompr_wasm.go` → new, functions that pass uncompressed data.

**Validation**:
1. `wasmbuild` compiles.
2. `twiggy top ... | grep compress` → 0 results.
3. PDF generated in WASM opens correctly in Chrome/Firefox (without compression).
4. Backend: PDFs remain compressed.

---

### Stage 4: Replace time → tinywasm/time@v0.4.0 — Savings ~75 KB
**Risk**: medium | **Complexity**: medium

**Context**: `time` stdlib is used in 3 fpdf files:
- `fpdf/def.go:497-498` — `creationDate time.Time` and `modDate time.Time` fields.
- `fpdf/time.go` — Get/Set functions for creation/modification dates.
- `fpdf/fpdf.go:20-21` — global `creationDate`/`modDate time.Time` fields.
- `fpdf/document.go:898-901` — formats dates with `creation.Format("20060102150405")`.

**Available tinywasm/time API**:
- `time.Now() int64` — unix nanoseconds.
- `time.FormatISO8601(nano) string` — "YYYY-MM-DDTHH:MM:SSZ".
- `time.FormatDateTime(value) string` — "YYYY-MM-DD HH:MM:SS".
- **Does not have** the custom `"YYYYMMDDHHmmss"` format that PDF needs.

**What to do**:
1. Change the type of `creationDate`/`modDate` fields from `time.Time` to `int64` (unix nano) in WASM builds.
2. Create a `formatPDFDate(nano int64) string` helper that produces `"D:YYYYMMDDHHmmss"`:
   - Use `time.FormatISO8601(nano)` → `"2026-04-02T15:30:45Z"`.
   - Extract components using slicing: `s[0:4]+s[5:7]+s[8:10]+s[11:13]+s[14:16]+s[17:19]`.
   - Result: `"20260402153045"`.
3. `timeOrNow()` → if `nano == 0`, return `time.Now()`; otherwise, return nano.

**Files**:
- `fpdf/def.go` → split time fields: `fpdf/def_time_back.go` (!wasm, time.Time) and `fpdf/def_time_wasm.go` (wasm, int64).
- `fpdf/time.go` → split: `fpdf/time_back.go` (!wasm) and `fpdf/time_wasm.go` (wasm, tinywasm/time).
- `fpdf/fpdf.go` → split globals: creationDate/modDate fields by build tag.
- `fpdf/document.go:898-901` → split: PDF date format by build tag.

**Struct field details**:
```go
// fpdf/def.go:497 — currently:
creationDate     time.Time
modDate          time.Time

// fpdf/fpdf.go:20 — globals:
creationDate time.Time
modDate      time.Time
```
These fields are used in the files mentioned above. The split by build tag must be consistent — all files referencing these fields must compile with the correct type.

**Split strategy**: creating a type alias or abstract interface is NOT recommended — it generates unnecessary complexity. Better: split the files that declare and use these fields by build tag.

**Validation**:
1. `wasmbuild` compiles.
2. `twiggy top ... | grep "time\."` → 0 results (except for runtime).
3. PDF generated in WASM has correct CreationDate/ModDate.
4. Backend: everything remains the same with `time.Time`.

---

### Stage 5: encoding/json → tinywasm/json — Savings ~90 KB
**Risk**: medium | **Complexity**: medium

**Context**: `encoding/json` is used in 4 points of fpdf:
1. `fpdf/fonts.go:114` — `json.Unmarshal(jsonFileBytes, &info)` where `info` is `fontDefType`.
2. `fpdf/fonts.go:946` — `json.Unmarshal(buf.Bytes(), &def)` in `loadfont()`, same struct.
3. `fpdf/def.go:670` — `json.Marshal(&fdt)` in `generateFontID()` (already resolved in Stage 1).
4. `fpdf/font.go:307` — `json.Marshal(def)` to write font files — **backend only**.

**Structs that need `fmt.Fielder`**:

```go
// fpdf/def.go:647
type fontDefType struct {
    Tp           string        // "Core", "TrueType"
    Name         string        // "Courier-Bold"
    Desc         FontDescType  // nested struct
    Up           int           // underline position
    Ut           int           // underline thickness (actually int but declared as int)
    Cw           []int         // character widths (can have up to 65536 entries)
    Enc          string        // "cp1252"
    Diff         string        // differences
    File         string        // "font.z"
    Size1, Size2 int           // Type1 values
    OriginalSize int
    N            int
    DiffN        int
    i            string        // private, not serialized by JSON (lowercase)
    utf8File     *utf8FontFile // private, not serialized
    usedRunes    map[int]int   // private, not serialized
}

// fpdf/def.go:610
type FontDescType struct {
    Ascent, Descent, CapHeight, Flags int
    FontBBox FontBoxType  // another nested struct
    ItalicAngle float64
    StemV, MissingWidth int
}
```

**`fmt.Fielder` implementation**:
- Only public fields are serialized (same semantics as `encoding/json`).
- `Cw []int` is the largest field — `tinywasm/json` must support `[]int`.
- `FontDescType` and `FontBoxType` are sub-structs — verify that `tinywasm/json` supports nested structs via `Fielder`.

**IMPORTANT**: verify that `tinywasm/json.Unmarshal` supports:
- `[]int` fields (slices of primitives).
- Nested structs that implement `Fielder`.
- `float64` fields.

**What to do**:
1. Implement `Schema()` and `Pointers()` in `fontDefType`, `FontDescType`, `FontBoxType`.
2. Split: `fpdf/fonts_json_back.go` (!wasm, `encoding/json`) and `fpdf/fonts_json_wasm.go` (wasm, `tinywasm/json`).
3. Unmarshal functions are abstracted: `unmarshalFontDef(data []byte, def *fontDefType) error`.
4. `font.go:307` is already backend only (uses `os.Create`) — add `!wasm` build tag if it doesn't have it.

**Files**:
- `fpdf/def.go` → implement `Fielder` in `fontDefType`, `FontDescType`, `FontBoxType`.
- `fpdf/fonts.go` → extract `json.Unmarshal` calls to a helper function by build tag.
- `fpdf/fonts_json_back.go` → (!wasm) uses `encoding/json`.
- `fpdf/fonts_json_wasm.go` → (wasm) uses `tinywasm/json`.
- `fpdf/font.go` → verify that it already has the `!wasm` build tag (uses `os.Create`).

**Validation**:
1. `wasmbuild` compiles.
2. `twiggy top ... | grep "encoding/json"` → 0 results.
3. Loading fonts works: generate PDF with text using embedded font.
4. Verify that character widths (Cw) are parsed correctly — test with varied text.
5. Backend: fonts still load with `encoding/json`.

---

### Stage 6: Remove fmt stdlib (indirect) — Savings ~45 KB
**Risk**: low | **Complexity**: low (if previous stages are completed)

**Context**: `fmt` stdlib is NOT imported directly in production code.
It is dragged in by:
- `encoding/json` (reflection → fmt for `Stringer`) — removed in Stage 5.
- `crypto/sha1` → `fmt.Sprintf("%x", ...)` in `generateFontID` — removed in Stage 1.
- Possibly other indirect imports.

**What to do**:
1. After completing Stages 1-5, compile and verify with `twiggy`.
2. If `fmt` stdlib persists: `twiggy paths ... fmt.pp` to trace what imports it.
3. Remove the remaining import chain.

**If it persists**, possible culprits:
- `errors.New` in some packages may drag in `fmt` via the `error` interface.
- `strconv` may drag in `fmt` (verify).
- Some transitive import from `tinywasm/json` or `tinywasm/time`.

**Validation**:
1. `twiggy top ... | grep "fmt\."` → 0 results (only `tinywasm/fmt`).
2. Everything still works.

---

## Execution Order

```
Stage 1 (crypto/protection) ──┐
Stage 2 (image/gif)          ──┼── Parallel, low risk (~80 KB)
Stage 3 (compress/zlib)      ──┘

Stage 4 (time → tinywasm/time) ── Independent (~75 KB)

Stage 5 (json → tinywasm/json) ── Major impact (~90 KB)

Stage 6 (fmt stdlib cleanup)   ── Verify after Stages 1-5 (~45 KB)
```

**Each stage is an independent commit** to facilitate rollback.

## Validation by stage
Each stage MUST:
1. Compile without errors: `wasmbuild`.
2. Measure size: `ls -lh web/public/client.wasm`.
3. Analyze with `twiggy`: `twiggy top web/public/client.wasm -n 30`.
4. Functional test: generate PDF with text, table, chart and image in WASM.
5. Backend tests: `go test ./...` (without wasm build tag).

## Round 2 (if 150 KB is not reached)
Evaluate in a separate plan:
1. Remove `regexp` → manual parsing in `htmlbasic.go` and `ttfparser.go`.
2. Remove `image/jpeg` + `image/png` → decode in JS (Canvas API) and pass raw pixels.
3. Strip `fpdf` dead code: layers, attachments, gradients, spot colors.
4. Strip "function names" subsection from WASM (~36 KB).

## tinywasm dependencies to use
- `github.com/tinywasm/json` — `encoding/json` replacement (zero-reflection, Fielder-based).
- `github.com/tinywasm/time` — `time` stdlib replacement (int64 unix nano, browser Date.now()).
- `github.com/tinywasm/unixid` — `crypto/sha1` replacement for `generateFontID`.
- `github.com/tinywasm/fmt` — already in use.

## Global Risks
- **Functionality regression**: each stage has stubs that maintain the public API.
- **Maintainability**: more files with build tags — mitigated by existing pattern (`env.front.go`/`env.back.go`).
- **Fonts (Stage 5)**: more delicate — corrupt fonts = unreadable PDFs. Test with multiple fonts and texts.
- **Uncompressed PDF (Stage 3)**: part of the PDF spec, all viewers support it.
- **tinywasm/time PDF format**: does not have a custom format — a helper with string slicing of `FormatISO8601` is needed.
