package physical

import (
	"github.com/weberc2/mono/fs/pkg/inode/data/block/indirect"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type ReadWriter struct {
	inner *Writer
}

func NewReadWriter(
	indirectReadWriter indirect.ReadWriter,
	inodeStore InodeStore,
	allocator *Bitmap,
) ReadWriter {
	return ReadWriter{
		inner: NewWriter(indirectReadWriter, inodeStore, allocator),
	}
}

func (rw *ReadWriter) ReadPhysical(
	inode *Inode,
	inodeBlock Block,
) (Block, error) {
	return Reader{
		rw.inner.indirectReadWriter.Reader(),
	}.ReadPhysical(inode, inodeBlock)
}

func (rw *ReadWriter) WritePhysical(
	inode *Inode,
	inodeBlock Block,
	physicalBlock Block,
) error {
	return rw.inner.WritePhysical(inode, inodeBlock, physicalBlock)
}
