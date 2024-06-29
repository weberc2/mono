// Adapted from
// https://trac.opensubtitles.org/projects/opensubtitles/wiki/HashSourceCodes#GO

package subtitles

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"strconv"
)

func Mediahash(file fs.File) (hash string, err error) {
	var h uint64
	fi, err := file.Stat()
	if err != nil {
		return
	}
	if sz := fi.Size(); sz < ChunkSize {
		err = fmt.Errorf(
			"computing mediahash: file is too small: "+
				"must be at least %d bytes; found %d bytes",
			ChunkSize,
			sz,
		)
		return
	}

	readerAt, ok := file.(io.ReaderAt)
	if !ok {
		panic("MediaHash() does not yet support non-os filesystems")
	}

	// Read head and tail blocks.
	buf := make([]byte, ChunkSize*2)
	err = readChunk(readerAt, 0, buf[:ChunkSize])
	if err != nil {
		err = fmt.Errorf("computing mediahash: %w", err)
		return
	}
	err = readChunk(readerAt, fi.Size()-ChunkSize, buf[ChunkSize:])
	if err != nil {
		err = fmt.Errorf("computing mediahash: %w", err)
		return
	}

	// Convert to uint64, and sum.
	var nums [(ChunkSize * 2) / 8]uint64
	reader := bytes.NewReader(buf)
	err = binary.Read(reader, binary.LittleEndian, &nums)
	if err != nil {
		err = fmt.Errorf("computing mediahash: %w", err)
		return
	}
	for _, num := range nums {
		h += num
	}

	hash = strconv.FormatInt(int64(h+uint64(fi.Size())), 16)
	return
}

// Read a chunk of a file at `offset` so as to fill `buf`.
func readChunk(file io.ReaderAt, offset int64, buf []byte) (err error) {
	n, err := file.ReadAt(buf, offset)
	if err != nil {
		return
	}
	if n != ChunkSize {
		return fmt.Errorf("invalid read %v", n)
	}
	return
}

const (
	ChunkSize = 65536 // 64k
)
