package physical

import (
	"fmt"
	"strings"

	"github.com/weberc2/mono/fs/pkg/inode/data/block/indirect"
	. "github.com/weberc2/mono/fs/pkg/types"
)

const (
	OutOfRangeErr ConstError = "block out of range"
)

type Reader struct {
	indirectReader indirect.Reader
}

func NewReader(indirectReader indirect.Reader) Reader {
	return Reader{indirectReader}
}

func (r Reader) Read(inode *Inode, inodeBlock Block) (Block, error) {
	var ind indirection
	if err := ind.fromInodeBlock(inode, inodeBlock); err != nil {
		return BlockNil, fmt.Errorf(
			"reading physical block for inode `%d`, block `%d`: %w",
			inode.Ino,
			inodeBlock,
			err,
		)
	}

	type errctx struct {
		b Block
		i indirect.Index
	}

	block := *ind.ptr
	ctx := []errctx{}
	for _, index := range ind.indices() {
		ctx = append(ctx, errctx{block, index})
		var err error
		block, err = r.indirectReader.ReadIndirect(block, index)
		if err != nil {
			var sb strings.Builder
			fmt.Fprintf(
				&sb,
				"reading physical block for inode `%d`, block `%d`",
				inode.Ino,
				inodeBlock,
			)
			for _, c := range ctx {
				fmt.Fprintf(&sb, ": reading block `%d`, index `%d`", c.b, c.i)
			}

			return BlockNil, fmt.Errorf("%s: %w", &sb, err)
		}
	}

	return block, nil
}

func (r Reader) singlyIndirect(
	ind Block,
	singlyIndirectIndex indirect.Index,
) (Block, error) {
	if ind == 0 {
		return BlockNil, nil
	}
	block, err := r.indirectReader.ReadIndirect(ind, singlyIndirectIndex)
	if err != nil {
		return BlockNil, fmt.Errorf(
			"reading singly indirect block `%d` at index `%d`: %w",
			ind,
			singlyIndirectIndex,
			err,
		)
	}
	return block, nil
}

func (r Reader) doublyIndirect(
	ind Block,
	singlyIndirectIndex indirect.Index,
	doublyIndirectIndex indirect.Index,
) (Block, error) {
	if ind == 0 {
		return BlockNil, nil
	}
	singlyIndirectBlock, err := r.indirectReader.ReadIndirect(
		ind,
		doublyIndirectIndex,
	)
	if err != nil {
		return BlockNil, fmt.Errorf(
			"reading doubly indirect block `%d` at index `%d`: %w",
			ind,
			doublyIndirectIndex,
			err,
		)
	}

	block, err := r.singlyIndirect(singlyIndirectBlock, singlyIndirectIndex)
	if err != nil {
		return BlockNil, fmt.Errorf(
			"reading doubly indirect block `%d` at index `%d`: %w",
			ind,
			doublyIndirectIndex,
			err,
		)
	}

	return block, nil
}

func (r Reader) triplyIndirect(
	ind Block,
	singlyIndirectIndex indirect.Index,
	doublyIndirectIndex indirect.Index,
	triplyIndirectIndex indirect.Index,
) (Block, error) {
	if ind == 0 {
		return BlockNil, nil
	}
	doublyIndirectBlock, err := r.indirectReader.ReadIndirect(
		ind,
		triplyIndirectIndex,
	)
	if err != nil {
		return BlockNil, fmt.Errorf(
			"reading triply indirect block `%d` at index `%d`: %w",
			ind,
			triplyIndirectIndex,
			err,
		)
	}

	block, err := r.doublyIndirect(
		doublyIndirectBlock,
		singlyIndirectIndex,
		doublyIndirectIndex,
	)
	if err != nil {
		return BlockNil, fmt.Errorf(
			"reading triply indirect block `%d` at index `%d`: %w",
			ind,
			triplyIndirectIndex,
			err,
		)
	}

	return block, nil
}
