# Minimal Image Support Plan for TinyGo/WASM

**Date:** November 3, 2025  
**Related:** [MINIMAL_WASM_SIZE.md](MINIMAL_WASM_SIZE.md#priority-1-remove-image-support-est-200-300-kb-savings)  
**Goal:** Reduce WASM binary by 200-300 KB by refactoring image support  
**Status:** Planning

## Current Situation

### Size Impact (from twiggy analysis)
- `image/png.readImagePass`: 22 KB
- `image/jpeg.processSOS`: 9 KB  
- `image/gif.readImageDescriptor`: 9 KB
- `fpdf.putimage`: 6 KB
- **Total measured:** ~211 KB in top 200 items
- **Estimated full impact:** 200-300 KB

### Current fpdf Image Dependencies
```go
// Backend only - REMOVE from WASM:
import (
    "image"
    "image/color"
    "image/gif"
    "image/jpeg"
    "image/png"
)

// WASM replacement:
import (
    "syscall/js"  // For browser API access
    // NO encoding/base64 - use Blob directly!
    // NO fmt - use tinystring.Fmt
)
```

### Image Functions in fpdf
- `Image()` - Main image insertion
- `ImageOptions()` - Image with options
- `RegisterImage()` - Pre-register image
- `RegisterImageReader()` - Register from reader
- `RegisterImageOptionsReader()` - Register with options
- `putimage()` - Internal: write image to PDF
- `parsepng()`, `parsejpg()`, `parsegif()` - Format parsers

## Strategy: Environment-Based Image Support

Following the pattern in `env.front.go` / `env.back.go`, implement dual image support:

### Backend (Native Go)
- Keep full image codec support
- Use standard library decoders
- Direct file system access

### Frontend (WASM/Browser)
- **Remove Go image codecs** (saves 200-300 KB)
- Use browser Canvas API for image processing
- Leverage native browser image decoding
- **Work with Blob/Uint8Array directly** - NO base64 encoding!
  - Base64 adds 33% size overhead
  - Blob URL is native and efficient
  - Avoids `encoding/base64` package inclusion

### Key Dependencies to Avoid in WASM
- ❌ `encoding/base64` - Use Blob instead
- ❌ `fmt,errors,strconv,strings` - Use `tinystring` instead
- ❌ `image/*` - Use Canvas API instead

## Implementation Plan

### Phase 1: Refactor Image Code Structure

#### 1.1 Create Image Interface
```go
// File: fpdf/image_interface.go
package fpdf

type ImageProcessor interface {
    // ParseImage decodes image data and returns dimensions and pixel data
    ParseImage(data []byte, format string) (*ImageInfoType, error)
    
    // GetImageInfo extracts metadata without full decode
    GetImageInfo(data []byte, format string) (*ImageInfoType, error)
}
```

#### 1.2 Backend Implementation
```go
// File: fpdf/image_native.go
//go:build !wasm
// +build !wasm

package fpdf

import (
    "image"
    "image/gif"
    "image/jpeg"
    "image/png"
)

type NativeImageProcessor struct{}

func (p *NativeImageProcessor) ParseImage(data []byte, format string) (*ImageInfoType, error) {
    // Current implementation using image/png, image/jpeg, image/gif
    // Keep existing parsepng(), parsejpg(), parsegif() logic
}
```

#### 1.3 Browser Implementation
```go
// File: fpdf/image_browser.go
//go:build wasm
// +build wasm

package fpdf

import (
    "syscall/js"
)

type BrowserImageProcessor struct{}

func (p *BrowserImageProcessor) ParseImage(data []byte, format string) (*ImageInfoType, error) {
    // Use browser Canvas API to decode image
    return p.decodeWithCanvas(data, format)
}

func (p *BrowserImageProcessor) decodeWithCanvas(data []byte, format string) (*ImageInfoType, error) {
    // 1. Create Image element in browser
    // 2. Load data as data URL or Blob
    // 3. Wait for image to load
    // 4. Create canvas with image dimensions
    // 5. Draw image to canvas
    // 6. Extract pixel data with getImageData()
    // 7. Convert to PDF-compatible format
}
```

### Phase 2: Browser Canvas API Integration

#### 2.1 Image Decoding Flow (WASM)
```javascript
// Browser-side pseudo-code
async function decodeImageFromBlob(blobOrBytes, format) {
    // Create Blob from bytes if needed
    const blob = (blobOrBytes instanceof Blob) 
        ? blobOrBytes 
        : new Blob([blobOrBytes], { type: getMimeType(format) });
    
    // Create Object URL (no base64 encoding needed!)
    const objectURL = URL.createObjectURL(blob);
    
    // Create image element
    const img = new Image();
    img.src = objectURL;
    
    // Wait for load
    await new Promise((resolve, reject) => {
        img.onload = resolve;
        img.onerror = reject;
    });
    
    // Create canvas
    const canvas = document.createElement('canvas');
    canvas.width = img.width;
    canvas.height = img.height;
    
    // Draw and extract pixels
    const ctx = canvas.getContext('2d');
    ctx.drawImage(img, 0, 0);
    const imageData = ctx.getImageData(0, 0, img.width, img.height);
    
    // Clean up object URL
    URL.revokeObjectURL(objectURL);
    
    return {
        width: img.width,
        height: img.height,
        data: imageData.data,  // RGBA pixels
        format: format
    };
}
```

#### 2.2 Go/WASM Implementation
```go
// File: fpdf/image_browser.go (detailed)
//go:build wasm

package fpdf

import (
    "syscall/js"
    
    . "github.com/tinywasm/fmt"
)

type BrowserImageProcessor struct {
    document js.Value
}

func NewBrowserImageProcessor() *BrowserImageProcessor {
    return &BrowserImageProcessor{
        document: js.Global().Get("document"),
    }
}

func (p *BrowserImageProcessor) ParseImage(data []byte, format string) (*ImageInfoType, error) {
    // Create Uint8Array from Go bytes
    uint8Array := js.Global().Get("Uint8Array").New(len(data))
    js.CopyBytesToJS(uint8Array, data)
    
    // Create Blob directly from bytes (no base64!)
    blobParts := []interface{}{uint8Array}
    blobOptions := map[string]interface{}{
        "type": p.getMimeType(format),
    }
    blob := js.Global().Get("Blob").New(blobParts, blobOptions)
    
    // Create object URL from Blob
    objectURL := js.Global().Get("URL").Call("createObjectURL", blob)
    defer js.Global().Get("URL").Call("revokeObjectURL", objectURL)
    
    // Create image element
    img := p.document.Call("createElement", "img")
    
    // Create channels for async loading
    loadDone := make(chan bool)
    loadError := make(chan error)
    
    // Set up load handler
    loadFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
        loadDone <- true
        return nil
    })
    defer loadFunc.Release()
    
    // Set up error handler
    errorFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
        loadError <- Err("failed to load image")
        return nil
    })
    defer errorFunc.Release()
    
    img.Call("addEventListener", "load", loadFunc)
    img.Call("addEventListener", "error", errorFunc)
    
    // Start loading from blob URL
    img.Set("src", objectURL)
    
    // Wait for load
    select {
    case <-loadDone:
        // Success, proceed to extract data
    case err := <-loadError:
        return nil, err
    }
    
    // Get dimensions
    width := img.Get("width").Int()
    height := img.Get("height").Int()
    
    // Create canvas
    canvas := p.document.Call("createElement", "canvas")
    canvas.Set("width", width)
    canvas.Set("height", height)
    
    // Get context and draw
    ctx := canvas.Call("getContext", "2d")
    ctx.Call("drawImage", img, 0, 0)
    
    // Extract pixel data
    imageData := ctx.Call("getImageData", 0, 0, width, height)
    pixelData := imageData.Get("data")
    
    // Convert Uint8ClampedArray to Go byte slice
    pixels := make([]byte, pixelData.Get("length").Int())
    js.CopyBytesToGo(pixels, pixelData)
    
    // Build ImageInfoType
    info := &ImageInfoType{
        w:    float64(width),
        h:    float64(height),
        cs:   "DeviceRGB",
        bpc:  8,
        f:    "FlateDecode",
        data: p.rgbaToRGB(pixels),  // Convert RGBA to RGB
    }
    
    return info, nil
}

func (p *BrowserImageProcessor) rgbaToRGB(rgba []byte) []byte {
    // Convert RGBA (4 bytes/pixel) to RGB (3 bytes/pixel)
    rgbLen := (len(rgba) / 4) * 3
    rgb := make([]byte, rgbLen)
    
    j := 0
    for i := 0; i < len(rgba); i += 4 {
        rgb[j] = rgba[i]     // R
        rgb[j+1] = rgba[i+1] // G
        rgb[j+2] = rgba[i+2] // B
        // Skip alpha channel
        j += 3
    }
    
    return rgb
}

func (p *BrowserImageProcessor) getMimeType(format string) string {
    switch format {
    case "png", "PNG":
        return "image/png"
    case "jpg", "jpeg", "JPG", "JPEG":
        return "image/jpeg"
    case "gif", "GIF":
        return "image/gif"
    default:
        return "image/png"
    }
}
```

### Phase 3: Integrate into Fpdf

#### 3.1 Modify Fpdf Structure
```go
// File: fpdf/fpdf.go
type Fpdf struct {
    // ... existing fields ...
    
    imageProcessor ImageProcessor  // Add this field
}
```

#### 3.2 Initialize Processor
```go
// File: fpdf/fpdf_init.go or in New()

func (f *Fpdf) initImageProcessor() {
    #ifdef wasm
        f.imageProcessor = NewBrowserImageProcessor()
    #else
        f.imageProcessor = NewNativeImageProcessor()
    #endif
}
```

#### 3.3 Update Image Functions
```go
// File: fpdf/fpdf.go

func (f *Fpdf) RegisterImageReader(imgName, tp string, r io.Reader) (info *ImageInfoType) {
    // Read data
    data, err := io.ReadAll(r)
    if err != nil {
        f.err = err
        return nil
    }
    
    // Use processor instead of direct codec
    info, err = f.imageProcessor.ParseImage(data, tp)
    if err != nil {
        f.err = err
        return nil
    }
    
    // Store and register as before
    f.images[imgName] = info
    return info
}
```

### Phase 4: Stub Out Unused Functions

#### 4.1 Optional: Remove Gob Encoding for Images (WASM only)
```go
// File: fpdf/def_gob_wasm.go
//go:build wasm

package fpdf

import . "github.com/tinywasm/fmt"

func (info *ImageInfoType) GobEncode() (buf []byte, err error) {
    return nil, Err("error", "gob encoding not supported in WASM")
}

func (info *ImageInfoType) GobDecode(buf []byte) (err error) {
    return Err("error", "gob decoding not supported in WASM")
}
```

### Phase 5: Testing Strategy

#### 5.1 Backend Tests (Unchanged)
```bash
# Run existing tests with native Go
go test ./fpdf -v -run Test_Image
go test ./fpdf -v -run Test_RegisterImage
```

#### 5.2 WASM Tests
```go
// File: fpdf/image_browser_test.go
//go:build wasm

func TestBrowserImageProcessor(t *testing.T) {
    // Create test image data (small PNG)
    // Test decoding via Canvas API
    // Verify dimensions and pixel data
}
```

#### 5.3 Integration Test
```bash
# Build WASM
tinygo build -tags wasm -o test.wasm -target wasm fpdf

# Verify image imports removed
twiggy top -n 200 test.wasm | grep image
# Should show no image/png, image/jpeg, image/gif
```

### Phase 6: Build Configuration

#### 6.1 Build Tags
```bash
# Minimal WASM (no native image codecs)
tinygo build -o main.wasm -target wasm -opt=z -no-debug \
    -tags wasm \
    src/cmd/webclient/main.go

# Native (full features)
go build -o main src/cmd/appserver/main.go
```

#### 6.2 Size Verification
```bash
# Before
ls -lh main_before.wasm  # ~1.1 MB

# After refactor
ls -lh main_after.wasm   # Target: ~800-900 KB

# Savings
echo "Reduction: ~200-300 KB"
```

## Alternative Approaches

### Option A: Compile-Time Image Exclusion (Simpler)
```go
// File: fpdf/image_stub.go
//go:build wasm

package fpdf

import "github.com/tinywasm/fmt"

func (f *Fpdf) Image(...) {
    f.err = tinystring.Fmt("error", "images not supported in WASM build")
}

func (f *Fpdf) RegisterImage(...) (*ImageInfoType, error) {
    return nil, tinystring.Fmt("error", "images not supported in WASM build")
}
```

**Pros:** Simple, minimal code changes  
**Cons:** No image support at all in WASM

### Option B: Hybrid (Recommended)
- WASM: Browser Canvas API for decoding
- Native: Standard library codecs
- Both: Same `ImageInfoType` output

**Pros:** Feature parity, maximum size savings  
**Cons:** More complex implementation

### Option C: Server-Side Rendering
- WASM: Send image data to server
- Server: Process image and return PDF-ready data
- WASM: Embed processed data

**Pros:** Zero client-side image processing  
**Cons:** Requires server, network latency

## Risk Assessment

### Technical Risks
1. **Canvas API async nature** - Needs promise/channel handling
2. **Browser compatibility** - Canvas 2D well supported, but test needed
3. **Memory usage** - Large images in canvas may cause issues
4. **Color space conversion** - RGBA→RGB needs validation

### Mitigation
- Implement timeout on image loading (5s max)
- Feature detection for Canvas API
- Size limits on images (e.g., max 4096x4096)
- Thorough color conversion testing

## Success Metrics

### Size Reduction
- **Target:** 200-300 KB reduction in WASM binary
- **Measure:** `ls -lh main.wasm` before/after
- **Verify:** `twiggy top -n 200 main.wasm | grep image`

### Functionality
- ✅ Backend: All image tests pass
- ✅ WASM: Can load and embed PNG/JPEG/GIF
- ✅ Output: PDFs with images render correctly
- ✅ Performance: <100ms for typical image processing

### Code Quality
- ✅ Clean separation: `image_interface.go`, `image_native.go`, `image_browser.go`
- ✅ No duplicate code
- ✅ Build tags properly applied
- ✅ Tests for both environments

## Timeline

| Phase | Duration | Deliverable |
|-------|----------|-------------|
| 1. Refactor structure | 2 days | Interface + native impl |
| 2. Browser Canvas API | 3 days | WASM image processor |
| 3. Integration | 1 day | Update Fpdf funcs |
| 4. Stub unused | 1 day | Remove gob, etc. |
| 5. Testing | 2 days | Full test suite |
| 6. Optimization | 1 day | Final size tweaks |
| **Total** | **10 days** | Production-ready |

## Next Steps

1. **Validation:**
   - [ ] Verify `GenerateSamplePDF()` uses images
   - [ ] Check if images are in actual use case
   - [ ] If not used, consider Option A (stub out)

2. **Prototype:**
   - [ ] Create `image_browser.go` with Canvas API
   - [ ] Test single image decode in WASM
   - [ ] Measure size impact

3. **Implementation:**
   - [ ] Follow Phase 1-6 if images needed
   - [ ] Document API changes if any
   - [ ] Update main README with build instructions

4. **Verification:**
   - [ ] Build both targets
   - [ ] Run test suite
   - [ ] Measure final binary size
   - [ ] Update MINIMAL_WASM_SIZE.md with results

---

**Status:** Ready for implementation decision  
**Decision needed:** Proceed with Option B (Browser Canvas API) or Option A (Stub out)?  
**Blockers:** None - all dependencies available
