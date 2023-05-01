package fs

import (
	"io"
)

type Buffer struct {
	data   []byte
	cursor int
}

func NewBuffer(data []byte) *Buffer { return &Buffer{data: data} }

func (b *Buffer) Write(p []byte) (int, error) {
	if int(b.cursor)+len(p) > len(b.data) {
		// if the buffer needs to grow, then we'll append p after the cursor
		b.data = append(b.data[:b.cursor], p...)
	} else {
		// otherwise we'll copy p after b.cursor in order to preserve any data
		// in the buffer following p
		copy(b.data[b.cursor:b.cursor+len(p)], p)
	}
	b.cursor += len(p)
	return len(p), nil
}

func (b *Buffer) Read(p []byte) (int, error) {
	if b.cursor >= len(b.data) {
		if len(b.data) == 0 {
			return 0, nil
		}
		return 0, io.EOF
	}
	n := copy(p, b.data[b.cursor:])
	b.cursor += n
	return n, nil
}

func (b *Buffer) Seek(offset int64, whence int) (int64, error) {
	var relativeTo int
	if whence == io.SeekCurrent {
		relativeTo = b.cursor
	} else if whence == io.SeekEnd {
		relativeTo = len(b.data)
	}
	b.cursor = Max(0, relativeTo+int(offset))
	if remainder := b.cursor - len(b.data); remainder > 0 {
		b.data = append(b.data, make([]byte, remainder)...)
	}
	return int64(b.cursor), nil
}

func (b *Buffer) Bytes() []byte { return b.data }

func (b *Buffer) Len() int { return len(b.data) }
