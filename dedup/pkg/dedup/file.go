package dedup

import (
	"errors"
	"fmt"
	"hash/adler32"
	"io"
	"os"
)

type File struct {
	Path               string
	Size               int64
	Ino                uint64
	FirstBlockChecksum uint32
	FinalBlockChecksum uint32
}

func (f *File) ChecksumBoundingBlocks() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf(
				"checksumming first/last blocks for file `%s`: %w",
				f.Path,
				err,
			)
		}
	}()

	var file *os.File
	if file, err = os.Open(f.Path); err != nil {
		return
	}
	defer func() { err = errors.Join(err, file.Close()) }()

	var buf [blockSize]byte
	var n int
	if n, err = file.Read(buf[:]); err != nil {
		err = fmt.Errorf("reading first block: %w", err)
		return
	}
	f.FirstBlockChecksum = adler32.Checksum(buf[:n])

	offset := f.Size - blockSize
	if offset < 0 {
		offset = 0
	}
	if _, err = file.Seek(offset, io.SeekStart); err != nil {
		err = fmt.Errorf("seeking to final block: %w", err)
		return
	}

	if n, err = file.Read(buf[:]); err != nil {
		err = fmt.Errorf("reading final block: %w", err)
		return
	}
	f.FinalBlockChecksum = adler32.Checksum(buf[:n])
	return
}

const blockSize = 1024
