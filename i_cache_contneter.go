package tinypdf

import (
	"io"
)

type ICacheContent interface {
	write(w io.Writer, protection *PDFProtection) error
}
