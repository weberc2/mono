package data

import (
	"github.com/weberc2/mono/fs/pkg/inode/data/block"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type ReadWriter struct {
	blockReadWriter block.ReadWriter
	inodeStore      InodeStore
}

func NewReadWriter(blocks block.ReadWriter, inodeStore InodeStore) ReadWriter {
	return ReadWriter{blocks, inodeStore}
}

func (rw *ReadWriter) Reader() Reader {
	return NewReader(rw.blockReadWriter.Reader())
}

func (rw *ReadWriter) Writer() Writer {
	return NewWriter(rw.blockReadWriter.Writer(), rw.inodeStore)
}

func (rw *ReadWriter) Read(inode *Inode, offset Byte, b []byte) (Byte, error) {
	r := NewReader(rw.blockReadWriter.Reader())
	return r.Read(inode, offset, b)
}

func (rw *ReadWriter) Write(
	inode *Inode,
	offset Byte,
	b []byte,
) (Byte, error) {
	w := NewWriter(rw.blockReadWriter.Writer(), rw.inodeStore)
	return w.Write(inode, offset, b)
}
