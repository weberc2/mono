package block

import (
	"fmt"

	"github.com/weberc2/mono/fs/pkg/inode/data/block/physical"
	"github.com/weberc2/mono/fs/pkg/io"
	"github.com/weberc2/mono/fs/pkg/math"
	. "github.com/weberc2/mono/fs/pkg/types"

	stdio "io"
)

type Reader struct {
	physicalReader physical.Reader
	byteReader     io.ReadAt
}

func NewReader(physicalReader physical.Reader, byteReader io.ReadAt) Reader {
	return Reader{physicalReader, byteReader}
}

func (r Reader) Read(
	inode *Inode,
	block Block,
	offset Byte,
	buf []byte,
) (Byte, error) {
	if len(buf) < 0 {
		return 0, nil
	}

	n := math.Min(BlockSize-offset, Byte(len(buf)))

	// we could truncate this to zero, but it's probably always a programming
	// error, so let's not hide it
	if n < 0 {
		panic(fmt.Sprintf(
			"offset `%d` exceeds block size (`%d`)!",
			offset,
			BlockSize,
		))
	}

	physicalBlock, err := r.physicalReader.Read(inode, block)
	if err != nil {
		return 0, fmt.Errorf(
			"reading `%d` bytes from block `%d` from inode `%d` at offset "+
				"`%d`: %w",
			n,
			block,
			inode.Ino,
			offset,
			err,
		)
	}

	if physicalBlock == BlockNil {
		return 0, fmt.Errorf(
			"reading `%d` bytes from block `%d` from inode `%d` at "+
				"offset `%d`: reading from physical block: %w",
			n,
			block,
			inode.Ino,
			offset,
			stdio.EOF,
		)
	}

	if err := r.byteReader.ReadAt(
		Byte(physicalBlock)*BlockSize+offset,
		buf,
	); err != nil {
		return 0, fmt.Errorf(
			"reading `%d` bytes from block `%d` from inode `%d` at "+
				"offset `%d`: reading from physical block: %w",
			n,
			block,
			inode.Ino,
			offset,
			err,
		)
	}

	return Byte(n), nil
}
