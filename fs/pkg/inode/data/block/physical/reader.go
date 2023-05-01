package physical

import (
	"fmt"

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

func (r Reader) ReadPhysical(inode *Inode, inodeBlock Block) (Block, error) {
	var ind indirection
	ind.fromBlock(inodeBlock)

	block, err := func() (Block, error) {
		switch ind.level {
		case levelDirect:
			return inode.DirectBlocks[ind.direct], nil
		case levelSingly:
			return r.singlyIndirect(inode.SinglyIndirectBlock, ind.singly)
		case levelDoubly:
			return r.doublyIndirect(
				inode.DoublyIndirectBlock,
				ind.singly,
				ind.doubly,
			)
		case levelTriply:
			return r.triplyIndirect(
				inode.TriplyIndirectBlock,
				ind.singly,
				ind.doubly,
				ind.triply,
			)
		case levelOutOfRange:
			return BlockNil, OutOfRangeErr
		}
		panic(fmt.Sprintf("invalid indirection level: %d", ind.level))
	}()

	if err != nil {
		return BlockNil, fmt.Errorf(
			"reading physical block for inode `%d`, block `%d`: %w",
			inode.Ino,
			inodeBlock,
			err,
		)
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
