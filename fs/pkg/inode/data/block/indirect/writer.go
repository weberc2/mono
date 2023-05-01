package indirect

import (
	"fmt"

	"github.com/weberc2/mono/fs/pkg/encode"
	"github.com/weberc2/mono/fs/pkg/io"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type Writer struct {
	writeAt io.WriteAt
}

func NewWriter(inner io.WriteAt) Writer {
	return Writer{inner}
}

func (w Writer) WriteIndirect(
	indirect Block,
	index Index,
	target Block,
) error {
	buf := new([BlockPointerSize]byte)
	encode.EncodeBlock(target, buf)
	if err := w.writeAt.WriteAt(offset(indirect, index), buf[:]); err != nil {
		return fmt.Errorf(
			"writing target block `%d` indirect block `%d` at index `%d`: %w",
			target,
			indirect,
			index,
			err,
		)
	}
	return nil
}
