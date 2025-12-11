# BUG: TinyGo Compilation Fails with Reflection Panic

## Error Summary

**Status**: CRITICAL - Blocks TinyGo/WASM compilation
**Date**: 2025-10-31
**Environment**: TinyGo WASM compilation
**Working**: Standard Go compiler
**Failing**: TinyGo compiler

## Error Message

```
panic: reflect: unimplemented: AssignableTo with interface
main.wasm:0x37238 Uncaught (in promise) RuntimeError: unreachable
    at main.runtime.panicOrGoexit (main.wasm:0x37238)
    at main.runtime._panic (main.wasm:0x3473)
    at main.(:4430/*internal/reflectlite.RawType).AssignableTo
    at main.interface:{...}.Implements$invoke (main.wasm:0xb4914)
    at main.encoding/gob.implementsInterface (main.wasm:0xb446b)
    at main.encoding/gob.userType (main.wasm:0xacded)
    at main.encoding/gob.mustGetTypeInfo (main.wasm:0xaa477)
    at main.encoding/gob.init (main.wasm:0x4cba9)
```

## Root Cause Analysis

### Primary Issue: `encoding/gob` Not Supported in TinyGo

The panic originates from `encoding/gob.init()` during WASM module initialization. TinyGo has **limited reflection support** and does NOT support:

1. **`reflect.AssignableTo()`** - Used by encoding/gob to check type compatibility
2. **`reflect.Implements()`** - Used by encoding/gob for interface checking
3. **Full reflection API** - TinyGo only supports basic `reflect.ValueOf()` and `Kind()`

### Secondary Issue: Recent tinystring Changes

The recent fix to support custom types in `tinystring.Fmt()` introduced `reflect` package usage:

**Files using reflection:**
- `tinystring/convert.go:4` - `import "reflect"`
- `tinystring/num_int.go:3` - `import "reflect"`
- `tinystring/num_float.go:3` - `import "reflect"`

**Reflection usage:**
- `reflect.ValueOf(value)` - Extract underlying value from custom types
- `reflect.Kind()` - Check if custom type wraps int/uint/float/string
- `rv.Int()`, `rv.Uint()`, `rv.Float()`, `rv.String()` - Extract primitive values

**TinyGo compatibility:**
✅ `reflect.ValueOf()` - SUPPORTED
✅ `reflect.Kind()` - SUPPORTED
✅ Basic value extraction (`.Int()`, `.Uint()`, `.Float()`, `.String()`) - SUPPORTED

### Location of encoding/gob Usage

**File:** `tinypdf/fpdf/def.go:7`
```go
import (
    "bytes"
    "crypto/sha1"
    "encoding/binary"
    "encoding/gob"  // ← NOT supported in TinyGo
    "encoding/json"
    // ...
)
```

**Investigation Result:** After thorough search of entire fpdf codebase:
- ❌ **NO gob.Encode() calls found**
- ❌ **NO gob.Decode() calls found**
- ❌ **NO gob.NewEncoder() calls found**
- ❌ **NO gob.NewDecoder() calls found**

**Conclusion:** `encoding/gob` is a **DEAD IMPORT** - imported but never used.

### What Each Encoding Actually Does

#### 1. encoding/gob - ❌ DEAD IMPORT
**Purpose:** NONE (never called in codebase)
**Status:** Safe to remove
**Impact:** Zero - no code depends on it

#### 2. encoding/json - ✅ ACTIVELY USED
**Purpose:** Parse attachment metadata for PDFs with embedded files
**Location:** `fpdf/attachments.go:1644`
```go
// Unmarshal JSON describing PDF attachments
err := json.Unmarshal(jsonData, &attachmentMap)
```
**Data Format:** 
```json
{
    "filename.pdf": {
        "description": "Invoice 2023",
        "relationship": "Alternative",
        "mimetype": "application/pdf"
    }
}
```
**TinyGo:** ✅ Fully compatible

#### 3. encoding/binary - ✅ ACTIVELY USED  
**Purpose:** Convert Go integers to big-endian bytes for PDF binary format
**Location:** `fpdf/attachments.go:1675-1681`
```go
// Write string length as 2-byte big-endian uint16
binary.BigEndian.PutUint16(b[0:2], uint16(len(s)))

// Write CRC32 checksum as 4-byte big-endian uint32  
binary.BigEndian.PutUint32(b[0:4], crc)
```
**Why needed:** PDF format requires numbers in big-endian byte order
**TinyGo:** ✅ Fully compatible

## TinyGo Reflection Limitations

### What IS Supported
- `reflect.ValueOf()` - Get value wrapper
- `reflect.TypeOf()` - Get type information
- `reflect.Kind()` - Get underlying kind (int, string, struct, etc.)
- Basic value extraction: `.Int()`, `.Uint()`, `.Float()`, `.String()`, `.Bool()`
- Struct field iteration (limited)

### What IS NOT Supported
- `reflect.AssignableTo()` - Type compatibility checks
- `reflect.Implements()` - Interface implementation checks
- `reflect.ConvertibleTo()` - Type conversion checks
- `reflect.MethodByName()` - Runtime method lookup
- `reflect.MakeFunc()` - Dynamic function creation
- Full `encoding/gob` package - Relies on unsupported reflection features

## Impact Assessment

### Current State
- ✅ PDF generation works with standard Go compiler
- ✅ `tinystring.Fmt()` handles custom types correctly (pdfVersion → "1.3")
- ❌ TinyGo compilation fails immediately at init time
- ❌ WASM module cannot load in browser

### Affected Components
1. **tinypdf/fpdf** - Uses `encoding/gob` (line 7 of def.go)
2. **tinystring** - Uses `reflect` for custom type handling (compatible with TinyGo)

### Critical Finding
The **tinystring reflection usage is NOT the problem**. The issue is `encoding/gob` initialization happening BEFORE any tinystring code runs. This is proven by the stack trace showing panic in `encoding/gob.init()`.

## Solution Strategy

### ✅ IMPLEMENTED: Conditional Compilation with Build Tags

**Approach:** Moved `GobEncode`/`GobDecode` methods to separate file with build tags

**Implementation:**
- Created `fpdf/def_gob.go` with build tag: `//go:build !tinygo && !js && !wasm`
- Removed `encoding/gob` import from `fpdf/def.go`
- Moved gob-dependent methods to conditional compilation file
- Standard Go builds: Methods available (backward compatible)
- TinyGo/WASM builds: Methods excluded (no gob dependency)

**Files modified:**
- `fpdf/def.go` - Removed gob import and GobEncode/GobDecode methods
- `fpdf/def_gob.go` - NEW file with gob methods (only for non-TinyGo builds)

**Status:** ✅ COMPLETE
**Result:** 
- Standard Go: All tests pass (backward compatible)
- TinyGo: Should compile without gob panic

### Option 1: Remove encoding/gob (CONSIDERED BUT NOT CHOSEN)

**Approach:** Find and eliminate all `encoding/gob` usage in fpdf

**Steps:**
1. Locate all gob.Encode/Decode calls in fpdf codebase
2. Determine purpose (likely font caching or state serialization)
3. Replace with TinyGo-compatible alternatives:
   - JSON serialization (`encoding/json` - TinyGo compatible)
   - Custom binary serialization using `encoding/binary`
   - Direct struct field access (no serialization)

**Pros:**
- Fully TinyGo compatible
- Smaller WASM binary size
- Faster initialization

**Cons:**
- Requires code changes in fpdf library
- May lose some fpdf features (if gob is critical)

### Option 2: Conditional Compilation with Build Tags

**Approach:** Use build tags to exclude gob-dependent code in WASM builds

**Implementation:**
```go
//go:build !tinygo

package fpdf

import "encoding/gob"
// ... gob-dependent code
```

```go
//go:build tinygo

package fpdf

// ... alternative implementation without gob
```

**Pros:**
- Maintains full functionality in standard Go
- Clean separation of TinyGo-specific code

**Cons:**
- Requires maintaining two code paths
- More complex build system

### Option 3: Fork fpdf with TinyGo Support

**Approach:** Create `fpdf-tinygo` fork without encoding/gob

**Pros:**
- Upstream fpdf remains unchanged
- Full control over TinyGo compatibility

**Cons:**
- Maintenance burden (keeping up with upstream)
- Potential divergence from original library

## Investigation Tasks

### Phase 1: Assess encoding/gob Usage ✅ COMPLETE
- [x] Search all `gob.Encode` calls in fpdf codebase → Found in GobEncode method
- [x] Search all `gob.Decode` calls in fpdf codebase → Found in GobDecode method
- [x] Document what features depend on gob → ImageInfoType serialization (unused)
- [x] Determine if gob is essential or optional → Optional (methods never called)

**Findings:**
- `GobEncode`/`GobDecode` methods defined but **never called** in codebase
- No external code uses these methods (grep search confirmed)
- Safe to exclude from TinyGo builds without losing functionality

### Phase 2: Implement Solution ✅ COMPLETE
- [x] Created `fpdf/def_gob.go` with conditional build tags
- [x] Moved gob methods to separate file
- [x] Removed gob import from main def.go
- [x] Tested with standard Go (all tests pass)

### Phase 3: Verify TinyGo Compilation (NEXT)
- [ ] Compile with TinyGo: `tinygo build -target=wasm`
- [ ] Verify no gob panic in browser console
- [ ] Test PDF generation in WASM environment
- [ ] Confirm browser displays PDF correctly

## Test Plan

### Test 1: Minimal TinyGo WASM with tinystring
```go
package main

import (
    "syscall/js"
    "github.com/tinywasm/fmt"
)

type customVersion uint16

func (v customVersion) String() string {
    return tinystring.Fmt("%d.%d", byte(v>>8), byte(v))
}

func main() {
    version := customVersion(0x0103) // 1.3
    result := tinystring.Fmt("PDF-%s", version)
    js.Global().Get("console").Call("log", result) // Should print: "PDF-1.3"
    select{}
}
```

**Expected:** Works without panic (proves tinystring reflection is TinyGo-compatible)

### Test 2: fpdf without gob
After removing `encoding/gob`:
```bash
GOOS=js GOARCH=wasm tinygo build -o test.wasm ./src/cmd/webclient/
```

**Expected:** Compiles successfully and generates PDF in browser

## Related Issues

- **BUG_PDF_FAILED_IN_BROWSER.md** - Initial PDF loading failure (RESOLVED with Blob API)
- **BUG_FMT_CUSTOM_TYPE.md** - tinystring.Fmt custom type support (RESOLVED with reflection)

## References

### TinyGo Documentation
- https://tinygo.org/docs/reference/lang-support/stdlib/ - Standard library support
- https://tinygo.org/docs/reference/lang-support/reflect/ - Reflection limitations

### Key TinyGo Limitations
> "The reflect package is only partially implemented. Many methods have not yet been 
> implemented, including AssignableTo, Implements, ConvertibleTo, and others needed 
> by encoding/gob."

### Alternative Serialization Options
- `encoding/json` - Full TinyGo support
- `encoding/binary` - Full TinyGo support
- Custom serialization - Type-specific encode/decode

## Next Steps

1. **Immediate**: Search fpdf codebase for all gob usage patterns
2. **Analysis**: Determine if gob is used for critical features or optional optimization
3. **Decision**: Choose between Option 1 (remove gob) vs Option 2 (build tags)
4. **Implementation**: Apply chosen solution and test in TinyGo/WASM
5. **Validation**: Verify PDF generation works in browser with TinyGo compilation

## Expected Outcome

After removing or conditionally excluding `encoding/gob`:
- ✅ TinyGo compilation succeeds
- ✅ WASM module loads in browser
- ✅ PDF generation works (header: "%PDF-1.3")
- ✅ Custom types format correctly via tinystring.Fmt()
- ✅ No runtime panics or reflection errors

## Risk Assessment

**Low Risk**: tinystring reflection usage (TinyGo compatible primitives)
**High Risk**: encoding/gob dependency (fundamentally incompatible)
**Mitigation**: Remove or replace gob with JSON/binary serialization
