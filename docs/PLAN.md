# PLAN: WASM binary size — Round 2

## Current Status: 364.3 KB (down from 738 KB — 50.6% reduction)

## Remaining target: ~150 KB

## Completed (Round 1)
- encoding/json → tinywasm/json: eliminated
- time stdlib → tinywasm/time: eliminated
- crypto/* (protection, sha1, md5): eliminated
- image/gif: eliminated
- compress/* (zlib write): eliminated (120 bytes remain for PNG decompress, acceptable)
- stdlib fmt: eliminated
- os from shared code: eliminated
- Duplicated build-tag files unified: fontid, fonts_json, time, types

## Completed (Bug fixes — post Round 1)
- **Blank PDF in WASM**: `f.compress` defaulted to `true` but `xmem.compress()` in WASM copies bytes without compressing. PDF headers declared `/Filter /FlateDecode` causing viewers to fail decompressing raw data → blank pages.
  - Added `init() { gl.noCompress = true }` in `fpdf/xcompr_wasm.go`
  - Added `f.compress` checks in `fpdf/fonts.go` (~lines 793, 807) around cidToGidMap and font stream FlateDecode headers
  - Added `f.compress` check in `fpdf/document.go` (~line 1062) for ICC profile FlateDecode header
- **def_gob.go build tag**: Fixed from `//go:build !tinygo && !js && !wasm` to `//go:build !wasm`

## Current breakdown (twiggy, 364.3 KB)

| Component | KB | % | Reducible? |
|---|---|---|---|
| web/ui.setupUI (demo) | 61 KB | 16.8% | YES — not part of the lib |
| runtime | 45 KB | 11.9% | NO — irreducible |
| rodata segments | ~45 KB | 12% | PARTIAL — reduces with code |
| tinywasm/fmt | ~25 KB | 6.7% | NO — core dependency |
| fpdf core | ~30 KB | 8% | PARTIAL — dead code |
| image/jpeg | ~16 KB | 4.3% | MAYBE — decode in JS |
| function names subsection | 12 KB | 3.2% | YES — debug metadata |
| tinywasm/json | ~10 KB | 2.7% | NO — needed for fonts |
| image/png | ~8 KB | 2.1% | MAYBE — decode in JS |
| other (fetch, syscall/js, etc) | ~112 KB | ~30% | Various |

## Blockers
- **tinywasm/json FieldIntSlice**: `tinywasm/fmt` and `tinywasm/json` need `FieldIntSlice` support for `fontDefType.Cw []int`. Plans created in both repos (`tinywasm/fmt/docs/PLAN.md` and `tinywasm/json/docs/PLAN.md`). Must be completed before fonts work with unified `tinywasm/json` on backend tests.

## Known Issues (not blocking optimization but pending)
- **UTF-8 garbled characters on backend**: `loadDefaultFont()` in `document.go:44-51` reads `fonts/Arial.ttf` relative to CWD via `readFile()`. If file not found, silently falls back to built-in Latin-1 Arial. Backend demos show garbled characters (e.g. "LÃ-nea") because the font file path doesn't resolve. Fix: ensure the caller sets the correct working directory or provides an absolute path to `DefaultFontPath`.

---

## STAGES

### Stage 1: Strip "function names" WASM subsection — Savings ~12 KB
**Risk**: low | **Complexity**: low

**Context**: The "function names" subsection (12 KB, 3.2%) is debug metadata embedded in the WASM binary. It contains human-readable names for every function — useful for debugging but unnecessary in production.

**What to do**:
1. Run `wasmbuild` with current config and note exact file size: `ls -lh web/public/client.wasm`
2. Try `wasm-opt --strip-debug -O2 web/public/client.wasm -o web/public/client.wasm` (install via `apt install binaryen` if missing)
3. If `wasm-opt` not available, try `wasm-tools strip web/public/client.wasm -o web/public/client.wasm` (install via `cargo install wasm-tools` if missing)
4. If neither tool is available, check if TinyGo's `-no-debug` flag works: modify `wasmbuild` command to add it

**Validation**:
1. `ls -lh web/public/client.wasm` — verify ~12 KB reduction
2. Load in browser — WASM still loads, PDF generates correctly with text, table, chart, image
3. `twiggy top web/public/client.wasm -n 30` — confirm "function names" subsection gone or reduced

---

### Stage 2: Remove fpdf dead code — Savings ~10-20 KB estimated
**Risk**: medium | **Complexity**: medium

**Context**: fpdf has many features not used in the WASM scope (generate-only PDFs with text, tables, charts, images). Unused features add code that TinyGo may not fully eliminate. Several files have NO build tag and compile into WASM unnecessarily.

**Candidate files to evaluate** (all currently lack `//go:build` tags and compile in WASM):

| File | Feature | Heavy import | Action if unused |
|---|---|---|---|
| `fpdf/htmlbasic.go` | HTML parser | `regexp` (HEAVY) | Add `//go:build !wasm` |
| `fpdf/layer.go` | PDF layers | none | Add `//go:build !wasm` |
| `fpdf/spotcolor.go` | Spot colors | none | Add `//go:build !wasm` |
| `fpdf/fpdftrans.go` | Transformations | none | Add `//go:build !wasm` if not used by charts |
| `fpdf/grid.go` | Grid drawing | none | Add `//go:build !wasm` if not used by charts |
| `fpdf/javascripts.go` | PDF JavaScript | none | Add `//go:build !wasm` |
| `fpdf/subwrite.go` | Subscript/superscript | none | Add `//go:build !wasm` if unused |
| `fpdf/label.go` | Axis label formatting | none | Keep if used by charts |
| `fpdf/font_afm.go` | AFM font parser | `bufio` | Add `//go:build !wasm` if unused |

**What to do for each file**:
1. Find all exported functions in the file: `grep -E '^func \(f \*Fpdf\)|^func [A-Z]' fpdf/<file>.go`
2. For each function, check if it's called from WASM code paths: `grep -r '<FunctionName>' fpdf/ web/ --include='*.go' -l`
3. Exclude files that have `//go:build !wasm` already (they won't match)
4. If NO WASM code path calls any function in the file → add `//go:build !wasm` at the top
5. If some functions are used and others not → leave the file as-is (don't split unless the savings are significant)
6. **Special case `htmlbasic.go`**: This imports `regexp` which is very heavy in TinyGo. Prioritize confirming whether `WriteHTML` or related functions are called from WASM paths. If not → `//go:build !wasm` gives the biggest single-file win.

**Validation**:
1. `wasmbuild` compiles without errors
2. `ls -lh web/public/client.wasm` — measure reduction
3. `twiggy top web/public/client.wasm -n 30` — verify removed components
4. `go test ./...` — backend tests still pass (files still compile for backend)
5. Browser: generate PDF with text, table, chart, image — all render correctly

---

### Stage 3: Evaluate image decoders → JS Canvas API — Savings ~24 KB
**Risk**: high | **Complexity**: high

**Context**: `image/jpeg` (16 KB) + `image/png` (8 KB) = 24 KB. In browser, images can be decoded via Canvas API (`createImageBitmap` + `getImageData`) and passed as raw RGBA pixels to Go.

**Trade-offs**:

| Aspect | Go decoders (current) | JS Canvas API |
|---|---|---|
| Binary size | +24 KB | 0 KB |
| Complexity | Low (direct import) | High (async JS interop) |
| Format support | PNG, JPEG only | All browser-supported formats |
| Color space | Controlled | Browser-dependent |
| Portability | Works everywhere | Browser-only |

**Decision**: evaluate in separate plan if Stages 1-2 don't reach target. The complexity is high and the savings moderate.

---

## Execution Order

```
Stage 1 (strip function names) ── Quick win (~12 KB)
Stage 2 (fpdf dead code)       ── Medium effort (~10-20 KB)
Stage 3 (image → JS Canvas)    ── Only if needed, separate plan
```

## Note on web/ui.setupUI (61 KB)
This is the demo application, not the library itself. It does not affect the size of the library when used by other projects. No action needed — it only appears in this binary because `web/client.go` imports it.

## Validation per stage
Each stage MUST:
1. Compile: `wasmbuild`
2. Measure: `ls -lh web/public/client.wasm`
3. Analyze: `twiggy top web/public/client.wasm -n 30`
4. Functional: generate PDF with text, table, chart and image in WASM
5. Backend: `go test ./...`
