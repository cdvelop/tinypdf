# tinypdf

tinypdf is a lightweight PDF generation library built with Go, designed to work seamlessly with both a backend server and a WebAssembly frontend. This project leverages TinyGo for efficient compilation and execution, enabling fast and responsive PDF generation directly in the browser. This library is a derivative of `github.com/jung-kurt/gofpdf`.

## Cross-Platform Support

tinypdf provides consistent functionality across different environments:

- **Backend (Server)**: Works with local file system operations
- **Frontend (Browser/WASM)**: Works with URLs and in-memory content

## File Handling

### Standard File Operations

The library handles files through the environment-agnostic `env` package:

```go
// Read file content
content, err := env.FileExists("path/to/file.txt") // Backend: local file, Frontend: URL

// Get file size
size, err := env.GetSize("path/to/file.txt") // Backend: file size, Frontend: content length

// Open file for reading with seek support
reader, err := env.FileOpen("path/to/file.txt") // Returns ReadSeekCloser
defer reader.Close()
```

### Large File Processing

For large files, tinypdf provides streaming capabilities to process content without loading it entirely into memory:

```go
// Process a large file in 8KB chunks
chunkSize := 8 * 1024
err := env.StreamFile("path/to/large-file.jpg", chunkSize, func(chunk []byte) error {
    // Process each chunk
    fmt.Printf("Processing %d bytes\n", len(chunk))
    return nil
})
```

### Built-in Large File Handlers

The library includes specialized functions for handling large files:

```go
// Load a large image without consuming excessive memory
pdf := tinypdf.New("P", "mm", "A4", "")
err := pdf.LoadLargeImage("path/to/large-image.jpg", "jpg")

// Stream PDF output to a writer (useful for HTTP responses)
http.HandleFunc("/generate-pdf", func(w http.ResponseWriter, r *http.Request) {
    pdf := tinypdf.New("P", "mm", "A4", "")
    // Add content to PDF...
    
    w.Header().Set("Content-Type", "application/pdf")
    pdf.StreamPDFOutput(w)
})
```
