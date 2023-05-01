package fs

import (
	"fmt"
	"io"
)

func ReadAt(volume io.ReadSeeker, offset Byte, buf []byte) error {
	_, err := (&offsetReader{volume, offset}).Read(buf)
	return err
}

type offsetReader struct {
	r      io.ReadSeeker
	offset Byte
}

func (reader *offsetReader) Read(buf []byte) (int, error) {
	if _, err := reader.r.Seek(
		int64(reader.offset),
		io.SeekStart,
	); err != nil {
		return 0, fmt.Errorf("seeking to `%d`: %w", reader.offset, err)
	}

	n, err := reader.r.Read(buf)
	if err != nil {
		return n, fmt.Errorf("reading at `%d`: %w", reader.offset, err)
	}
	return n, nil
}
