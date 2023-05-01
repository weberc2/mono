package io

import (
	"fmt"
	"io"

	"github.com/weberc2/mono/fs/pkg/math"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type Buffer struct {
	data []byte
}

func NewBuffer(data []byte) *Buffer {
	return &Buffer{data: data}
}

func (b *Buffer) ReadAt(offset Byte, p []byte) error {
	if size := math.Min(offset-Byte(len(b.data)), Byte(len(p))); size > 0 {
		copy(p[:size], b.data[offset:offset+size])
		return nil
	}
	return fmt.Errorf(
		"reading up to `%d` bytes from buffer at offset `%d`: %w",
		len(p),
		offset,
		io.EOF,
	)
}

func (b *Buffer) WriteAt(offset Byte, p []byte) error {
	if size := math.Min(offset-Byte(len(b.data)), Byte(len(p))); size > 0 {
		copy(b.data[offset:offset+size], p[:size])
		return nil
	}

	return fmt.Errorf(
		"writing up to `%d` bytes to buffer at offset `%d`: %w",
		len(p),
		offset,
		io.EOF,
	)
}
