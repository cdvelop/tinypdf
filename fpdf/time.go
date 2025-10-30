package tinypdf

import "time"

// SetDefaultCreationDate sets the default value of the document creation date
// that will be used when initializing a new Fpdf instance. See
// SetCreationDate() for more details.
func SetDefaultCreationDate(tm time.Time) {
	gl.creationDate = tm
}

// SetDefaultModificationDate sets the default value of the document modification date
// that will be used when initializing a new Fpdf instance. See
// SetCreationDate() for more details.
func SetDefaultModificationDate(tm time.Time) {
	gl.modDate = tm
}

// GetCreationDate returns the document's internal CreationDate value.
func (f *Fpdf) GetCreationDate() time.Time {
	return f.creationDate
}

// SetCreationDate fixes the document's internal CreationDate value. By
// default, the time when the document is generated is used for this value.
// This method is typically only used for testing purposes to facilitate PDF
// comparison. Specify a zero-value time to revert to the default behavior.
func (f *Fpdf) SetCreationDate(tm time.Time) {
	f.creationDate = tm
}

// GetModificationDate returns the document's internal ModDate value.
func (f *Fpdf) GetModificationDate() time.Time {
	return f.modDate
}

// SetModificationDate fixes the document's internal ModDate value.
// See `SetCreationDate` for more details.
func (f *Fpdf) SetModificationDate(tm time.Time) {
	f.modDate = tm
}

// returns Now() if tm is zero
func timeOrNow(tm time.Time) time.Time {
	if tm.IsZero() {
		return time.Now()
	}
	return tm
}
