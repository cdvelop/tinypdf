//go:build !wasm

package fpdf

import "time"

// SetDefaultCreationDate sets the default value of the document creation date
// that will be used when initializing a new Fpdf instance. See
// SetCreationDate() for more details.
func SetDefaultCreationDate(tm time.Time) {
	gl.creationDate = pdfTime(tm)
}

// SetDefaultModificationDate sets the default value of the document modification date
// that will be used when initializing a new Fpdf instance. See
// SetCreationDate() for more details.
func SetDefaultModificationDate(tm time.Time) {
	gl.modDate = pdfTime(tm)
}

// GetCreationDate returns the document's internal CreationDate value.
func (f *Fpdf) GetCreationDate() time.Time {
	return time.Time(f.creationDate)
}

// SetCreationDate fixes the document's internal CreationDate value. By
// default, the time when the document is generated is used for this value.
// This method is typically only used for testing purposes to facilitate PDF
// comparison. Specify a zero-value time to revert to the default behavior.
func (f *Fpdf) SetCreationDate(tm time.Time) {
	f.creationDate = pdfTime(tm)
}

// GetModificationDate returns the document's internal ModDate value.
func (f *Fpdf) GetModificationDate() time.Time {
	return time.Time(f.modDate)
}

// SetModificationDate fixes the document's internal ModDate value.
// See `SetCreationDate` for more details.
func (f *Fpdf) SetModificationDate(tm time.Time) {
	f.modDate = pdfTime(tm)
}

// returns Now() if tm is zero
func timeOrNow(tm pdfTime) time.Time {
	t := time.Time(tm)
	if t.IsZero() {
		return time.Now()
	}
	return t
}

func formatPDFDate(tm pdfTime) string {
	t := timeOrNow(tm)
	return "D:" + t.Format("20060102150405")
}
