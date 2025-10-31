# BUG: PDF Failed to Load in Browser

**Fecha**: 2025-10-31  
**Estado**: En Investigación  
**Severidad**: Alta  
**Entorno**: WebAssembly (WASM) / Browser

## Descripción del Problema

Al generar un PDF en el navegador usando la aplicación TinyPDF compilada a WebAssembly, el documento se genera exitosamente según los logs de la consola, pero el navegador muestra el error "Failed to load PDF document" cuando intenta renderizar el PDF en el elemento `<embed>`.

### Logs de la Consola (Exitosos)
```
GOLITE initialized
DOM fully loaded
TinyPDF inicializado correctamente
Aplicación lista
Iniciando generación de PDF...
PDF generado exitosamente
```

### Síntomas
- ✅ El PDF se genera sin errores según los logs
- ✅ Los bytes del PDF se convierten a base64 correctamente
- ✅ Se crea el data URL: `data:application/pdf;base64,[data]`
- ❌ El navegador no puede cargar/renderizar el PDF
- ❌ Mensaje de error: "Failed to load PDF document"

## Flujo Actual del Código

### 1. Generación del PDF (`src/web/ui/pdf.go`)
```go
func GeneratePDF() {
    // ... configuración del PDF ...
    
    // Obtener el PDF como bytes
    var buf []byte
    err := pdf.Output(&bytesWriter{data: &buf})
    
    // Convertir a base64
    base64Data := base64.StdEncoding.EncodeToString(buf)
    dataURL := "data:application/pdf;base64," + base64Data
    
    ShowPDF(dataURL)
}
```

### 2. Escritura de Archivos (`env.front.go`)
```go
func (tp *TinyPDF) writeFile(filePath string, content []byte) error {
    localStorage := js.Global().Get("localStorage")
    
    // Codificar el contenido en base64 para almacenarlo
    encoded := base64.StdEncoding.EncodeToString(content)
    localStorage.Call("setItem", filePath, encoded)
    
    return nil
}
```

**NOTA**: Actualmente `writeFile` no se está llamando explícitamente en el flujo de generación del PDF en el navegador.

## Análisis del Problema

### Hipótesis Principal: Doble Codificación Base64 Innecesaria

El flujo actual puede estar causando problemas por:

1. **Conversión redundante**: Los bytes del PDF se convierten a base64 en `GeneratePDF()` para crear un data URL
2. **localStorage no utilizado**: La función `writeFile()` que guarda en localStorage con base64 no se invoca en el flujo actual
3. **Posible corrupción de datos**: El data URL con base64 puede estar malformado o truncado

### Investigación Técnica

#### Problemas Potenciales con Base64 en Data URLs:

1. **Limitaciones de tamaño**: Los data URLs tienen límites de tamaño en navegadores:
   - Chrome: ~512MB (teórico, pero problemas de rendimiento antes)
   - Firefox: Sin límite oficial pero problemas de memoria
   - Safari: ~10MB recomendado

2. **Rendimiento**: La conversión base64 aumenta el tamaño en ~33%:
   - PDF original: X bytes
   - Base64 codificado: X * 1.33 bytes
   - Más overhead del string "data:application/pdf;base64,"

3. **Codificación correcta**: Problemas comunes:
   - Caracteres especiales no escapados
   - Padding incorrecto del base64
   - Newlines en el base64 (aunque EncodeToString no debería generarlos)

## Propuesta de Mejora: Uso de Blob API

### Ventajas de Usar Blob en lugar de Base64

#### 1. **Rendimiento Mejorado**
```javascript
// Actual (ineficiente):
data:application/pdf;base64,JVBERi0... (33% más grande)

// Con Blob (eficiente):
blob:http://localhost:8080/uuid (referencia pequeña)
```

#### 2. **Manejo Nativo del Navegador**
- Los Blobs son objetos nativos del navegador para datos binarios
- Mejor manejo de memoria (garbage collection automático)
- Sin conversión de encoding necesaria

#### 3. **Código Más Limpio**
```go
// Propuesta de implementación en WASM:
func ShowPDFFromBytes(pdfBytes []byte) {
    // Crear un Uint8Array desde los bytes de Go
    jsBytes := js.Global().Get("Uint8Array").New(len(pdfBytes))
    js.CopyBytesToJS(jsBytes, pdfBytes)
    
    // Crear Blob con tipo MIME correcto
    blobParts := js.Global().Get("Array").New()
    blobParts.Call("push", jsBytes)
    
    blobOptions := js.Global().Get("Object").New()
    blobOptions.Set("type", "application/pdf")
    
    blob := js.Global().Get("Blob").New(blobParts, blobOptions)
    
    // Crear URL del Blob
    blobURL := js.Global().Get("URL").Call("createObjectURL", blob)
    
    // Usar el blob URL directamente
    embed.Set("src", blobURL)
}
```

### 4. **Mejor Gestión de Recursos**
```go
// Limpiar URLs de Blob cuando ya no se necesiten
func CleanupBlobURL(blobURL string) {
    js.Global().Get("URL").Call("revokeObjectURL", blobURL)
}
```

## Modificaciones Propuestas

### Archivo: `src/web/ui/pdf.go`

#### Función `GeneratePDF()` - Modificar:
```go
func GeneratePDF() {
    TP.Log("Iniciando generación de PDF...")
    
    titleText := GetTitleText()
    pdf := TP.Fpdf
    
    // ... configuración del PDF ...
    
    var buf []byte
    err := pdf.Output(&bytesWriter{data: &buf})
    if err != nil {
        TP.Log("Error generando PDF:", err.Error())
        ShowError("Error al generar PDF: " + err.Error())
        return
    }
    
    // NUEVO: Usar Blob en lugar de base64
    ShowPDFFromBytes(buf)
    
    TP.Log("PDF generado exitosamente")
}
```

#### Nueva función `ShowPDFFromBytes()`:
```go
func ShowPDFFromBytes(pdfBytes []byte) {
    // Crear Uint8Array desde bytes de Go
    jsArray := js.Global().Get("Uint8Array").New(len(pdfBytes))
    js.CopyBytesToJS(jsArray, pdfBytes)
    
    // Crear Blob
    array := js.Global().Get("Array").New(jsArray)
    blobOpts := map[string]interface{}{"type": "application/pdf"}
    blob := js.Global().Get("Blob").New(array, blobOpts)
    
    // Crear Object URL
    blobURL := js.Global().Get("URL").Call("createObjectURL", blob).String()
    
    // Mostrar en embed
    ShowPDF(blobURL)
}
```

### Archivo: `env.front.go`

#### Consideración para `writeFile()`:
La función `writeFile()` actualmente no se usa en el flujo de generación del PDF. Tiene dos opciones:

**Opción A: Mantener base64 para localStorage (archivos persistentes)**
```go
// Mantener la implementación actual para cuando se necesite
// guardar PDFs persistentemente en localStorage
func (tp *TinyPDF) writeFile(filePath string, content []byte) error {
    // ... implementación actual ...
}
```

**Opción B: Usar IndexedDB para archivos grandes**
```go
// Para archivos grandes, IndexedDB es mejor que localStorage
func (tp *TinyPDF) writeFile(filePath string, content []byte) error {
    // TODO: Implementar usando IndexedDB API
    // Ventaja: Soporta Blobs nativamente
    // Sin conversión base64 necesaria
}
```

## Plan de Acción

### Fase 1: Implementación de Blob API ✅
1. ✅ Documentar el problema actual
2. ✅ Implementar `ShowPDFFromBytes()` en `pdf.go`
3. ✅ Modificar `GeneratePDF()` para usar Blob
4. ✅ Agregar llamada explícita a `pdf.Close()` antes de `Output()`
5. ⏳ Probar la generación y visualización del PDF

### Hallazgo Importante:
**Problema encontrado**: El PDF no se estaba cerrando correctamente antes de llamar a `Output()`.
- Aunque `Output()` llama internamente a `Close()` si `state < 3`, es mejor práctica llamarlo explícitamente
- Esto asegura que todos los metadatos y estructura del PDF se escriban correctamente
- **Solución**: Agregar `pdf.Close()` antes de `pdf.Output()`

### Fase 2: Optimizaciones
1. ⏳ Implementar limpieza de Blob URLs (`revokeObjectURL`)
2. ⏳ Agregar botón de descarga del PDF
3. ⏳ Considerar IndexedDB para `writeFile()` si se necesita persistencia

### Fase 3: Testing
1. ⏳ Verificar que el PDF se renderiza correctamente
2. ⏳ Probar con PDFs de diferentes tamaños
3. ⏳ Verificar uso de memoria y performance

## Referencias

- [MDN - Blob API](https://developer.mozilla.org/en-US/docs/Web/API/Blob)
- [MDN - URL.createObjectURL()](https://developer.mozilla.org/en-US/docs/Web/API/URL/createObjectURL)
- [syscall/js - CopyBytesToJS](https://pkg.go.dev/syscall/js#CopyBytesToJS)
- [Data URLs vs Blob URLs Performance](https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/Data_URIs)

## Notas Adicionales

### ¿Por qué Base64 es problemático aquí?

1. **Overhead de memoria**: Base64 convierte 3 bytes en 4 caracteres, aumentando el tamaño
2. **Conversión CPU-intensiva**: Codificar/decodificar base64 consume CPU
3. **Strings inmutables**: En Go y JavaScript, los strings son inmutables, creando copias adicionales
4. **Límites del navegador**: Algunos navegadores tienen problemas con data URLs muy grandes

### ¿Cuándo usar Base64?

- Para datos pequeños (< 100KB)
- Cuando se necesita embeber en HTML/CSS
- Para transmisión en JSON/XML
- **NO para visualización directa de archivos grandes**

### ¿Por qué Blob es mejor?

- Representa datos binarios de forma nativa
- Manejo eficiente de memoria por el navegador
- Soporte para streaming y procesamiento parcial
- APIs nativas para descarga, visualización, etc.
- No requiere conversión de encoding

---

**Estado**: Documentación completa. Esperando aprobación para implementar la solución propuesta.
