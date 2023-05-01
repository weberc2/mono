package block

import (
	"github.com/weberc2/mono/fs/pkg/inode/data/block/physical"
	"github.com/weberc2/mono/fs/pkg/io"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type ReadWriter struct {
	physicalReader physical.Reader
	volume         io.Volume
}

func (rw *ReadWriter) ReadBlock(
	inode *Inode,
	block Block,
	offset Byte,
	buf []byte,
) (Byte, error) {
	r := NewReader(rw.physicalReader, rw.volume)
	return r.ReadBlock(inode, block, offset, buf)
}

func (rw *ReadWriter) WriteBlock(
	inode *Inode,
	block Block,
	offset Byte,
	buf []byte,
) (Byte, error) {
	w := NewWriter(rw.physicalReader, rw.volume)
	return w.WriteBlock(inode, block, offset, buf)
}
