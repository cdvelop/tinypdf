package tinypdf

// buff for pdf content
type buff struct {
	pos   int // Cambiado de position a int para evitar conflictos de tipo
	datas []byte
}

// Write : write []byte to buffer
func (b *buff) Write(p []byte) (int, error) {
	for len(b.datas) < b.pos+len(p) {
		b.datas = append(b.datas, 0)
	}
	i := 0
	max := len(p)
	for i < max {
		b.datas[i+b.pos] = p[i]
		i++
	}
	b.pos += i
	return 0, nil
}

// Len : len of buffer
func (b *buff) Len() int {
	return len(b.datas)
}

// Bytes : get bytes
func (b *buff) Bytes() []byte {
	return b.datas
}

// position : get current position
func (b *buff) position() int {
	return b.pos
}

// SetPosition : set current position
func (b *buff) SetPosition(pos int) {
	b.pos = pos
}
