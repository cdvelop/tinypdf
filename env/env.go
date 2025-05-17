// Package env provides environment-agnostic file and logging operations
// for both backend (OS) and frontend (browser/WASM) environments.
package env

import (
	"bytes"
	"io"
)

// FileWriter defines a function type for writing files
type FileWriter func(filename string, data []byte) error

// Logger defines a function type for logging
type Logger func(a ...any)

type ReaderCloser interface {
	io.Reader
	io.Closer
}

type ReadSeekCloser interface {
	ReaderCloser
	io.Seeker
}

// ReadAll reads all data from the given reader and returns it as a byte slice
func ReadAll(r io.Reader) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// BufferFromReader returns a new buffer populated with the contents of the specified Reader
func BufferFromReader(r io.Reader) (*bytes.Buffer, error) {
	b := new(bytes.Buffer)
	_, err := b.ReadFrom(r)
	return b, err
}
