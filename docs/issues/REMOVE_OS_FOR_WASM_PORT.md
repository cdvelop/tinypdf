# Remove OS Package Dependency for WASM Portability

## Current State

The `fpdf` package uses `os` package in multiple locations, blocking WASM portability:

**Main Usage Areas:**
- `font.go`: `os.Open()`, `os.ReadFile()` - Font loading (AFM, TTF, encoding maps)
- `fonts.go`: `os.Stat()`, `os.ReadFile()`, `os.Open()` - UTF-8 font files
- `util.go`: `os.Stat()`, `os.Open()` - File validation, Unicode translators
- `fpdf.go`: `os.Open()` - Image loading
- `ttfparser.go`: `os.Open()` - TTF parsing
- `svgbasic.go`: `os.ReadFile()` - SVG loading
- `document.go`: Currently uses `writeFile` injected function ✓

**Current Injection:**
```go
writeFile func(filePath string, content []byte) error
```

## Objective

Eliminate all `os` package usage by injecting file I/O functions, enabling clean frontend (WASM) and backend implementations.

## Proposed Strategy

Replace direct `os` calls with injected function interfaces. Each environment provides implementation:
- **Backend**: Standard `os` package functions
- **Frontend/WASM**: HTTP fetch, IndexedDB, or memory-based implementations

## Refactoring Items

### 1. File Reading Interface
**File**: `docs/issues/proposals/01_FILE_READER_INTERFACE.md`
**Scope**: Replace all `os.Open()` and `os.ReadFile()` calls

### 2. File Validation Operations  
**File**: `docs/issues/proposals/02_FILE_VALIDATION.md`
**Scope**: Handle `os.Stat()`, `fileExist()`, `fileSize()` in util.go

### 3. TTF Parser File Access
**File**: `docs/issues/proposals/03_TTF_PARSER_IO.md`
**Scope**: Refactor `ttfparser.go` file operations

### 4. Font Loader Enhancement
**File**: `docs/issues/proposals/04_FONT_LOADER_EXPANSION.md`
**Scope**: Extend existing `FontLoader` interface pattern

### 5. Backward Compatibility Strategy
**File**: `docs/issues/proposals/05_BACKWARD_COMPAT.md`
**Scope**: Default implementations for existing code

## Success Criteria

- Zero `os` package imports in production code (tests excluded)
- Backend usage unchanged with default implementations
- WASM builds compile without filesystem dependencies
- No breaking changes for current API consumers (if possible)
- Tests only need to update `NewDocPdfTest()` helper (similar to current `writeFile` injection)

## Evaluation Required

Each proposal document contains:
- Specific implementation approach
- Pros/cons analysis
- Breaking change assessment
- Migration effort estimation
