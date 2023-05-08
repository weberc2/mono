package physical

import (
	"fmt"

	"github.com/weberc2/mono/fs/pkg/alloc"
	"github.com/weberc2/mono/fs/pkg/inode/data/block/indirect"
	. "github.com/weberc2/mono/fs/pkg/types"
)

const (
	OutOfBlocksErr ConstError = "out of free blocks"
)

type ReadWriter struct {
	allocator  alloc.BlockAllocator
	inodeStore InodeStore
	indirects  indirect.ReadWriter
}

func NewReadWriter(
	allocator alloc.BlockAllocator,
	indirects indirect.ReadWriter,
	inodeStore InodeStore,
) ReadWriter {
	return ReadWriter{
		allocator:  allocator,
		indirects:  indirects,
		inodeStore: inodeStore,
	}
}

func (rw *ReadWriter) Reader() Reader {
	return Reader{rw.indirects.Reader()}
}

func (rw *ReadWriter) Read(inode *Inode, block Block) (Block, error) {
	return Reader{rw.indirects.Reader()}.Read(inode, block)
}

func (rw *ReadWriter) ReadAlloc(inode *Inode, indirect Block) (Block, error) {
	var ind indirection
	if err := ind.fromInodeBlock(inode, indirect); err != nil {
		return BlockNil, fmt.Errorf(
			"getting physical block for inode `%d`, block `%d`: %w",
			inode.Ino,
			indirect,
			err,
		)
	}

	// if the inode points to an invalid block, allocate a valid block, stick it
	// on the inode (store the updated inode).
	if err := rw.ensureToplevel(inode, &ind); err != nil {
		return BlockNil, fmt.Errorf(
			"getting physical block for inode `%d`, block `%d`: %w",
			inode.Ino,
			indirect,
			err,
		)
	}

	// now that the inode definitely points to a valid block, descend into any
	// indirects.
	p, err := rw.readIndirect(*ind.ptr, ind.indices())
	if err != nil {
		return BlockNil, fmt.Errorf(
			"getting physical block for inode `%d`, block `%d`: "+
				"traversing %s block: %w",
			inode.Ino,
			indirect,
			ind.level,
			err,
		)
	}
	return p, nil
}

func (rw *ReadWriter) Write(
	inode *Inode,
	indirect Block,
	physical Block,
) error {
	var ind indirection
	if err := ind.fromInodeBlock(inode, indirect); err != nil {
		return fmt.Errorf(
			"mapping inode `%d` block `%d` to physical block `%d`: "+
				"storing updated inode: %w",
			inode.Ino,
			indirect,
			physical,
			err,
		)
	}

	if ind.level == levelDirect {
		*ind.ptr = physical
		if err := rw.inodeStore.Put(inode); err != nil {
			return fmt.Errorf(
				"mapping inode `%d` block `%d` to physical block `%d`: "+
					"storing updated inode: %w",
				inode.Ino,
				indirect,
				physical,
				err,
			)
		}
		return nil
	}

	if err := rw.ensureToplevel(inode, &ind); err != nil {
		return fmt.Errorf(
			"mapping inode `%d` block `%d` to physical block `%d`: %w",
			inode.Ino,
			indirect,
			physical,
			err,
		)
	}

	// since this isn't a direct block, we are guaranteed to have at least one
	// index. Since `indices()` returns a slice in ascending order (singly
	// indirect index comes first), we have to iterate backwards (start with
	// the most indirect index).
	indices := ind.indices()
	index := indices[len(indices)-1]
	nextIndices := indices[:len(indices)-1]

	if err := rw.writeIndirect(
		indirect,
		physical,
		index,
		nextIndices,
	); err != nil {
		return fmt.Errorf(
			"mapping inode `%d` block `%d` to physical block `%d`: %w",
			inode.Ino,
			indirect,
			physical,
			err,
		)
	}

	return nil
}

func (rw *ReadWriter) writeIndirect(
	indirect Block,
	physical Block,
	index indirect.Index,
	nextIndices []indirect.Index, // ordered from singly -> triply
) error {
	if len(nextIndices) < 1 {
		if err := rw.indirects.WriteIndirect(
			indirect,
			index,
			physical,
		); err != nil {
			return err
		}
		return nil
	}

	nextBlock, err := rw.indirects.ReadIndirect(indirect, index)
	if err != nil {
		return err
	}

	return rw.writeIndirect(
		nextBlock,
		physical,
		nextIndices[len(nextIndices)-1],
		nextIndices[:len(nextIndices)-1],
	)
}

func (rw *ReadWriter) ensureToplevel(inode *Inode, ind *indirection) error {
	if *ind.ptr == BlockNil {
		var err error
		*ind.ptr, err = rw.allocOne()
		if err != nil {
			return fmt.Errorf("allocating %s block: %w", ind.level, err)
		}
		if err := rw.inodeStore.Put(inode); err != nil {
			rw.allocator.Free(*ind.ptr)
			*ind.ptr = BlockNil
			return fmt.Errorf(
				"allocating %s block: storing updated inode: %w",
				ind.level,
				err,
			)
		}
	}
	return nil
}

func (rw *ReadWriter) readIndirect(
	b Block,
	indices []indirect.Index, // ordered from singly -> triply
) (Block, error) {
	// if there are no indices, then `b` is the physical block; return it.
	if len(indices) < 1 {
		return b, nil
	}

	index := indices[len(indices)-1]
	indices = indices[:len(indices)-1]
	nextBlock, err := rw.indirects.ReadIndirect(b, index)
	if err != nil {
		return BlockNil, fmt.Errorf(
			"reading block pointer at index `%d` from block `%d`",
			index,
			b,
		)
	}

	if nextBlock != BlockNil {
		physical, err := rw.readIndirect(nextBlock, indices)
		if err != nil {
			return BlockNil, fmt.Errorf(
				"block `%d`, index `%d`: %w",
				b,
				index,
				err,
			)
		}
		return physical, nil
	}

	// if `nextBlock` is nil, then we need to allocate it and any blocks beneath
	// it. We also need to persist references to any allocated blocks. E.g., if
	// `nextBlock` is meant to be the singly indirect block, then we need to
	// allocate the singly indirect block and the physical block and we also
	// need to write the singly indirect block pointer into the doubly indirect
	// block and we need to write the physical block pointer into the singly
	// indirect block.
	physical, err := rw.allocAll(b, index, indices)
	if err != nil {
		return BlockNil, fmt.Errorf(
			"block `%d`, index `%d`: %w",
			b,
			index,
			err,
		)
	}

	return physical, nil
}

func (rw *ReadWriter) allocAll(
	outerBlock Block,
	outerIndex indirect.Index,
	indices []indirect.Index,
) (Block, error) {
	// allocate a new block to be pointed to by (outerBlock, outerIndex)
	b, err := rw.allocOne()
	if err != nil {
		return BlockNil, fmt.Errorf(
			"allocating block to store in (block `%d`, index `%d`): %w",
			outerBlock,
			outerIndex,
			err,
		)
	}

	// write the newly-allocated block pointer to the parent block
	if err := rw.indirects.WriteIndirect(
		outerBlock,
		outerIndex,
		b,
	); err != nil {
		// free the block if we can't persist a pointer to it in the outer
		// block.
		rw.allocator.Free(b)
		return BlockNil, fmt.Errorf(
			"writing newly-allocated block pointer `%d` to parent block `%d` "+
				"at index `%d`: %w",
			b,
			outerBlock,
			outerIndex,
			err,
		)
	}

	// if there are more indices, recurse.
	// NB: We are allocating the outer nodes and persisting the references
	// *before* recursing so we don't leak blocks on error (if a block is
	// allocated, it should always be pointed to by another block).
	if len(indices) > 0 {
		return rw.allocAll(
			b,
			indices[len(indices)-1],
			indices[:len(indices)-1],
		)
	}

	// otherwise `b` is the physical block; return it
	return b, nil
}

func (rw *ReadWriter) allocOne() (Block, error) {
	b, ok := rw.allocator.Alloc()
	if !ok {
		return BlockNil, OutOfBlocksErr
	}
	return b, nil
}
