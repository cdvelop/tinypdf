---
name: Font refactor — fontManager
about: Move font handling to a standalone fontManager package, add automatic TTF discovery and wasm/std loaders
title: feat(fontManager): automatic TTF discovery + wasm/std loaders
labels: enhancement, refactor
assignees: ''
---

## Context
Tinypdf is being refactored to run in the browser (WASM) and be TinyGo-compatible. We need a clean, TinyGo-friendly redesign of font handling.

## Goal
Move font handling into a standalone `fontManager/` package that automatically discovers and registers TTF fonts from a directory, and provide separate loaders for WASM and standard builds.

## Tasks
- [x] Create `fontManager/` package with the following files:
  - `loader.go` (shared detection/mapping logic)
  - `getStdlib.go` (`//go:build !wasm`) — uses `os.ReadDir`/`os.ReadFile`
  - `getWasm.go` (`//go:build wasm`) — uses `syscall/js` to obtain files from the host/browser
  - `font_id.go` — generate IDs with `github.com/cdvelop/unixid`
  - tests: `loader_test.go` and additional unit tests
- [x] Implement exported API: `LoadFonts(dir string) (map[string]FontFamily, error)` (or equivalent) and document `FontFamily` structure (Regular/Bold/Italic).
- [ ] Modify TinyPDF constructor (`tinypdf.go`) to load fonts from default `fonts/` directory at startup.
- [ ] Add `func (t *TinyPDF) ChangeFontsDir(newPath string) error` to change and reload fonts at runtime.
- [ ] Replace prohibited stdlib usage in new code: use `github.com/cdvelop/tinystring` instead of `fmt`, `strconv`, `errors` where applicable.
- [ ] Replace `crypto/sha1` usage for font IDs with `github.com/cdvelop/unixid`.
- [ ] Add `fonts/arial.ttf` as a test asset and write tests that reuse existing local fonts.
- [ ] Remove or migrate `font_embed/` (no default embedded fonts remain).
- [ ] Update docs: `./fontManager.md` describing the new API and behavior.
- [ ] Run `go test ./...` and fix any failing tests.

## Must requirements
- Support only `.ttf` fonts.
- Font detection must be name-based. Heuristics (case-insensitive):
  - Family = base filename (strip common suffixes like `-regular`, `-bold`, `-italic`).
  - Regular: contains `regular`/`normal` or no suffix.
  - Bold: contains `bold` or `b`.
  - Italic: contains `italic` or `it`.
  - If only one file exists for a family, register it for Regular/Bold/Italic to preserve minimal compatibility.
- No legacy compatibility bridges; remove embedded defaults.
- Constructor parameter types must be static (avoid `any`/loose string configs for typed values).
- No use of `fmt`, `strconv`, `errors` in new code; use `github.com/cdvelop/tinystring`.

## Questions to resolve (answer before implementation)
1. If a family has only one TTF file, should we register it for all styles? (Recommended: Yes)
2. Style detection: prefer a simple substring heuristic or a broader suffix list? (Recommend start simple and document.)
3. In WASM, how will the browser expose font files? (Options: host JS provides directory listing, or user passes File objects.)
4. Error policy when required font is missing: strict failure (preferred) or warn+continue?
5. Tests for WASM: mock the loader or provide JS fixtures?

## Acceptance criteria
- `fontManager` provides a stdlib loader (`!wasm`) and a wasm loader (`wasm`) with correct build tags.
- TinyPDF `New` loads fonts by default from `fonts/` and `ChangeFontsDir` works.
- Tests added for font discovery and TinyPDF integration; `fonts/arial.ttf` is used as test asset.
- New code uses `github.com/cdvelop/tinystring` and `github.com/cdvelop/unixid` where required.
- `font_embed` directory removed or migrated to `fontManager`.
- Documentation updated and `go test ./...` passes (or failing tests are documented with reasons).

## How to test locally
1. Run unit tests for `fontManager`:
   ```sh
   go test ./fontManager -v
   ```
2. Run full test suite:
   ```sh
   go test ./... 
   ```
3. Manual check: create a TinyPDF instance and call `ChangeFontsDir("fonts/")`, then run a small generation flow that exercises font registration.

## Notes
- Keep the `fontManager` package independent and reusable outside tinypdf.
- Implement stdlib first (easier to test), then add wasm loader and a small wasm stub for CI later.
