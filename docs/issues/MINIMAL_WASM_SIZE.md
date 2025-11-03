# WebAssembly Binary Size Analysis - TinyPDF

**Date:** November 3, 2025  
**Target:** `src/cmd/webclient/main.go`  
**Tool:** twiggy v0.8.0  
**Goal:** Minimize WebAssembly binary size

## Build Sizes

| Compiler | Size | Optimization | Notes |
|----------|------|--------------|-------|
| TinyGo | 1.1 MB | -opt=z | Production build (4.3x smaller) |
| Go std | 4.8 MB | default | Analysis build |

**Conclusion:** TinyGo is significantly better for WASM. Analysis performed on Go build for better tooling support.


## Top Size Contributors (Twiggy Analysis)

### Largest Functions/Data

| Size | % | Item | Category |
|------|---|------|----------|
| 252 KB | 5.11% | data[48337] | Static data |
| 131 KB | 2.66% | data[48335] | Static data |
| 93 KB | 1.88% | Function names subsection | Debug info |
| 75 KB | 1.52% | data[2796] | Static data |
| 41 KB | 0.84% | `GenerateCutFont` | Font processing |
| 26 KB | 0.53% | `putfonts` | PDF font output |
| 25 KB | 0.51% | `encoding/json.literalStore` | JSON decoder |
| 25 KB | 0.51% | `time.parse` | Time parsing |
| 25 KB | 0.50% | `fpdf.New` | PDF initialization |
| 25 KB | 0.50% | `fmt.printValue` | Reflection printing |
| 24 KB | 0.49% | `runtime.initMetrics` | Go runtime |
| 22 KB | 0.45% | `image/png.readImagePass` | **PNG decoder** |
| 17 KB | 0.35% | `time.Time.appendFormat` | Time formatting |
| 17 KB | 0.35% | `fpdf.CellFormat` | PDF cell rendering |
| 14 KB | 0.30% | `tinystring.formatValue` | String formatting |
| 14 KB | 0.27% | `encoding/json.typeFields` | JSON reflection |
| 12 KB | 0.25% | `fpdf.generateCIDFontMap` | Font mapping |
| 9 KB | 0.19% | `image/jpeg.processSOS` | **JPEG decoder** |
| 9 KB | 0.18% | `image/gif.readImageDescriptor` | **GIF decoder** |
| 6 KB | 0.13% | `fpdf.putimage` | **Image output** |
| 6 KB | 0.13% | `compress/flate.huffmanBlock` | Compression |
| 5 KB | 0.11% | `compress/flate.deflate` | Compression |

**Image codecs total in top 200:** ~211 KB (4.3%)


## Key Findings

### 1. Static Data Dominates (>50%)
Large data arrays suggest embedded fonts, Unicode tables, or other static resources.

### 2. Image Codecs Present (211+ KB measured)
All three image formats included:
- `image/png`: 22 KB+ in decoder alone
- `image/jpeg`: 9 KB+ in decoder 
- `image/gif`: 9 KB+ in decoder
- Plus `fpdf.putimage` and related: 6 KB

**Finding:** If `GenerateSamplePDF()` doesn't use images, these are wasted bytes.

### 3. JSON Encoding Heavy (60+ KB)
Multiple JSON functions present despite minimal JSON usage expected in PDF generation.

### 4. Font Processing Large (53+ KB)
- `GenerateCutFont`: 41 KB
- `generateCIDFontMap`: 12 KB

Likely needed for UTF-8 font support.

### 5. fpdf Package Imports
```
image, image/color, image/gif, image/jpeg, image/png
crypto/md5, crypto/rc4, crypto/sha1
encoding/json, encoding/xml, encoding/gob
compress/zlib, regexp
```
Total: 31 standard library packages

## Optimization Recommendations


## Optimization Recommendations

### Priority 1: Remove Image Support (Est. 200-300 KB savings)

**Action:**
```go
// Create fpdf/fpdf_nowasm.go with build tag
//go:build !wasm

// Keep image imports

// Create fpdf/fpdf_wasm.go
//go:build wasm

// Remove: image/png, image/jpeg, image/gif imports
// Stub out: Fpdf.Image(), Fpdf.putimage()
```

**Impact:** 200-300 KB reduction (18-27%)

### Priority 2: Remove Unused Encodings (Est. 80-120 KB savings)

Check if actually used:
- `encoding/gob` - likely unused for web
- `encoding/xml` - possibly unused  
- Crypto packages if no PDF encryption needed

**Impact:** 80-120 KB reduction (7-11%)

### Priority 3: Optimize Static Data

The largest items are static data arrays. Investigate:
- Are embedded fonts necessary for web?
- Can fonts be loaded dynamically?
- Are Unicode tables fully needed?

**Impact:** Potentially 100-200 KB (9-18%)

### Priority 4: Strip Debug Info

Function names subsection: 93 KB. Build with:
```bash
tinygo build -no-debug -opt=z ...
```

Already applied, but verify complete removal.

## Implementation Strategy

### Phase 1: Quick Wins (Target: 400-600 KB)

1. Add `//go:build !wasm` tags to image-related code in fpdf
2. Create WASM stubs for image functions
3. Test that `GenerateSamplePDF()` works without images

### Phase 2: Feature Flags (Target: 300-500 KB)

```go
//go:build wasm && !pdfimages
// Minimal build

//go:build wasm && pdfimages  
// Full features
```

### Phase 3: Dynamic Loading

For rarely-used features:
- Load fonts on demand
- Stream instead of embed

## Testing Commands

```bash
# Build minimal
tinygo build -tags wasm,noimages -o main.wasm -target wasm -opt=z -no-debug src/cmd/webclient/main.go

# Analyze
twiggy top -n 50 main.wasm
twiggy top -n 200 main.wasm | grep image

# Size comparison
ls -lh main*.wasm
```

## Expected Results

| Build | Current | Target | Reduction |
|-------|---------|--------|-----------|
| Baseline | 1.1 MB | - | - |
| -images | - | 800-900 KB | 200-300 KB (18-27%) |
| -images -json | - | 700-800 KB | 300-400 KB (27-36%) |
| Minimal | - | 400-600 KB | 500-700 KB (45-64%) |

## Next Steps

1. **Verify image usage:**
   ```bash
   grep -r "\.Image\|ImageOptions\|putimage" src/cmd/webclient/
   grep -r "\.Image\|ImageOptions" pdfgen.go
   ```

2. **Create build tags:**
   - `fpdf/image.go` → `fpdf/image_full.go` with `//go:build !wasm`
   - `fpdf/image_stub.go` with `//go:build wasm`

3. **Test:**
   - Build and verify functionality
   - Measure actual size reduction
   - Document any breaking changes
