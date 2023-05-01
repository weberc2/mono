package indirect

import (
	"fmt"

	"github.com/weberc2/mono/fs/pkg/encode"
	"github.com/weberc2/mono/fs/pkg/io"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type Reader struct {
	readAt io.ReadAt
}

func NewReader(inner io.ReadAt) Reader {
	return Reader{inner}
}

func (r Reader) ReadIndirect(
	indirect Block,
	index Index,
) (Block, error) {
	buf := new([BlockPointerSize]byte)
	if err := r.readAt.ReadAt(offset(indirect, index), buf[:]); err != nil {
		return 0, fmt.Errorf(
			"reading indirect block `%d` at index `%d`: %w",
			indirect,
			index,
			err,
		)
	}
	return encode.DecodeBlock(buf), nil
}
