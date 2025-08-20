package fontManager

// osFile is a minimal interface for parsing TTF without depending on the io package
type osFile interface {
	Read([]byte) (int, error)
	Seek(offset int64, whence int) (int64, error)
}

// Local seek constants matching the standard library's io package so
// callers do not need to import io just for Seek whence values.
const (
	SeekStart   = 0 // io.SeekStart
	SeekCurrent = 1 // io.SeekCurrent
	SeekEnd     = 2 // io.SeekEnd
)
