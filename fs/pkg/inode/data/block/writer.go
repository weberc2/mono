package block

import (
	"fmt"
	stdio "io"

	"github.com/weberc2/mono/fs/pkg/inode/data/block/physical"
	"github.com/weberc2/mono/fs/pkg/io"
	"github.com/weberc2/mono/fs/pkg/math"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type Writer struct {
	physicalReadWriter physical.ReadWriter
	byteWriter         io.WriteAt
}

func NewWriter(physical physical.ReadWriter, byteWriter io.WriteAt) Writer {
	return Writer{physicalReadWriter: physical, byteWriter: byteWriter}
}

func (w *Writer) Write(
	inode *Inode,
	block Block,
	offset Byte,
	buf []byte,
) (Byte, error) {
	if len(buf) < 1 {
		return 0, nil
	}

	n := math.Min(BlockSize-offset, Byte(len(buf)))

	physicalBlock, err := w.physicalReadWriter.ReadAlloc(inode, block)
	if err != nil {
		return 0, fmt.Errorf(
			"reading `%d` bytes from inode `%d` block `%d` at offset `%d`: "+
				"reading physical block: %w",
			n,
			inode.Ino,
			block,
			offset,
			err,
		)
	}

	if err := w.byteWriter.WriteAt(
		Byte(physicalBlock)*BlockSize+offset,
		buf,
	); err != nil {
		return 0, fmt.Errorf(
			"reading `%d` bytes from inode `%d` block `%d` at offset `%d`: "+
				"writing to physical block: %w",
			n,
			inode.Ino,
			block,
			offset,
			stdio.EOF,
		)
	}

	return n, nil
}
