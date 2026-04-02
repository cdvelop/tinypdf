# TinyPDF Web Demo

Web application that generates PDFs in the browser using WebAssembly.

## Prerequisites

- [Go](https://go.dev/) 1.21+
- [TinyGo](https://tinygo.org/) 0.39+

## Build

Compile the WASM client:

```bash
tinygo build -o public/client.wasm -target wasm -no-debug ./client.go
```

> The `-no-debug` flag strips debug info to reduce binary size.

## Run

Start the static file server:

```bash
go run server.go
```

Open [http://localhost:4430](http://localhost:4430) in your browser.

### Options

```bash
go run server.go -port 8080 -public-dir ./public
```

Environment variables `PORT` and `PUBLIC_DIR` are also supported.

## Project structure

```
web/
  client.go      WASM entry point (build tag: wasm)
  server.go      Static file server (build tag: !wasm)
  public/
    index.html   HTML shell
    script.js    Go WASM bootstrap (instantiates client.wasm)
    style.css    Styles
    client.wasm  Compiled WASM binary (generated)
  ui/
    ui.go        DOM setup and event handlers
    pdf.go       PDF generation logic
```
