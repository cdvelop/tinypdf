# Remove Embedded Fonts Support - Refactoring Guide

## Objective

Remove all embedded font support from `tinypdf/fpdf` package and replace it with a configurable `fontLoader` function that loads TTF fonts from filesystem (backend) or URL (frontend).

## Related Documentation

- **[TTF Font Parsing from Bytes](../TTF_PARSE_BYTES.md)** - Technical documentation on how `tinypdf/fpdf` parses TrueType fonts from `[]byte` data. Essential for understanding how fontLoader integrates with the existing TTF parser.

## Breaking Changes

- All embedded fonts removed (courier, helvetica, times, symbol, zapfdingbats)
- Only TTF font format supported
- No default fonts available
- Users must provide `fontLoader` function or get an error
- `FontLoader` interface replaced with simple function signature

## Architecture Changes

### Current System
```go
// Interface-based font loading
type FontLoader interface {
    Open(name string) (io.Reader, error)
}

// Embedded fonts via embed.FS
//go:embed font_embed/*.json font_embed/*.map
var embFS embed.FS
```

### New System
```go
// Simple function-based font loading
fontLoader func(fontPath string) ([]byte, error)

// No embedded fonts
// No FontLoader interface
```

## Files to Modify

### 1. DELETE: `fpdf/embedded.go`
**Action:** Remove entire file
- Contains `embFS` embed directive
- Contains `coreFontReader()` method
- No longer needed

### 2. DELETE: `fpdf/font_embed/` directory
**Action:** Remove entire directory and all contents
- Contains `.json` and `.map` embedded font files
- No longer needed

### 3. MODIFY: `fpdf/def.go`

#### Remove FontLoader interface
**Location:** Line ~392-395

**Current:**
```go
type FontLoader interface {
    Open(name string) (io.Reader, error)
}
```

**Action:** Delete entire interface definition

#### Modify Fpdf struct
**Location:** Line ~420-532 (Fpdf struct)

**Find field:**
```go
fontLoader       FontLoader                                  // used to load font files from arbitrary locations
```

**Replace with:**
```go
fontLoader       func(fontPath string) ([]byte, error)       // function to load TTF font files
```

#### Remove coreFonts field
**Location:** Line ~420-532 (Fpdf struct)

**Find field:**
```go
coreFonts        map[string]bool                             // array of core font names
```

**Action:** Delete this field completely

#### Add fontCache field
**Location:** Line ~420-532 (Fpdf struct)

**Add field after fontLoader:**
```go
fontCache        []fontCacheEntry                            // slice to cache loaded fonts (TinyGo compatible)
```

**Add fontCacheEntry type before Fpdf struct:**
```go
// fontCacheEntry stores cached font data
type fontCacheEntry struct {
    path string
    data []byte
}
```

### 4. MODIFY: `fpdf/fpdf.go`

#### Update New() function - Remove coreFonts initialization
**Location:** Line ~32-240

**Find:**
```go
// Core fonts
f.coreFonts = map[string]bool{
    "courier":      true,
    "helvetica":    true,
    "times":        true,
    "symbol":       true,
    "zapfdingbats": true,
}
```

**Action:** Delete entire coreFonts initialization block

#### Update New() function - Initialize fontLoader with error
**Location:** Line ~32-240 (after readFile/writeFile/fileSize initialization)

**Find:**
```go
// Initialize fileSize with a function that returns an error by default
f.fileSize = func(filePath string) (int64, error) {
    return 0, Errf("fileSize function not configured for this environment")
}
```

**Add after:**
```go
// Initialize fontLoader with a function that returns an error by default
f.fontLoader = func(fontPath string) ([]byte, error) {
    return nil, Errf("fontLoader function not configured for this environment")
}
// Initialize fontCache as empty slice
f.fontCache = make([]fontCacheEntry, 0, 8)
```

#### Update New() function - Add fontLoader option handling
**Location:** Line ~32-240 (inside options loop)

**Find:**
```go
case FileSizeFunc:
    f.fileSize = v
```

**Add after:**
```go
case func(string) ([]byte, error):
    f.fontLoader = v
```

### 5. MODIFY: `fpdf/font.go`

**Action:** Search for all usages of:
- `coreFontReader()`
- `f.coreFonts`
- `embFS`
- JSON font parsing
- `.z` font file handling

**Important:** Integration with TTF parser documented in [TTF_PARSE_BYTES.md](../TTF_PARSE_BYTES.md)

#### Create TtfParseBytes() wrapper function

**Add to `fpdf/ttfparser.go`:**

```go
// TtfParseBytes extracts various metrics from a TrueType font byte data.
// This is a wrapper around TtfParse for direct []byte input.
func TtfParseBytes(fontData []byte) (TtfRec TtfType, err error) {
    var t ttfParser
    t.file = bytes.NewReader(fontData)

    version, err := t.ReadStr(4)
    if err != nil {
        return
    }
    if version == "OTTO" {
        err = Errf("fonts based on PostScript outlines are not supported")
        return
    }
    if version != "\x00\x01\x00\x00" {
        err = Errf("unrecognized file format")
        return
    }
    numTables := int(t.ReadUShort())
    t.Skip(3 * 2) // searchRange, entrySelector, rangeShift
    t.tables = make(map[string]uint32)
    var tag string
    for j := 0; j < numTables; j++ {
        tag, err = t.ReadStr(4)
        if err != nil {
            return
        }
        t.Skip(4) // checkSum
        offset := t.ReadULong()
        t.Skip(4) // length
        t.tables[tag] = offset
    }
    err = t.ParseComponents()
    if err != nil {
        return
    }
    TtfRec = t.rec
    return
}
```

#### Replace coreFontReader() calls

**Find pattern:**
```go
r := f.coreFontReader(familyStr, styleStr)
```

**Replace with:**
```go
fontPath := familyStr + styleStr + ".ttf"
fontData, err := f.fontLoader(fontPath)
if err != nil {
    f.SetError(err)
    return
}
// Parse TTF font data using existing parser
// See: docs/TTF_PARSE_BYTES.md for parser details
ttfInfo, err := TtfParseBytes(fontData)
if err != nil {
    f.SetError(err)
    return
}
// Convert ttfInfo to font definition and store in f.fonts
```

**Note:** The `TtfParseBytes()` function needs to be created as a wrapper around the existing `TtfParse()` that accepts `[]byte` directly. See [TTF_PARSE_BYTES.md](../TTF_PARSE_BYTES.md) for implementation details.

#### Remove core font checks

**Find pattern:**
```go
if f.coreFonts[familyStr] {
    // special handling for core fonts
}
```

**Action:** Remove all core font special cases. All fonts must be loaded via `fontLoader`.

#### Update font caching

**Current:** Fonts cached after loading from embed or filesystem

**New:** Fonts cached in `f.fontCache` slice after loading from `fontLoader` (TinyGo compatible)

**Implementation:**
```go
// Helper function to check cache (add to fpdf.go)
func (f *Fpdf) getCachedFont(fontPath string) ([]byte, bool) {
    for i := 0; i < len(f.fontCache); i++ {
        if f.fontCache[i].path == fontPath {
            return f.fontCache[i].data, true
        }
    }
    return nil, false
}

// Helper function to add to cache (add to fpdf.go)
func (f *Fpdf) addFontToCache(fontPath string, data []byte) {
    f.fontCache = append(f.fontCache, fontCacheEntry{
        path: fontPath,
        data: data,
    })
}

// In font loading logic (in font.go)
fontPath := buildFontPath(family, style)

// Check cache first
if cachedData, found := f.getCachedFont(fontPath); found {
    // Use cached data
    parsedFont := parseTTF(cachedData)
    f.fonts[fontKey] = parsedFont
    return
}

// Load font data
fontData, err := f.fontLoader(fontPath)
if err != nil {
    f.SetError(err)
    return
}

// Cache the raw font data
f.addFontToCache(fontPath, fontData)

// Parse TTF and store in fonts map
parsedFont := parseTTF(fontData)
f.fonts[fontKey] = parsedFont
```

### 6. MODIFY: `fpdf/exampleDir_test.go`

#### Update NewDocPdfTest() function
**Location:** Line ~28-59

**Find:**
```go
pdf := tinypdf.New(options...)
```

**Add before:**
```go
// add default fontLoader function using os for tests
options = append(options, func(fontPath string) ([]byte, error) {
    // Build full path: rootTestDir/fonts/fontPath
    fullPath := filepath.Join(string(rootTestDir), "fonts", fontPath)
    return os.ReadFile(fullPath)
})

pdf := tinypdf.New(options...)
```

### 7. MODIFY: `tinypdf.go`

#### Update TinyPDF struct

**Current:**
```go
type TinyPDF struct {
    Fpdf   *fpdf.Fpdf
    logger func(message ...any)
}
```

**Add field:**
```go
type TinyPDF struct {
    Fpdf       *fpdf.Fpdf
    logger     func(message ...any)
    fontLoader func(fontPath string) ([]byte, error)
}
```

#### Update New() function
**Location:** Line ~20-35

**Find:**
```go
// Crear instancia de Fpdf con las opciones y las funciones de IO
options = append(options, fpdf.WriteFileFunc(tp.writeFile))
options = append(options, fpdf.ReadFileFunc(tp.readFile))
options = append(options, fpdf.FileSizeFunc(tp.fileSize))

tp.Fpdf = fpdf.New(options...)
```

**Replace with:**
```go
// Crear instancia de Fpdf con las opciones y las funciones de IO
options = append(options, fpdf.WriteFileFunc(tp.writeFile))
options = append(options, fpdf.ReadFileFunc(tp.readFile))
options = append(options, fpdf.FileSizeFunc(tp.fileSize))
options = append(options, tp.fontLoader)

tp.Fpdf = fpdf.New(options...)
```

### 8. MODIFY: `env.back.go`

**Create new file:** `env.back.go` (if doesn't exist) or modify existing

#### Add fontLoader initialization in initIO()

**Location:** After logger initialization

**Add:**
```go
func (tp *TinyPDF) initIO() {
    // Inicializar logger para backend usando fmt.Println
    tp.logger = func(message ...any) {
        fmt.Println(message...)
    }
    
    // Inicializar fontLoader para backend usando os.ReadFile
    tp.fontLoader = func(fontPath string) ([]byte, error) {
        // fontPath comes as relative path like "fonts/Arial.ttf"
        // or "customfonts/MyFont.ttf"
        return os.ReadFile(fontPath)
    }
}
```

#### Update writeFile (if needed)

**No changes needed** - already uses `os.WriteFile`

#### Update readFile (if needed)

**No changes needed** - already uses `os.ReadFile`

### 9. CREATE/MODIFY: `env.front.go`

**Location:** Existing file for WASM environment

**Important:** Add dependency to `go.mod`:
```bash
go get github.com/tinywasm/fetch
```

#### Add imports

**Add to imports:**
```go
import (
    "fmt"
    "syscall/js"
    "github.com/tinywasm/fetch"
)
```

#### Add fontLoader initialization in initIO()

**Find:**
```go
func (tp *TinyPDF) initIO() {
    // Inicializar logger para frontend usando console.log
    tp.logger = func(message ...any) {
        console := js.Global().Get("console")
        if !console.IsUndefined() {
            console.Call("log", fmt.Sprint(message...))
        }
    }
}
```

**Add after logger:**
```go
    // Inicializar fontLoader para frontend usando fetchgo
    tp.fontLoader = tp.loadFontFromURL
}

// loadFontFromURL loads TTF fonts from current domain using fetchgo
// Note: Caching is handled by fpdf.Fpdf.fontCache, not here
func (tp *TinyPDF) loadFontFromURL(fontPath string) ([]byte, error) {
    // Build URL using current domain
    // fontPath comes as "fonts/Arial.ttf"
    location := js.Global().Get("location")
    if location.IsUndefined() {
        return nil, fmt.Errorf("window.location not available")
    }
    
    origin := location.Get("origin").String()
    fullURL := origin + "/" + fontPath
    
    // Create fetchgo client for single request
    client := &fetchgo.Client{
        RequestType: fetchgo.RequestRaw, // We want raw bytes
    }
    
    // Channel to receive result
    resultChan := make(chan []byte, 1)
    errorChan := make(chan error, 1)
    
    // Send GET request for font file
    client.SendRequest("GET", fullURL, nil, func(result any, err error) {
        if err != nil {
            errorChan <- fmt.Errorf("failed to fetch font %s: %w", fontPath, err)
            return
        }
        
        // Result should be []byte
        if fontData, ok := result.([]byte); ok {
            resultChan <- fontData
        } else {
            errorChan <- fmt.Errorf("unexpected result type from fetchgo: %T", result)
        }
    })
    
    // Wait for result (blocking call)
    select {
    case data := <-resultChan:
        return data, nil
    case err := <-errorChan:
        return nil, err
    }
```

**Note:** This implementation uses `fetchgo` library which handles the fetch API complexity. Font caching is handled by `fpdf.Fpdf.fontCache` slice, not in this function.

### 10. SEARCH AND REPLACE: All references

#### Search for these patterns across all files:

1. **`coreFontReader`** - Remove all calls, replace with `fontLoader`
2. **`f.coreFonts`** - Remove all references
3. **`embFS`** - Remove all references
4. **`embed.FS`** - Remove import and usage
5. **`.json` font files** - Replace with `.ttf` font loading
6. **`.map` font files** - Remove all references
7. **`.z` compressed fonts** - Remove all references
8. **`FontLoader` interface** - Replace with function type

#### Files likely to contain references:

- `fpdf/font.go` - Main font loading logic
- `fpdf/font_afm.go` - AFM font support (may need updates)
- `fpdf/fonts.go` - Font utilities
- `fpdf/fonts_test.go` - Font tests (update to use TTF)
- `fpdf/def_test.go` - Definition tests
- `fpdf/fpdf_test.go` - Main tests

## Implementation Steps

### Phase 1: Preparation
1. Backup current codebase
2. Ensure all tests pass before changes
3. Identify all font usage in tests
4. Prepare TTF font files for testing

### Phase 2: Core Changes
1. Delete `fpdf/embedded.go`
2. Delete `fpdf/font_embed/` directory
3. Modify `fpdf/def.go` - Remove FontLoader interface, update Fpdf struct
4. Modify `fpdf/fpdf.go` - Update New() function
5. Modify `fpdf/font.go` - Replace all font loading logic

### Phase 3: Environment-specific Changes
1. Modify `env.back.go` - Add fontLoader for filesystem
2. Modify `env.front.go` - Add fontLoader for URL fetch with caching
3. Modify `tinypdf.go` - Update New() to pass fontLoader

### Phase 4: Test Updates
1. Modify `fpdf/exampleDir_test.go` - Update NewDocPdfTest()
2. Update all test files that use fonts
3. Add TTF font files to test fixtures
4. Verify all tests pass

### Phase 5: Cleanup
1. Search for any remaining references to:
   - `coreFonts`
   - `coreFontReader`
   - `embFS`
   - `FontLoader` interface
   - JSON font format
   - `.z` font files
2. Remove unused imports (embed, io, etc.)
3. Update comments and documentation

## Testing Strategy

### Backend Tests
```go
// Test with local TTF file
pdf := NewDocPdfTest()
pdf.AddFont("Arial", "", "fonts/Arial.ttf")
```

### Frontend Tests
```go
// Test with URL fetch (mocked)
tp := tinypdf.New()
// Should fetch from current domain/fonts/Arial.ttf
pdf.AddFont("Arial", "", "fonts/Arial.ttf")
```

### Error Cases
```go
// Test missing fontLoader
pdf := fpdf.New() // No fontLoader provided
pdf.AddFont("Arial", "", "fonts/Arial.ttf")
// Should return: "fontLoader function not configured"

// Test missing font file
pdf := tinypdf.New(fontLoader)
pdf.AddFont("Missing", "", "fonts/Missing.ttf")
// Should return: "file not found" or "fetch failed"
```

## Key Points

1. **No default fonts** - All fonts must be explicitly loaded
2. **TTF only** - No JSON, MAP, or Z font formats supported
3. **Caching in Fpdf struct** - Fonts cached in `fontCache` slice (TinyGo compatible, no maps)
4. **Relative paths** - Font paths are relative (e.g., "fonts/Arial.ttf")
5. **Environment-specific** - Backend uses filesystem, frontend uses fetch
6. **No base64** - Fonts stored as raw []byte in memory
7. **Breaking change** - Existing code will break if using embedded fonts
8. **TinyGo compatible** - Uses slices instead of maps for caching

## Expected Behavior After Refactoring

### Backend (Go)
```go
tp := tinypdf.New()
// Internally calls os.ReadFile("fonts/Arial.ttf")
tp.Fpdf.AddFont("Arial", "", "fonts/Arial.ttf")
```

### Frontend (WASM)
```go
tp := tinypdf.New()
// Internally fetches https://currentdomain.com/fonts/Arial.ttf
// Caches in fpdf.Fpdf.fontCache slice for subsequent use
tp.Fpdf.AddFont("Arial", "", "fonts/Arial.ttf")
```

### Tests
```go
pdf := NewDocPdfTest()
// Uses rootTestDir/fonts/ as base path
pdf.AddFont("Arial", "", "Arial.ttf")
```

## Migration Notes for Users

Users upgrading to this version must:

1. Remove all references to core fonts (courier, helvetica, times, symbol, zapfdingbats)
2. Provide TTF font files
3. Use `AddFont()` to explicitly load all fonts before use
4. For backend: Place TTF files in accessible filesystem location
5. For frontend: Host TTF files on same domain, accessible via HTTP

## Validation Checklist

- [ ] `fpdf/embedded.go` deleted
- [ ] `fpdf/font_embed/` directory deleted
- [ ] `FontLoader` interface removed from `def.go`
- [ ] `fontCacheEntry` type added to `def.go`
- [ ] `fontLoader` function field added to Fpdf struct
- [ ] `fontCache` slice field added to Fpdf struct
- [ ] `coreFonts` field removed from Fpdf struct
- [ ] `New()` in `fpdf.go` initializes `fontLoader` with error
- [ ] `New()` in `fpdf.go` initializes `fontCache` slice
- [ ] `New()` in `fpdf.go` handles fontLoader in options
- [ ] `New()` in `fpdf.go` removes coreFonts initialization
- [ ] `getCachedFont()` helper function added to `fpdf.go`
- [ ] `addFontToCache()` helper function added to `fpdf.go`
- [ ] `font.go` updated to use `fontLoader` instead of `coreFontReader`
- [ ] `font.go` updated to use slice-based caching
- [ ] All core font special cases removed
- [ ] `env.back.go` implements fontLoader with `os.ReadFile`
- [ ] `env.front.go` implements fontLoader with fetch (no local caching)
- [ ] `tinypdf.go` passes fontLoader to fpdf.New()
- [ ] `NewDocPdfTest()` provides test fontLoader
- [ ] No references to `embFS` remain
- [ ] No references to `coreFonts` remain
- [ ] No references to JSON font format remain
- [ ] No references to `.z` font files remain
- [ ] All tests pass with TTF fonts
- [ ] Font caching works using slice (TinyGo compatible)
- [ ] Error handling works when fontLoader not configured

## End of Document

Total lines: ~490
