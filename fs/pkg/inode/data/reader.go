package data

import (
	"fmt"

	"github.com/weberc2/mono/fs/pkg/inode/data/block"
	"github.com/weberc2/mono/fs/pkg/math"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type Reader struct {
	blockReader block.Reader
}

func (r *Reader) Read(inode *Inode, offset Byte, b []byte) (Byte, error) {
	maxLength := math.Min(Byte(len(b)), inode.Size-offset)
	var chunkBegin Byte = 0

	for chunkBegin < maxLength {
		chunkBlock := Block((offset + chunkBegin) / BlockSize)
		chunkOffset := (offset + chunkBegin) % BlockSize
		chunkLength := math.Min(maxLength-chunkBegin, BlockSize-chunkOffset)

		actual, err := r.blockReader.ReadBlock(
			inode,
			chunkBlock,
			chunkOffset,
			b[chunkBegin:chunkBegin+chunkLength],
		)
		if err != nil {
			return chunkBegin, fmt.Errorf(
				"reading up to `%d` bytes from inode `%d` at offset `%d`: %w",
				len(b),
				inode.Ino,
				offset,
				err,
			)
		}
		if actual != chunkLength {
			panic(fmt.Sprintf(
				"intended to read `%d` bytes; actually read `%d` bytes",
				chunkLength,
				actual,
			))
		}

		chunkBegin += chunkLength
	}

	return chunkBegin, nil
}
