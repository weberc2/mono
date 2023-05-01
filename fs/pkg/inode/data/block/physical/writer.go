package physical

import (
	"fmt"

	"github.com/weberc2/mono/fs/pkg/inode/data/block/indirect"
	. "github.com/weberc2/mono/fs/pkg/types"
)

const (
	OutOfBlocksErr ConstError = "out of free blocks"
)

type Writer struct {
	indirectReadWriter indirect.ReadWriter
	inodeStore         InodeStore
	allocator          *Bitmap
}

func NewWriter(
	indirectReadWriter indirect.ReadWriter,
	inodeStore InodeStore,
	allocator *Bitmap,
) *Writer {
	return &Writer{
		indirectReadWriter: indirectReadWriter,
		inodeStore:         inodeStore,
		allocator:          allocator,
	}
}

func (w *Writer) WritePhysical(
	inode *Inode,
	inodeBlock Block,
	physicalBlock Block,
) error {
	if err := w.writePhysicalHelper(
		inode,
		inodeBlock,
		physicalBlock,
	); err != nil {
		return fmt.Errorf(
			"writing physical block `%d` for inode `%d` block `%d`: %w",
			physicalBlock,
			inode.Ino,
			inodeBlock,
			err,
		)
	}
	return nil
}

func (w *Writer) writePhysicalHelper(
	inode *Inode,
	inodeBlock Block,
	physicalBlock Block,
) error {
	var (
		clone = *inode
		ind   indirection
		err   error
	)
	ind.fromBlock(inodeBlock)

	switch ind.level {
	case levelDirect:
		inode.DirectBlocks[ind.direct] = physicalBlock
	case levelSingly:
		err = w.singlyIndirect(
			&clone,
			ind.singly,
			physicalBlock,
		)
	case levelDoubly:
		err = w.doublyIndirect(
			&clone,
			ind.singly,
			ind.doubly,
			physicalBlock,
		)
	case levelTriply:
		err = w.triplyIndirect(
			&clone,
			ind.singly,
			ind.doubly,
			ind.triply,
			physicalBlock,
		)
	}

	if err != nil {
		return err
	}

	if clone != *inode {
		if err := w.inodeStore.Put(&clone); err != nil {
			return err
		}
	}

	*inode = clone
	return nil
}

func (w *Writer) singlyIndirect(
	inode *Inode,
	singlyIndirectIndex indirect.Index,
	physicalBlock Block,
) error {
	if inode.SinglyIndirectBlock == BlockNil {
		var err error
		inode.SinglyIndirectBlock, err = w.alloc()
		if err != nil {
			return fmt.Errorf(
				"writing block `%d` to singly indirect index `%d`: "+
					"allocating singly indirect block: %w",
				physicalBlock,
				singlyIndirectIndex,
				err,
			)
		}
	}

	if err := w.indirectReadWriter.WriteIndirect(
		inode.SinglyIndirectBlock,
		singlyIndirectIndex,
		physicalBlock,
	); err != nil {
		return fmt.Errorf(
			"writing block `%d` to singly indirect index `%d`: %w",
			physicalBlock,
			singlyIndirectIndex,
			err,
		)
	}
	return nil
}

func (w *Writer) doublyIndirect(
	inode *Inode,
	singlyIndirectIndex indirect.Index,
	doublyIndirectIndex indirect.Index,
	physicalBlock Block,
) error {
	if err := w.doublyIndirectHelper(
		inode,
		singlyIndirectIndex,
		doublyIndirectIndex,
		physicalBlock,
	); err != nil {
		return fmt.Errorf(
			"writing physical block `%d` for inode `%d` at (doubly indirect "+
				"index `%d`, singly indirect index `%d`): %w",
			physicalBlock,
			inode.Ino,
			doublyIndirectIndex,
			singlyIndirectIndex,
			err,
		)
	}
	return nil
}

func (w *Writer) doublyIndirectHelper(
	inode *Inode,
	singlyIndirectIndex indirect.Index,
	doublyIndirectIndex indirect.Index,
	physicalBlock Block,
) error {
	var singlyIndirectBlock Block

	if inode.DoublyIndirectBlock == BlockNil {
		doublyIndirectBlock, err := w.alloc()
		if err != nil {
			return fmt.Errorf("allocating doubly indirect block: %w", err)
		}
		singlyIndirectBlock, err = w.alloc()
		if err != nil {
			return fmt.Errorf("allocating singly indirect block: %w", err)
		}
		inode.DoublyIndirectBlock = doublyIndirectBlock
	} else {
		var err error
		singlyIndirectBlock, err = w.indirectReadWriter.ReadIndirect(
			inode.DoublyIndirectBlock,
			doublyIndirectIndex,
		)
		if err != nil {
			return fmt.Errorf("reading singly indirect block: %w", err)
		}
		if singlyIndirectBlock == BlockNil {
			singlyIndirectBlock, err = w.alloc()
			if err != nil {
				return fmt.Errorf("allocating singly indirect block: %w", err)
			}
			if err := w.indirectReadWriter.WriteIndirect(
				inode.DoublyIndirectBlock,
				doublyIndirectIndex,
				singlyIndirectBlock,
			); err != nil {
				return fmt.Errorf(
					"writing singly indirect block address to doubly "+
						"indirect block: %w",
					err,
				)
			}
		}
	}

	if err := w.indirectReadWriter.WriteIndirect(
		singlyIndirectBlock,
		singlyIndirectIndex,
		physicalBlock,
	); err != nil {
		return fmt.Errorf(
			"writing physical block to singly indirect block: %w",
			err,
		)
	}

	return nil
}

func (w *Writer) triplyIndirect(
	inode *Inode,
	singlyIndirectIndex indirect.Index,
	doublyIndirectIndex indirect.Index,
	triplyIndirectIndex indirect.Index,
	physicalBlock Block,
) error {
	if err := w.triplyIndirectHelper(
		inode,
		singlyIndirectIndex,
		doublyIndirectIndex,
		triplyIndirectIndex,
		physicalBlock,
	); err != nil {
		return fmt.Errorf(
			"writing physical block `%d` for inode `%d` at (triply indirect "+
				"index `%d`, doubly indirect index `%d`, singly indirect "+
				"index `%d`): %w",
			physicalBlock,
			inode.Ino,
			triplyIndirectIndex,
			doublyIndirectIndex,
			singlyIndirectIndex,
			err,
		)
	}

	return nil
}

func (w *Writer) triplyIndirectHelper(
	inode *Inode,
	singlyIndirectIndex indirect.Index,
	doublyIndirectIndex indirect.Index,
	triplyIndirectIndex indirect.Index,
	physicalBlock Block,
) error {
	var (
		singlyIndirectBlock, doublyIndirectBlock Block
		err                                      error
	)

	// If the inode's triply indirect block is nil, then we need to:
	// * allocate blocks at each level of indirection (triply, doubly, and
	//   singly)
	// * write the physical block address into the singly indirect block
	// * write the newly allocated singly indirect block address into the
	//   doubly indirect block
	// * write the newly allocated doubly indirect block address into the
	//   triply indirect block
	// * write the newly allocated triply indirect block address to the inode's
	//   `TriplyIndirectBlock` field
	if inode.TriplyIndirectBlock == BlockNil {
		inode.TriplyIndirectBlock, err = w.alloc()
		if err != nil {
			return fmt.Errorf("allocating triply indirect block: %w", err)
		}
		doublyIndirectBlock, err = w.alloc()
		if err != nil {
			return fmt.Errorf("allocating doubly indirect block: %w", err)
		}
		singlyIndirectBlock, err = w.alloc()
		if err != nil {
			return fmt.Errorf("allocating singly indirect block: %w", err)
		}

		if err := w.indirectReadWriter.WriteIndirect(
			singlyIndirectBlock,
			singlyIndirectIndex,
			physicalBlock,
		); err != nil {
			return fmt.Errorf(
				"writing physical block address into singly indirect block: "+
					"%w",
				err,
			)
		}

		if err := w.indirectReadWriter.WriteIndirect(
			doublyIndirectBlock,
			doublyIndirectIndex,
			singlyIndirectBlock,
		); err != nil {
			return fmt.Errorf(
				"writing singly indirect block address into doubly indirect "+
					"block: %w",
				err,
			)
		}

		if err := w.indirectReadWriter.WriteIndirect(
			inode.TriplyIndirectBlock,
			triplyIndirectIndex,
			doublyIndirectBlock,
		); err != nil {
			return fmt.Errorf(
				"writing doubly indirect block address into triply indirect "+
					"block: %w",
				err,
			)
		}

		return nil
	}

	// If the triply indirect block is valid, then we need to try to read the
	// doubly indirect block's address out of the triply indirect block.
	doublyIndirectBlock, err = w.indirectReadWriter.ReadIndirect(
		inode.TriplyIndirectBlock,
		triplyIndirectIndex,
	)
	if err != nil {
		return fmt.Errorf(
			"checking triply indirect block for doubly indirect block "+
				"address: %w",
			err,
		)
	}

	// If the doubly indirect block is invalid, then we need to:
	// * allocate a new doubly indirect block
	// * allocate a new singly indirect block
	// * write the physical block address into the singly indirect block
	// * write the newly allocated singly indirect block address into the
	//   doubly indirect block
	// * write the newly allocated doubly indirect block address into the
	//   triply indirect block
	if doublyIndirectBlock == BlockNil {
		doublyIndirectBlock, err = w.alloc()
		if err != nil {
			return fmt.Errorf("allocating doubly indirect block: %w", err)
		}
		singlyIndirectBlock, err = w.alloc()
		if err != nil {
			return fmt.Errorf("allocating singly indirect block: %w", err)
		}

		if err := w.indirectReadWriter.WriteIndirect(
			singlyIndirectBlock,
			singlyIndirectIndex,
			physicalBlock,
		); err != nil {
			return fmt.Errorf(
				"writing physical block address to singly indirect block: %w",
				err,
			)
		}

		if err := w.indirectReadWriter.WriteIndirect(
			doublyIndirectBlock,
			doublyIndirectIndex,
			singlyIndirectBlock,
		); err != nil {
			return fmt.Errorf(
				"writing singly indirect block address into the doubly "+
					"indirect block: %w",
				err,
			)
		}

		if err := w.indirectReadWriter.WriteIndirect(
			inode.TriplyIndirectBlock,
			triplyIndirectIndex,
			doublyIndirectBlock,
		); err != nil {
			return fmt.Errorf(
				"writing doubly indirect block address into the triply "+
					"indirect block: %w",
				err,
			)
		}

		return nil
	}

	// if the doubly indirect block is valid, then we need to try to read the
	// singly indirect block's address from the doubly indirect block.
	singlyIndirectBlock, err = w.indirectReadWriter.ReadIndirect(
		doublyIndirectBlock,
		doublyIndirectIndex,
	)
	if err != nil {
		return fmt.Errorf(
			"checking doubly indirect block for singly indirect "+
				"block address: %w",
			err,
		)
	}

	// if the singly indirect block's address is invalid, then we need to:
	// * allocate the singly indirect block
	// * write the physical block address into the singly indirect block
	// * write the address of the newly allocated singly indirect block into
	//   the doubly indirect block
	if singlyIndirectBlock == BlockNil {
		singlyIndirectBlock, err = w.alloc()
		if err != nil {
			return fmt.Errorf("allocating singly indirect block: %w", err)
		}

		if err := w.indirectReadWriter.WriteIndirect(
			singlyIndirectBlock,
			singlyIndirectIndex,
			physicalBlock,
		); err != nil {
			return fmt.Errorf(
				"writing physical block address to singly indirect block: %w",
				err,
			)
		}

		if err := w.indirectReadWriter.WriteIndirect(
			doublyIndirectBlock,
			doublyIndirectIndex,
			singlyIndirectBlock,
		); err != nil {
			return fmt.Errorf(
				"writing singly indirect block address into the doubly "+
					"indirect block: %w",
				err,
			)
		}

		return nil
	}

	// if the singly indirect block address is valid, then we only have to
	// write the physical block address into the singly indirect block
	if err := w.indirectReadWriter.WriteIndirect(
		singlyIndirectBlock,
		singlyIndirectIndex,
		physicalBlock,
	); err != nil {
		return fmt.Errorf(
			"writing physical block address to singly indirect block: %w",
			err,
		)
	}

	return nil
}

func (w *Writer) alloc() (Block, error) {
	if b, ok := w.allocator.Alloc(); ok {
		return Block(b), nil
	}
	return 0, OutOfBlocksErr
}
