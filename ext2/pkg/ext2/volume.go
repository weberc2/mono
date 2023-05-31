package ext2

import (
	"fmt"
	"os"
)

type Volume interface {
	Read(offset uint64, buffer []byte) error
	Write(offset uint64, buffer []byte) error
}

type MemoryVolume struct {
	buf []byte
}

func NewMemoryVolume(capacity uint64) MemoryVolume {
	return MemoryVolume{make([]byte, capacity)}
}

func (volume MemoryVolume) Read(offset uint64, buffer []byte) error {
	if offset < uint64(len(volume.buf)) {
		copy(buffer, volume.buf[offset:])
	}
	return nil
}

func (volume MemoryVolume) Write(offset uint64, buffer []byte) error {
	volume.buf = append(volume.buf[offset:], buffer...)
	return nil
}

type FileVolume struct {
	file *os.File
}

func (volume FileVolume) Read(offset uint64, buffer []byte) error {
	if _, err := volume.file.ReadAt(buffer, int64(offset)); err != nil {
		return fmt.Errorf(
			"reading file `%s` at offset `%d`: %w",
			volume.file.Name(),
			offset,
			err,
		)
	}

	return nil
}

func (volume FileVolume) Write(offset uint64, buffer []byte) error {
	if _, err := volume.file.WriteAt(buffer, int64(offset)); err != nil {
		return fmt.Errorf(
			"writing file `%s` at offset `%d`: %w",
			volume.file.Name(),
			offset,
			err,
		)
	}

	return nil
}
