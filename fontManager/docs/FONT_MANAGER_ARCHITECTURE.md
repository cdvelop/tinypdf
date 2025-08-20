type FontMetrics struct {
type EmbeddedProvider struct {
type FilesystemProvider struct {
type NetworkProvider struct {
# Arquitectura del FontManager - Requisitos y Propuesta (actualizada)

## Resumen del encargo

Actualizar la propuesta de refactorización para cumplir las restricciones siguientes:
- No usar paquetes estándar de `fmt`, `strconv` ni `errors` para formateo o creación de errores; usar únicamente `tinystring` (p. ej. `tinystring.Errf`, `tinystring.Fmt`) por ser liviano y compatible con TinyGo.
- No usar `embed` ni fuentes embebidas en el binario: la librería NO incluirá fuentes por defecto.
- Las fuentes se obtendrán exclusivamente de:
    - el directorio `fonts/` cuando se ejecute en servidor (build tag `!wasm`), o
    - por HTTP en el navegador, usando `syscall/js` y la API `fetch` (build tag `wasm`).

El `FontManager` deberá ser la única entrada (punto central) para listar, cargar y parsear fuentes.

## Checklist de requisitos (extraídos de tu petición)

- [x] Usar solo `tinystring` para errores y formateo
- [x] No usar `embed` ni incluir fuentes por defecto
- [x] Cargar fuentes desde `fonts/` en servidor (filesystem)
- [x] Cargar fuentes por HTTP en navegador vía `syscall/js` (fetch)
- [x] Diseñar `FontManager` como única entrada para parsing/obtención de TTF
 - [x] Procesar únicamente archivos con extensión `.ttf` (ignorar otros ficheros en `fonts/`)

## Revisión rápida del diseño recomendado

Objetivo: mantener modularidad y compatibilidad con TinyGo, pero adaptando la estrategia de I/O para que NO dependa de `embed` y para que los clientes obtengan fuentes siempre vía `FontManager`.

### Cambios principales respecto a la versión anterior

- Quitar cualquier mención a `EmbeddedProvider` o a `//go:embed`.
- Definir dos providers concretos:
    - `FilesystemProvider` (build tag `!wasm`) — lee `fonts/` usando `os` y `io`.
    - `BrowserHTTPProvider` (build tag `wasm`) — realiza peticiones HTTP con `syscall/js` y `fetch`, devolviendo bytes.
- Mantener el parser de TTF independiente del origen de los bytes: la función de bajo nivel seguirá siendo `TtfParse(r Reader)` (ya está implementada en `fontManager/ttf_parser.go`).
- Exponer en `FontManager` una API única para parsear por nombre de fuente: `func (fm *FontManager) ParseTtfByName(name string) (TtfType, error)` — internamente llamará al provider para obtener los bytes y luego a `TtfParse(bytes.NewReader(data))`.
- Usar `tinystring.Errf` para todos los errores dentro del módulo.

## API propuesta (centrada en `FontManager` como única entrada)

Firma mínima recomendada:

```go
type FontProvider interface {
        ListFonts() ([]string, error)
        LoadFont(name string) ([]byte, error)
}

type FontManager struct {
        provider FontProvider
        // ... registry, logger (opcional, usar tinystring.Fmt para mensajes)
}

// Constructor por entorno
// - servidor: NewFontManagerFromFS(basePath string, logger func(...any))
// - navegador: NewFontManagerFromBrowser(baseURL string, logger func(...any))

// Operaciones clave
func (fm *FontManager) ListFonts() ([]string, error)
func (fm *FontManager) LoadAndCreateFontDef(name string) (*FontDef, error)
func (fm *FontManager) ParseTtfByName(name string) (TtfType, error)
```

Notas de implementación:
- `ParseTtfByName` debe:
    1. llamar a `fm.provider.LoadFont(name)`
    2. construir un `bytes.Reader` y pasar a `TtfParse(r Reader)` (la función existente)
    3. devolver `TtfType` o `tinystring.Errf(...)` en caso de fallo

Nota importante sobre el directorio `fonts/`:

- Nunca se asumirá qué archivos contiene `fonts/`. El `FilesystemProvider` debe listar todos los nombres, filtrar y procesar sólo aquellos con la extensión `.ttf` (comparación insensible a mayúsculas). Cualquier otro archivo en `fonts/` debe ser ignorado sin generar fallos.

Ejemplo de política al listar archivos en `fonts/`:

- `ReadDir(basePath)` → por cada entrada: si es archivo y `strings.HasSuffix(strings.ToLower(name), ".ttf")` entonces incluir en la lista; en caso contrario, omitir. Errores de lectura del directorio deben reportarse con `tinystring.Errf`.

## Detalles del `BrowserHTTPProvider` (WASM)

Requisitos:
- Usar `syscall/js` para invocar `fetch()` y leer el `ArrayBuffer` devuelto.
- Implementar con build tag `//go:build wasm`.

Esquema (pseudo-código):

```go
//go:build wasm
package fontManager

import "syscall/js"

type BrowserHTTPProvider struct {
        baseURL string // Base URL donde se sirven las fuentes (por ejemplo "/fonts/")
}

func (p *BrowserHTTPProvider) LoadFont(name string) ([]byte, error) {
        // usar syscall/js para fetch(p.baseURL + name)
        // obtener ArrayBuffer, copiar a []byte y devolver
        // en errores devolver tinystring.Errf("could not fetch %s: %v", name, err)
}
```

Puntos importantes:
- `syscall/js` permite interoperar con la promesa devuelta por `fetch`; hay que transformar la promesa en una espera (then) y bloquear hasta obtener el resultado o usar callbacks apropiados.
- TinyGo soporta `syscall/js`, por lo que esta ruta es viable para el objetivo de compatibilidad.

## Detalles del `FilesystemProvider` (servidor / !wasm)

Implementación simple (build tag `!wasm`) que usa `os.ReadDir` y `os.ReadFile`.
Errores formateados con `tinystring.Errf`.

## Manejo de errores y logging

- Todos los mensajes de error y formateos deben usar `tinystring` (ej.: `tinystring.Errf("could not read font: %w", err)` o similar API que provea `tinystring`).
- No utilizar `fmt`, `strconv` ni `errors` dentro del módulo. Esto garantiza que la dependencia sea tiny y TinyGo-friendly.

## Compatibilidad con compresión (opcional)

- Si `compress/zlib` se usa, controlarlo por build tags (`!tinygo`) y en TinyGo dejar la opción sin compresión o con una alternativa ligera. Esto no afecta la política sobre `fmt`/`errors`.

## Plan de trabajo recomendado (pasos siguientes)

1. Implementar `FontProvider` concreto para `FilesystemProvider` (build tag `!wasm`) y exponer constructor `NewFontManagerFromFS(basePath string, logger func(...any))`.
2. Implementar `BrowserHTTPProvider` (build tag `wasm`) usando `syscall/js` y exponer `NewFontManagerFromBrowser(baseURL string, logger func(...any))`.
3. Añadir método `ParseTtfByName` en `FontManager` que use `LoadFont` + `TtfParse(reader)`.
4. Reescribir los tests actuales para invocar `fm.ParseTtfByName("calligra.ttf")` o similar, en lugar de `fontManager.TtfParse()` sin parámetros.

Cuando confirmes que este diseño te parece correcto, procedo a:
- crear las plantillas (`FilesystemProvider`, `BrowserHTTPProvider`) y
- adaptar `fontManager/loader.go` y `fontManager/fontManager.go` para que deleguen en `provider.LoadFont` y en `ParseTtfByName`.

Espero tu corrección o aprobación para avanzar con la implementación del código.
    fonts map[string][]byte // Fuentes embebidas
