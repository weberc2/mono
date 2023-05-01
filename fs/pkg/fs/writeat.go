package fs

import (
	"fmt"
	"io"
)

func WriteAt(volume io.WriteSeeker, offset Byte, p []byte) error {
	_, err := (&offsetWriter{volume, offset}).Write(p)
	return err
}

type offsetWriter struct {
	w      io.WriteSeeker
	offset Byte
}

func (writer *offsetWriter) Write(data []byte) (int, error) {
	if _, err := writer.w.Seek(
		int64(writer.offset),
		io.SeekStart,
	); err != nil {
		return 0, fmt.Errorf("seeking to `%d`: %w", writer.offset, err)
	}

	n, err := writer.w.Write(data)
	if err != nil {
		return 0, fmt.Errorf("writing at `%d`: %w", writer.offset, err)
	}
	return n, nil
}
