package main

import (
	"fmt"
	"os"
)

type Volume interface {
	Read(offset uint64, buffer []byte) error
	Write(offset uint64, buffer []byte) error
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
