package block

import (
	"github.com/weberc2/mono/fs/pkg/inode/data/block/physical"
	"github.com/weberc2/mono/fs/pkg/io"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type ReadWriter struct {
	physicalReadWriter physical.ReadWriter
	volume             io.Volume
}

func NewReadWriter(
	readWriter physical.ReadWriter,
	volume io.Volume,
) ReadWriter {
	return ReadWriter{readWriter, volume}
}

func (rw *ReadWriter) Reader() Reader {
	return NewReader(rw.physicalReadWriter.Reader(), rw.volume)
}

func (rw *ReadWriter) Writer() Writer {
	return NewWriter(rw.physicalReadWriter, rw.volume)
}

func (rw *ReadWriter) Read(
	inode *Inode,
	block Block,
	offset Byte,
	buf []byte,
) (Byte, error) {
	r := NewReader(rw.physicalReadWriter.Reader(), rw.volume)
	return r.Read(inode, block, offset, buf)
}

func (rw *ReadWriter) Write(
	inode *Inode,
	block Block,
	offset Byte,
	buf []byte,
) (Byte, error) {
	w := NewWriter(rw.physicalReadWriter, rw.volume)
	return w.Write(inode, block, offset, buf)
}
