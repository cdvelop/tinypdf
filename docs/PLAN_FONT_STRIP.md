# PLAN: Strip tablas innecesarias del font subset TTF

## Objetivo
Reducir el tamano del PDF de ~20KB a ~9KB eliminando tablas TrueType que no son necesarias para renderizado en PDF.

## Contexto
El font subset embebido en el PDF pesa 20,156 bytes (16,386 comprimido). De esos, 12,826 bytes (64%) son tablas innecesarias para PDF.

| Tabla | Bytes | Para que sirve | Necesaria en PDF? |
|-------|-------|----------------|-------------------|
| name | 6,925 | Metadatos: nombre, copyright, licencia | No — PDF usa FontDescriptor |
| prep | 2,813 | Programa de hinting para pantalla | No — PDF no rasteriza |
| cvt | 1,584 | Valores de control para hinting | No — PDF no rasteriza |
| fpgm | 1,456 | Instrucciones TrueType para hinting | No — PDF no rasteriza |
| post | 32 | Nombres PostScript de glifos | No — PDF usa CIDToGIDMap |
| gasp | 16 | Grid-fitting and scan-conversion | No — PDF no rasteriza |

Tablas que se mantienen (obligatorias): glyf, loca, head, hhea, hmtx, maxp, cmap, OS/2.

## Estimacion
- Subset actual: 20,156 bytes raw -> 16,386 comprimido
- Subset sin tablas: ~7,330 bytes raw -> ~5,900 comprimido
- PDF total estimado: ~8-9 KB (vs 20 KB actual)

## Cambio

### Archivo: `fpdf/utf8fontfile.go`

Un solo bloque a modificar. Lineas 659-667 actuales:

```go
utf.setOutTable("name", utf.getTableData("name"))
utf.setOutTable("cvt ", utf.getTableData("cvt "))
utf.setOutTable("fpgm", utf.getTableData("fpgm"))
utf.setOutTable("prep", utf.getTableData("prep"))
utf.setOutTable("gasp", utf.getTableData("gasp"))

postTable := utf.getTableData("post")
postTable = append(append([]byte{0x00, 0x03, 0x00, 0x00}, postTable[4:16]...), []byte{...}...)
utf.setOutTable("post", postTable)
```

Reemplazar por:

```go
// Minimal post table (format 3 = no glyph names, required by some viewers)
utf.setOutTable("post", []byte{
    0x00, 0x03, 0x00, 0x00, // format 3.0
    0x00, 0x00, 0x00, 0x00, // italicAngle
    0x00, 0x00,             // underlinePosition
    0x00, 0x00,             // underlineThickness
    0x00, 0x00, 0x00, 0x00, // isFixedPitch
    0x00, 0x00, 0x00, 0x00, // minMemType42
    0x00, 0x00, 0x00, 0x00, // maxMemType42
    0x00, 0x00, 0x00, 0x00, // minMemType1
    0x00, 0x00, 0x00, 0x00, // maxMemType1
})
```

Se eliminan: `name`, `cvt `, `fpgm`, `prep`, `gasp`.
Se mantiene: `post` minimo (32 bytes, format 3) porque algunos visores PDF lo requieren.

### Riesgo
Bajo. Las tablas eliminadas solo afectan hinting en pantalla y metadatos. Todos los visores PDF modernos (Acrobat, Chrome, Firefox, Preview) renderizan correctamente sin ellas. El format 3 de post es el estandar para subsets.

### Verificacion
1. `go test ./...` — tests existentes
2. Generar demo.pdf y verificar tamano < 10KB
3. Abrir en Chrome, Firefox y un lector PDF nativo
4. Verificar que caracteres UTF-8 (acentos, n) se renderizan correctamente
