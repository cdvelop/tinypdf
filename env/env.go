// Package env provides environment-agnostic file and logging operations
// for both backend (OS) and frontend (browser/WASM) environments.
package env

import (
	"bytes"
	"fmt"
	"io"
	"runtime"
	"strings"
)

// ReadSeekCloser is an interface that combines io.ReadSeeker and io.Closer
type ReadSeekCloser interface {
	io.ReadSeeker
	io.Closer
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

// ByteReader implements ReadSeekCloser using a byte slice
type ByteReader struct {
	data   []byte
	pos    int64
	closed bool
}

// NewByteReader creates a new ByteReader from a byte slice
func NewByteReader(data []byte) *ByteReader {
	return &ByteReader{
		data:   data,
		pos:    0,
		closed: false,
	}
}

// Read implements io.Reader for ByteReader
func (br *ByteReader) Read(p []byte) (n int, err error) {
	if br.closed {
		return 0, io.ErrClosedPipe
	}

	if br.pos >= int64(len(br.data)) {
		return 0, io.EOF
	}

	n = copy(p, br.data[br.pos:])
	br.pos += int64(n)
	return n, nil
}

// Seek implements io.Seeker for ByteReader
func (br *ByteReader) Seek(offset int64, whence int) (int64, error) {
	if br.closed {
		return 0, io.ErrClosedPipe
	}

	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = br.pos + offset
	case io.SeekEnd:
		newPos = int64(len(br.data)) + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}

	if newPos < 0 {
		return 0, fmt.Errorf("negative position: %d", newPos)
	}

	br.pos = newPos
	return br.pos, nil
}

// Close implements io.Closer for ByteReader
func (br *ByteReader) Close() error {
	if br.closed {
		return io.ErrClosedPipe
	}
	br.closed = true
	return nil
}

// IsDebug builds the debug info
func IsDebug() bool {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return false
	}
	return strings.HasSuffix(file, "_test.go") && line > 0
}
