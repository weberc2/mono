package indirect

import (
	"github.com/weberc2/mono/fs/pkg/io"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type ReadWriter struct {
	volume io.Volume
}

func NewReadWriter(volume io.Volume) ReadWriter {
	return ReadWriter{volume}
}

func (rw ReadWriter) ReadIndirect(
	indirect Block,
	index Index,
) (Block, error) {
	return rw.Reader().ReadIndirect(indirect, index)
}

func (rw ReadWriter) WriteIndirect(
	indirect Block,
	index Index,
	target Block,
) error {
	return Writer{rw.volume}.WriteIndirect(indirect, index, target)
}

func (rw ReadWriter) Reader() Reader {
	return Reader{rw.volume}
}

func (rw ReadWriter) Writer() Writer {
	return Writer{rw.volume}
}
