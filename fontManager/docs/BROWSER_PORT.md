# Propuesta para Refactorización de fontManager - Compatible con TinyGo

## Objetivo
Separar fontManager de tinypdf y crear un módulo independiente compatible con TinyGo, comenzando con hacer funcionar el test `ExampleTtfParse()`.

## Análisis del Estado Actual

### Problema Principal
El test `ExampleTtfParse()` está intentando llamar a `fontManager.TtfParse()` sin parámetros, pero la función actual `TtfParse(r Reader)` requiere un Reader como parámetro.

### Dependencias Actuales
- **tinystring**: Usado para `Errf()` (manejo de errores)
- **compress/zlib**: Para compresión de datos de fuente
- **encoding/binary**: Para lectura de datos binarios
- **os/filepath**: Solo en build tag `!wasm`

### Estructura Actual
```
fontManager/
├── ttf_parser.go          # Parser TTF con Reader interface
├── interfaces.go          # Reader interface para abstracción
├── loader_stdlib.go       # Carga desde filesystem (!wasm)
├── loader_wasm.go         # Carga para WASM (stub)
├── font_types.go          # Estructuras de datos
├── fontManager.go         # Manager principal
└── fonts/calligra.ttf     # Font de prueba
```

## Propuesta de Implementación - Fase 1

### 1. Crear Función TtfParse() Sin Parámetros

**Ubicación**: `fontManager/ttf_parser.go`

**Opción A - Con Font Embebida (Recomendada para el test)**
```go
// TtfParse() sin parámetros que use la fuente de ejemplo embebida
func TtfParse() (TtfType, error) {
    reader := NewEmbeddedFontReader("calligra.ttf")
    return TtfParseFromReader(reader)
}

// Renombrar la función actual
func TtfParseFromReader(r Reader) (TtfType, error) {
    // Código actual de TtfParse(r Reader)
}
```

**Opción B - Con Path por Defecto**
```go
func TtfParse() (TtfType, error) {
    // Buscar en directorio por defecto
    fontPath := getDefaultFontPath("calligra.ttf")
    reader, err := NewFileReader(fontPath)
    if err != nil {
        return TtfType{}, err
    }
    defer reader.Close()
    return TtfParseFromReader(reader)
}
```

### 2. Implementar EmbeddedFontReader (Compatible TinyGo)

**Archivo**: `fontManager/embedded_reader.go`

```go
//go:embed fonts/calligra.ttf
var calligraFont []byte

type EmbeddedReader struct {
    data []byte
    pos  int64
}

func NewEmbeddedFontReader(name string) *EmbeddedReader {
    switch name {
    case "calligra.ttf":
        return &EmbeddedReader{data: calligraFont}
    default:
        return nil
    }
}

func (r *EmbeddedReader) Read(p []byte) (int, error) {
    // Implementación de lectura
}

func (r *EmbeddedReader) Seek(offset int64, whence int) (int64, error) {
    // Implementación de seek
}
```

### 3. Compatibilidad TinyGo

**Consideraciones**:
- ✅ `embed` es soportado por TinyGo
- ✅ `encoding/binary` es soportado
- ⚠️ `compress/zlib` puede tener limitaciones
- ❌ `os` no disponible en WASM

**Soluciones**:
1. **Para compress/zlib**: Crear implementación alternativa o usar build tags
2. **Para os**: Ya manejado con build tags (!wasm)

## Preguntas y Alternativas

### Pregunta 1: ¿Cuál opción prefieres para TtfParse()?

**Opción A - Font Embebida**
- ✅ Pro: Autocontenida, funciona sin filesystem
- ✅ Pro: Compatible con WASM/browser
- ✅ Pro: Test determinístico
- ❌ Contra: Aumenta tamaño del binario
- ❌ Contra: Solo una fuente disponible

**Opción B - Path por Defecto**
- ✅ Pro: Flexible, puede usar cualquier fuente
- ✅ Pro: No aumenta tamaño binario
- ❌ Contra: Dependiente de filesystem
- ❌ Contra: No funciona en WASM

**Opción C - Configuración Global**
```go
var defaultFont *EmbeddedReader

func SetDefaultFont(reader Reader) {
    defaultFont = reader
}

func TtfParse() (TtfType, error) {
    if defaultFont == nil {
        defaultFont = NewEmbeddedFontReader("calligra.ttf")
    }
    return TtfParseFromReader(defaultFont)
}
```

### Pregunta 2: ¿Cómo manejar compress/zlib en TinyGo?

**Opción A - Build Tags**
```go
//go:build !tinygo
// Usar zlib estándar

//go:build tinygo  
// Implementación simplificada o sin compresión
```

**Opción B - Dependencia Externa**
- Buscar implementación de zlib compatible con TinyGo
- Ejemplo: `github.com/klauspost/compress`

**Opción C - Compresión Opcional**
```go
func createFontDefFromTtf(ttf TtfType, fontData []byte) (*FontDef, error) {
    def := &FontDef{...}
    
    if compressionAvailable() {
        def.Data = compress(fontData)
    } else {
        def.Data = fontData // Sin compresión
    }
    
    return def, nil
}
```

### Pregunta 3: ¿Estructura del módulo independiente?

**Propuesta**:
```
github.com/cdvelop/envfonts/
├── go.mod
├── ttf_parser.go          # Parser principal
├── embedded_reader.go     # Reader para fonts embebidas  
├── file_reader.go         # Reader para archivos (!wasm)
├── types.go              # Estructuras de datos
├── interfaces.go         # Interfaces
└── fonts/
    └── calligra.ttf      # Font de ejemplo
```

**Dependencias mínimas**:
- Solo stdlib de Go
- Sin dependencias externas para máxima compatibilidad

### Pregunta 4: ¿Manejo de errores sin tinystring?

**Opción A - fmt.Errorf estándar**
```go
import "fmt"

// Reemplazar Errf() con fmt.Errorf()
err = fmt.Errorf("fonts based on PostScript outlines are not supported")
```

**Opción B - Errores personalizados**
```go
type FontError struct {
    Message string
    Cause   error
}

func (e FontError) Error() string {
    if e.Cause != nil {
        return e.Message + ": " + e.Cause.Error()
    }
    return e.Message
}
```

## Plan de Implementación Paso a Paso

### Paso 1: Preparar el Test
1. Crear función `TtfParse()` sin parámetros
2. Usar EmbeddedReader con calligra.ttf
3. Verificar que el test pase

### Paso 2: Refactorizar Dependencias
1. Reemplazar `Errf()` con `fmt.Errorf()`
2. Manejar compress/zlib con build tags
3. Verificar compatibilidad TinyGo

### Paso 3: Modularizar
1. Crear nuevo módulo github.com/cdvelop/envfonts
2. Mover código refactorizado
3. Actualizar imports en tinypdf

### Paso 4: Testing Completo
1. Tests unitarios para todas las funciones
2. Test de integración con tinypdf
3. Test de compilación con TinyGo

## Decisiones Requeridas

Por favor, indícame tu preferencia para:

1. **Función TtfParse()**: ¿Opción A (embebida), B (path) o C (configurable)?
2. **Compresión**: ¿Build tags, dependencia externa o opcional?
3. **Manejo errores**: ¿fmt.Errorf o errores personalizados?
4. **Nombre del módulo**: ¿envfonts o prefieres otro nombre?

Una vez que tengas estas decisiones, procederé con la implementación paso a paso.