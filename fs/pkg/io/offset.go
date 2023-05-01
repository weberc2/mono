package io

import (
	"fmt"

	. "github.com/weberc2/mono/fs/pkg/types"
)

type OffsetReadAt struct {
	inner  ReadAt
	offset Byte
}

func NewOffsetReadAt(inner ReadAt, offset Byte) *OffsetReadAt {
	return &OffsetReadAt{inner: inner, offset: offset}
}

func (r *OffsetReadAt) ReadAt(offset Byte, b []byte) error {
	if err := r.inner.ReadAt(offset+r.offset, b); err != nil {
		return fmt.Errorf(
			"reading additional offset `%d` from base offset `%d` (total "+
				"offset `%d` bytes): %w",
			offset,
			r.offset,
			offset+r.offset,
			err,
		)
	}
	return nil
}

type OffsetWriteAt struct {
	inner  WriteAt
	offset Byte
}

func NewOffsetWriteAt(inner WriteAt, offset Byte) *OffsetWriteAt {
	return &OffsetWriteAt{inner: inner, offset: offset}
}

func (r *OffsetWriteAt) WriteAt(offset Byte, b []byte) error {
	if err := r.inner.WriteAt(offset+r.offset, b); err != nil {
		return fmt.Errorf(
			"writing additional offset `%d` from base offset `%d` (total "+
				"offset `%d` bytes): %w",
			offset,
			r.offset,
			offset+r.offset,
			err,
		)
	}
	return nil
}

type OffsetVolume struct {
	inner  Volume
	offset Byte
}

func NewVolume(inner Volume, offset Byte) *OffsetVolume {
	return &OffsetVolume{inner: inner, offset: offset}
}

func (v *OffsetVolume) ReadAt(offset Byte, b []byte) error {
	return NewOffsetReadAt(v.inner, v.offset).ReadAt(offset, b)
}

func (v *OffsetVolume) WriteAt(offset Byte, b []byte) error {
	return NewOffsetWriteAt(v.inner, v.offset).WriteAt(offset, b)
}
