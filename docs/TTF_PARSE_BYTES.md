# TrueType Font (TTF) Parsing from Bytes - Technical Documentation

## Overview

This document describes how the `tinypdf/fpdf` library parses TrueType font files from `[]byte` data. The parser extracts font metrics and glyph mappings needed for PDF generation without external dependencies.

## File Location

**Parser Implementation:** `/fpdf/ttfparser.go`

## Architecture

### Main Components

1. **`TtfType` struct** - Stores extracted font metrics
2. **`ttfParser` struct** - Internal parser with state
3. **`TtfParse()` function** - Public API entry point

## Data Structures

### TtfType - Font Metrics Container

```go
type TtfType struct {
    Embeddable             bool              // Can this font be embedded in PDF?
    UnitsPerEm             uint16            // Font design units per em
    PostScriptName         string            // PostScript font name
    Bold                   bool              // Is this a bold font?
    ItalicAngle            int16             // Italic angle (0 for upright)
    IsFixedPitch           bool              // Is this a monospaced font?
    TypoAscender           int16             // Typographic ascender
    TypoDescender          int16             // Typographic descender
    UnderlinePosition      int16             // Underline position
    UnderlineThickness     int16             // Underline thickness
    Xmin, Ymin, Xmax, Ymax int16             // Font bounding box
    CapHeight              int16             // Height of capital letters
    Widths                 []uint16          // Glyph widths array
    Chars                  map[uint16]uint16 // Unicode to glyph ID mapping
}
```

### ttfParser - Parser State

```go
type ttfParser struct {
    rec              TtfType           // Collected font data
    file             io.ReadSeeker     // Byte stream reader
    tables           map[string]uint32 // TTF table offsets
    numberOfHMetrics uint16            // Number of horizontal metrics
    numGlyphs        uint16            // Total number of glyphs
}
```

## Public API

### TtfParse - Parse TTF Font from File

```go
func TtfParse(fileStr string, readFile func(string) ([]byte, error)) (TtfRec TtfType, err error)
```

**Parameters:**
- `fileStr` - Path to TTF font file
- `readFile` - Function to read file into `[]byte` (e.g., `os.ReadFile`)

**Returns:**
- `TtfRec` - Parsed font metrics
- `err` - Error if parsing fails

**Usage Example:**

```go
ttf, err := fpdf.TtfParse("fonts/Arial.ttf", os.ReadFile)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Font: %s\n", ttf.PostScriptName)
fmt.Printf("Units per Em: %d\n", ttf.UnitsPerEm)
```

## Parsing Process Flow

### 1. Entry Point - TtfParse()

```go
func TtfParse(fileStr string, readFile func(string) ([]byte, error)) (TtfRec TtfType, err error) {
    // Step 1: Read entire font file into memory
    data, err := readFile(fileStr)
    if err != nil {
        return
    }

    // Step 2: Create parser with bytes.Reader for seeking
    var t ttfParser
    t.file = bytes.NewReader(data)

    // Step 3: Validate TTF format
    version, err := t.ReadStr(4)
    if version == "OTTO" {
        return // PostScript outlines not supported
    }
    if version != "\x00\x01\x00\x00" {
        return // Invalid TTF format
    }

    // Step 4: Parse table directory
    numTables := int(t.ReadUShort())
    t.Skip(3 * 2) // Skip searchRange, entrySelector, rangeShift
    t.tables = make(map[string]uint32)
    
    for j := 0; j < numTables; j++ {
        tag, _ := t.ReadStr(4)      // Table name (e.g., "head", "cmap")
        t.Skip(4)                    // checkSum
        offset := t.ReadULong()      // Table offset in file
        t.Skip(4)                    // length
        t.tables[tag] = offset       // Store table location
    }

    // Step 5: Parse all font tables
    err = t.ParseComponents()
    
    // Step 6: Return collected metrics
    TtfRec = t.rec
    return
}
```

**Key Points:**
- Entire font loaded into memory as `[]byte`
- Uses `bytes.NewReader` for random access (seeking)
- Validates TTF magic number `\x00\x01\x00\x00`
- Rejects PostScript-based fonts (OTTO format)
- Builds table directory for quick access to font tables

### 2. Parse All Components - ParseComponents()

```go
func (t *ttfParser) ParseComponents() (err error) {
    err = t.ParseHead()    // Font header (bounds, units)
    if err == nil {
        err = t.ParseHhea() // Horizontal header (metrics count)
        if err == nil {
            err = t.ParseMaxp() // Maximum profile (glyph count)
            if err == nil {
                err = t.ParseHmtx() // Horizontal metrics (widths)
                if err == nil {
                    err = t.ParseCmap() // Character to glyph mapping
                    if err == nil {
                        err = t.ParseName() // Font naming table
                        if err == nil {
                            err = t.ParseOS2() // OS/2 metrics
                            if err == nil {
                                err = t.ParsePost() // PostScript info
                            }
                        }
                    }
                }
            }
        }
    }
    return
}
```

**Parsing Order:**
1. `head` - Font header (must be first)
2. `hhea` - Horizontal header (needed for hmtx)
3. `maxp` - Maximum profile (needed for hmtx)
4. `hmtx` - Horizontal metrics (glyph widths)
5. `cmap` - Character mapping (Unicode → Glyph ID)
6. `name` - Font names (PostScript name)
7. `OS/2` - Extended metrics (embedding permissions)
8. `post` - PostScript information (italic angle, fixed pitch)

**Note:** Order matters! Some tables depend on data from previous tables.

## TTF Table Parsing Details

### head Table - Font Header

**Purpose:** Basic font metrics and bounding box

```go
func (t *ttfParser) ParseHead() (err error) {
    err = t.Seek("head")
    t.Skip(3 * 4)                   // version, fontRevision, checkSumAdjustment
    magicNumber := t.ReadULong()
    if magicNumber != 0x5F0F3CF5 {  // Validate magic number
        err = Errf("incorrect magic number")
        return
    }
    t.Skip(2)                       // flags
    t.rec.UnitsPerEm = t.ReadUShort()
    t.Skip(2 * 8)                   // created, modified timestamps
    t.rec.Xmin = t.ReadShort()
    t.rec.Ymin = t.ReadShort()
    t.rec.Xmax = t.ReadShort()
    t.rec.Ymax = t.ReadShort()
    return
}
```

**Extracted Data:**
- `UnitsPerEm` - Font coordinate system scale (typically 1000 or 2048)
- `Xmin, Ymin, Xmax, Ymax` - Font bounding box in font units

**Validation:**
- Magic number must be `0x5F0F3CF5`

### hhea Table - Horizontal Header

**Purpose:** Horizontal metrics metadata

```go
func (t *ttfParser) ParseHhea() (err error) {
    err = t.Seek("hhea")
    if err == nil {
        t.Skip(4 + 15*2)                      // version + 15 metrics fields
        t.numberOfHMetrics = t.ReadUShort()   // Count of glyph width entries
    }
    return
}
```

**Extracted Data:**
- `numberOfHMetrics` - Used by `hmtx` parser to read correct number of widths

### maxp Table - Maximum Profile

**Purpose:** Font maximums (glyph count)

```go
func (t *ttfParser) ParseMaxp() (err error) {
    err = t.Seek("maxp")
    if err == nil {
        t.Skip(4)                      // version
        t.numGlyphs = t.ReadUShort()   // Total glyph count
    }
    return
}
```

**Extracted Data:**
- `numGlyphs` - Total number of glyphs in font

### hmtx Table - Horizontal Metrics

**Purpose:** Glyph width data

```go
func (t *ttfParser) ParseHmtx() (err error) {
    err = t.Seek("hmtx")
    if err == nil {
        t.rec.Widths = make([]uint16, 0, 8)
        
        // Read width for each metric
        for j := uint16(0); j < t.numberOfHMetrics; j++ {
            t.rec.Widths = append(t.rec.Widths, t.ReadUShort())
            t.Skip(2) // Skip left side bearing (lsb)
        }
        
        // Extend widths array for remaining glyphs (use last width)
        if t.numberOfHMetrics < t.numGlyphs {
            lastWidth := t.rec.Widths[t.numberOfHMetrics-1]
            for j := t.numberOfHMetrics; j < t.numGlyphs; j++ {
                t.rec.Widths = append(t.rec.Widths, lastWidth)
            }
        }
    }
    return
}
```

**Extracted Data:**
- `Widths` - Array of glyph widths (one per glyph ID)

**Logic:**
- First `numberOfHMetrics` glyphs have explicit widths
- Remaining glyphs use the last width (optimization)

### cmap Table - Character to Glyph Mapping

**Purpose:** Map Unicode code points to glyph IDs

**Most Complex Table** - Uses format 4 (segment mapping)

```go
func (t *ttfParser) ParseCmap() (err error) {
    // Step 1: Find cmap table
    if err = t.Seek("cmap"); err != nil {
        return
    }
    
    t.Skip(2) // version
    numTables := int(t.ReadUShort())
    
    // Step 2: Find Unicode subtable (platform 3, encoding 1)
    offset31 := int64(0)
    for j := 0; j < numTables; j++ {
        platformID := t.ReadUShort()
        encodingID := t.ReadUShort()
        offset := int64(t.ReadULong())
        if platformID == 3 && encodingID == 1 { // Windows Unicode BMP
            offset31 = offset
        }
    }
    if offset31 == 0 {
        err = Errf("no Unicode encoding found")
        return
    }
    
    // Step 3: Parse format 4 subtable
    startCount := make([]uint16, 0, 8)
    endCount := make([]uint16, 0, 8)
    idDelta := make([]int16, 0, 8)
    idRangeOffset := make([]uint16, 0, 8)
    t.rec.Chars = make(map[uint16]uint16)
    
    t.file.Seek(int64(t.tables["cmap"])+offset31, io.SeekStart)
    format := t.ReadUShort()
    if format != 4 {
        err = Errf("unexpected subtable format: %d", format)
        return
    }
    
    t.Skip(2 * 2) // length, language
    segCount := int(t.ReadUShort() / 2)
    t.Skip(3 * 2) // searchRange, entrySelector, rangeShift
    
    // Step 4: Read segment arrays
    for j := 0; j < segCount; j++ {
        endCount = append(endCount, t.ReadUShort())
    }
    t.Skip(2) // reservedPad
    for j := 0; j < segCount; j++ {
        startCount = append(startCount, t.ReadUShort())
    }
    for j := 0; j < segCount; j++ {
        idDelta = append(idDelta, t.ReadShort())
    }
    offset, _ := t.file.Seek(int64(0), io.SeekCurrent)
    for j := 0; j < segCount; j++ {
        idRangeOffset = append(idRangeOffset, t.ReadUShort())
    }
    
    // Step 5: Build Unicode → Glyph ID map
    for j := 0; j < segCount; j++ {
        c1 := startCount[j]
        c2 := endCount[j]
        d := idDelta[j]
        ro := idRangeOffset[j]
        
        if ro > 0 {
            t.file.Seek(offset+2*int64(j)+int64(ro), io.SeekStart)
        }
        
        for c := c1; c <= c2; c++ {
            if c == 0xFFFF {
                break
            }
            var gid int32
            if ro > 0 {
                gid = int32(t.ReadUShort())
                if gid > 0 {
                    gid += int32(d)
                }
            } else {
                gid = int32(c) + int32(d)
            }
            if gid >= 65536 {
                gid -= 65536
            }
            if gid > 0 {
                t.rec.Chars[c] = uint16(gid)
            }
        }
    }
    return
}
```

**Extracted Data:**
- `Chars` - Map of `Unicode → Glyph ID` (e.g., `'A' → 36`)

**Format 4 Logic:**
- Divides Unicode space into segments
- Each segment has start/end code points
- Uses delta or direct lookup for glyph ID calculation
- Handles missing glyphs (gid = 0)

**Example:**
- Segment: start=0x0041 ('A'), end=0x005A ('Z'), delta=5
- 'A' (0x0041) → Glyph ID = 0x0041 + 5 = 70

### name Table - Font Naming

**Purpose:** Extract PostScript font name

```go
func (t *ttfParser) ParseName() (err error) {
    err = t.Seek("name")
    if err == nil {
        tableOffset, _ := t.file.Seek(0, io.SeekCurrent)
        t.rec.PostScriptName = ""
        t.Skip(2) // format
        count := t.ReadUShort()
        stringOffset := t.ReadUShort()
        
        for j := uint16(0); j < count && t.rec.PostScriptName == ""; j++ {
            t.Skip(3 * 2) // platformID, encodingID, languageID
            nameID := t.ReadUShort()
            length := t.ReadUShort()
            offset := t.ReadUShort()
            
            if nameID == 6 { // PostScript name
                t.file.Seek(int64(tableOffset)+int64(stringOffset)+int64(offset), io.SeekStart)
                s, _ := t.ReadStr(int(length))
                s = Convert(s).Replace("\x00", "", -1).String()
                
                // Remove invalid PostScript characters
                re, _ := regexp.Compile(`[(){}<> /%[\]]`)
                t.rec.PostScriptName = re.ReplaceAllString(s, "")
            }
        }
        if t.rec.PostScriptName == "" {
            err = Errf("the name PostScript was not found")
        }
    }
    return
}
```

**Extracted Data:**
- `PostScriptName` - Font name for PDF (e.g., "Arial-BoldMT")

**Name ID 6:** PostScript name (required for PDF)

### OS/2 Table - OS/2 and Windows Metrics

**Purpose:** Extended font metrics and embedding permissions

```go
func (t *ttfParser) ParseOS2() (err error) {
    err = t.Seek("OS/2")
    if err == nil {
        version := t.ReadUShort()
        t.Skip(3 * 2) // xAvgCharWidth, usWeightClass, usWidthClass
        fsType := t.ReadUShort()
        
        // Check embedding permissions
        t.rec.Embeddable = (fsType != 2) && (fsType&0x200) == 0
        
        t.Skip(11*2 + 10 + 4*4 + 4)
        fsSelection := t.ReadUShort()
        t.rec.Bold = (fsSelection & 32) != 0
        t.Skip(2 * 2) // usFirstCharIndex, usLastCharIndex
        t.rec.TypoAscender = t.ReadShort()
        t.rec.TypoDescender = t.ReadShort()
        
        if version >= 2 {
            t.Skip(3*2 + 2*4 + 2)
            t.rec.CapHeight = t.ReadShort()
        } else {
            t.rec.CapHeight = 0
        }
    }
    return
}
```

**Extracted Data:**
- `Embeddable` - Can font be embedded in PDF?
- `Bold` - Is this a bold font?
- `TypoAscender` - Typographic ascender
- `TypoDescender` - Typographic descender  
- `CapHeight` - Height of capital letters (version 2+)

**fsType Flags:**
- Bit 1: Restricted embedding
- Bit 9 (0x200): Bitmap embedding only

### post Table - PostScript Information

**Purpose:** PostScript-specific metrics

```go
func (t *ttfParser) ParsePost() (err error) {
    err = t.Seek("post")
    if err == nil {
        t.Skip(4) // version
        t.rec.ItalicAngle = t.ReadShort()
        t.Skip(2) // Skip decimal part of italic angle
        t.rec.UnderlinePosition = t.ReadShort()
        t.rec.UnderlineThickness = t.ReadShort()
        t.rec.IsFixedPitch = t.ReadULong() != 0
    }
    return
}
```

**Extracted Data:**
- `ItalicAngle` - Italic angle in degrees (0 = upright)
- `UnderlinePosition` - Underline position
- `UnderlineThickness` - Underline thickness
- `IsFixedPitch` - Is this a monospaced font?

## Helper Functions - Binary Reading

### Seek to Table

```go
func (t *ttfParser) Seek(tag string) (err error) {
    ofs, ok := t.tables[tag]
    if !ok {
        return Errf("table not found: %s", tag)
    }
    _, err = t.file.Seek(int64(ofs), io.SeekStart)
    return
}
```

### Skip Bytes

```go
func (t *ttfParser) Skip(n int) {
    t.file.Seek(int64(n), io.SeekCurrent)
}
```

### Read String

```go
func (t *ttfParser) ReadStr(length int) (str string, err error) {
    buf := make([]byte, length)
    n, err := t.file.Read(buf)
    if n == length {
        str = string(buf)
    }
    return
}
```

### Read Unsigned Short (uint16)

```go
func (t *ttfParser) ReadUShort() (val uint16) {
    binary.Read(t.file, binary.BigEndian, &val)
    return
}
```

### Read Signed Short (int16)

```go
func (t *ttfParser) ReadShort() (val int16) {
    binary.Read(t.file, binary.BigEndian, &val)
    return
}
```

### Read Unsigned Long (uint32)

```go
func (t *ttfParser) ReadULong() (val uint32) {
    binary.Read(t.file, binary.BigEndian, &val)
    return
}
```

**Important:** All multi-byte integers in TTF files are **big-endian**.

## Usage for Bytes → TTF Refactoring

### Current Flow (File-based)

```go
// Current: Read from filesystem
ttf, err := fpdf.TtfParse("fonts/Arial.ttf", os.ReadFile)
```

### Proposed Flow (Bytes-based)

To adapt this for `[]byte` input from `fontLoader`:

```go
// Option 1: Modify TtfParse to accept []byte directly
func TtfParseBytes(fontData []byte) (TtfRec TtfType, err error) {
    var t ttfParser
    t.file = bytes.NewReader(fontData) // Create reader from bytes
    
    // Validate format
    version, err := t.ReadStr(4)
    // ... rest of parsing logic
}

// Usage:
fontData, err := fontLoader("fonts/Arial.ttf")
if err != nil {
    return err
}
ttf, err := fpdf.TtfParseBytes(fontData)
```

```go
// Option 2: Wrap existing TtfParse with in-memory reader
func (f *Fpdf) loadTTFFromBytes(fontData []byte) (TtfType, error) {
    // Create temporary in-memory "file"
    readFunc := func(path string) ([]byte, error) {
        return fontData, nil // Return pre-loaded bytes
    }
    return TtfParse("", readFunc)
}
```

**Recommendation:** Use **Option 1** - cleaner API, no fake file path needed.

## Memory Considerations

### Current Implementation

- **Loads entire font into memory** via `readFile(fileStr)`
- Uses `bytes.NewReader` for seeking (no additional memory copy)
- Font data stays in memory until parser completes

### For WASM/Frontend

- Font fetched once via `fetchgo`
- Cached in `Fpdf.fontCache` slice (raw `[]byte`)
- Parser creates temporary `bytes.NewReader` (no copy)
- After parsing, only `TtfType` metrics kept (small)

**Memory Profile:**
- Font file: ~50KB - 500KB (TTF)
- Parsed metrics: ~5KB - 50KB (depends on glyph count)
- Cache: Raw font bytes (shared across multiple uses)

## TinyGo Compatibility

### Compatible Features ✅

- `bytes.NewReader` - Supported
- `binary.BigEndian` - Supported
- `regexp.Compile` - Supported (with limitations)
- Slices and maps - Supported
- `io.ReadSeeker` interface - Supported

### Potential Issues ⚠️

- **Large allocations** - TinyGo has smaller heap
- **Regex complexity** - Keep patterns simple
- **Error handling** - Avoid complex error types

### Recommendations

1. **Pre-validate fonts** - Don't parse invalid fonts at runtime
2. **Limit font size** - Keep fonts under 200KB if possible
3. **Cache parsed metrics** - Avoid re-parsing same font
4. **Use font subsets** - Include only needed glyphs (reduces size)

## Error Handling

### Common Errors

1. **"fonts based on PostScript outlines are not supported"**
   - Font uses CFF/PostScript curves, not TrueType
   - Solution: Convert to TrueType format

2. **"unrecognized file format"**
   - Invalid magic number (not `\x00\x01\x00\x00`)
   - Solution: Verify file is valid TTF

3. **"incorrect magic number"**
   - Invalid `head` table magic (not `0x5F0F3CF5`)
   - Solution: File corrupted or not a TTF

4. **"no Unicode encoding found"**
   - Missing platform 3, encoding 1 in `cmap`
   - Solution: Use a font with Unicode support

5. **"table not found: <name>"**
   - Required table missing from font
   - Solution: Use a complete TrueType font

6. **"the name PostScript was not found"**
   - Missing name ID 6 in `name` table
   - Solution: Font must have PostScript name

### Error Prevention

```go
// Validate font before parsing
func validateTTF(data []byte) error {
    if len(data) < 12 {
        return errors.New("file too small to be TTF")
    }
    if !bytes.Equal(data[0:4], []byte{0x00, 0x01, 0x00, 0x00}) {
        return errors.New("not a TrueType font")
    }
    return nil
}
```

## Performance

### Parse Time

- **Small font** (100 glyphs): ~1-2ms
- **Medium font** (1000 glyphs): ~5-10ms  
- **Large font** (10000 glyphs): ~20-50ms

### Optimization Tips

1. **Cache parsed fonts** - Don't parse same font multiple times
2. **Use font subsets** - Reduce glyph count = faster parsing
3. **Lazy parse** - Only parse tables when needed
4. **Pre-load fonts** - Parse during initialization, not per-request

## TTF Format References

### Official Specifications

- **Microsoft Typography:** https://docs.microsoft.com/en-us/typography/opentype/spec/
- **Apple TrueType Reference:** https://developer.apple.com/fonts/TrueType-Reference-Manual/

### Key Tables Documentation

- **head:** https://docs.microsoft.com/en-us/typography/opentype/spec/head
- **hhea:** https://docs.microsoft.com/en-us/typography/opentype/spec/hhea
- **maxp:** https://docs.microsoft.com/en-us/typography/opentype/spec/maxp
- **hmtx:** https://docs.microsoft.com/en-us/typography/opentype/spec/hmtx
- **cmap:** https://docs.microsoft.com/en-us/typography/opentype/spec/cmap
- **name:** https://docs.microsoft.com/en-us/typography/opentype/spec/name
- **OS/2:** https://docs.microsoft.com/en-us/typography/opentype/spec/os2
- **post:** https://docs.microsoft.com/en-us/typography/opentype/spec/post

## Summary

### Key Takeaways

1. **Parser accepts `[]byte`** via `readFile` function parameter
2. **Uses `bytes.NewReader`** for random access (seeking)
3. **Parses 8 TTF tables** in specific order
4. **Extracts font metrics** (widths, bounds, names)
5. **Builds Unicode → Glyph ID mapping** from `cmap`
6. **Validates format** (magic numbers, table presence)
7. **TinyGo compatible** (uses standard library only)

### Integration with fontLoader Refactoring

**Current:**
```go
// File-based loading
ttf, err := TtfParse("fonts/Arial.ttf", os.ReadFile)
```

**After refactoring:**
```go
// Bytes-based loading
fontData, err := f.fontLoader("fonts/Arial.ttf")
ttf, err := TtfParseBytes(fontData)
```

**Changes needed:**
1. Create `TtfParseBytes([]byte)` function (wrapper or new implementation)
2. Integrate with `fontLoader` function in `Fpdf` struct
3. Cache raw font bytes in `fontCache` slice
4. Parse once, reuse metrics for multiple PDF pages

---

**Document Version:** 1.0  
**Last Updated:** 2025-01-04  
**Related:** [REMOVE_EMBED_FONTS.md](issues/REMOVE_EMBED_FONTS.md)
