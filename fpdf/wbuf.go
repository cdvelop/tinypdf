package fpdf

type wbuffer struct {
	p []byte
	c int
}

func (w *wbuffer) u8(v uint8) {
	w.p[w.c] = v
	w.c++
}

func (w *wbuffer) bytes() []byte { return w.p }
