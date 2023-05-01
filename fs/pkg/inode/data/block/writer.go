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
	physicalReader physical.Reader
	byteWriter     io.WriteAt
}

func NewWriter(physicalReader physical.Reader, byteWriter io.WriteAt) Writer {
	return Writer{physicalReader: physicalReader, byteWriter: byteWriter}
}

func (w *Writer) WriteBlock(
	inode *Inode,
	block Block,
	offset Byte,
	buf []byte,
) (Byte, error) {
	if len(buf) < 1 {
		return 0, nil
	}

	n := math.Min(BlockSize-offset, Byte(len(buf)))

	physicalBlock, err := w.physicalReader.ReadPhysical(inode, block)
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

	if physicalBlock == BlockNil {
		return 0, fmt.Errorf(
			"reading `%d` bytes from inode `%d` block `%d` at offset `%d`: "+
				"physical block is invalid: %w",
			n,
			inode.Ino,
			block,
			offset,
			stdio.EOF,
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
