package data

import (
	"fmt"

	"github.com/weberc2/mono/fs/pkg/inode/data/block"
	"github.com/weberc2/mono/fs/pkg/math"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type Writer struct {
	blockWriter block.Writer
	inodeStore  InodeStore
}

func NewWriter(blockWriter block.Writer, inodeStore InodeStore) Writer {
	return Writer{blockWriter, inodeStore}
}

func (w *Writer) Write(inode *Inode, offset Byte, b []byte) (Byte, error) {
	var chunkBegin Byte

	for chunkBegin < Byte(len(b)) {
		chunkBlock := Block(offset + chunkBegin/BlockSize)
		chunkOffset := (offset + chunkBegin) % BlockSize
		chunkLength := math.Min(Byte(len(b)), BlockSize-chunkOffset)

		actual, err := w.blockWriter.Write(
			inode,
			chunkBlock,
			chunkOffset,
			b[chunkBegin:chunkBegin+chunkLength],
		)
		if err != nil {
			return chunkBegin, fmt.Errorf(
				"writing up to `%d` bytes from inode `%d` at offset `%d`: %w",
				len(b),
				inode.Ino,
				offset,
				err,
			)
		}

		if actual != chunkLength {
			panic(fmt.Sprintf(
				"intended to write `%d` bytes; actually wrote `%d` bytes",
				chunkLength,
				actual,
			))
		}

		chunkBegin += chunkLength
	}

	if inode.Size < offset+chunkBegin {
		clone := *inode
		clone.Size = offset + chunkBegin
		if err := w.inodeStore.Put(&clone); err != nil {
			return chunkBegin, fmt.Errorf(
				"writing up to `%d` bytes from inode `%d` at offset `%d`: "+
					"updating inode size: %w",
				len(b),
				inode.Ino,
				offset,
				err,
			)
		}
		*inode = clone
	}

	return chunkBegin, nil
}
