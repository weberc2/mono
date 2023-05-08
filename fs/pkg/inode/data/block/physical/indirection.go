package physical

import (
	"fmt"

	"github.com/weberc2/mono/fs/pkg/inode/data/block/indirect"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type level int

const (
	levelDirect level = iota - 1
	levelSingly
	levelDoubly
	levelTriply
	levelOutOfRange
)

func (level level) String() string {
	switch level {
	case levelDirect:
		return "direct"
	case levelSingly:
		return "singly indirect"
	case levelDoubly:
		return "doubly indirect"
	case levelTriply:
		return "triply indirect"
	case levelOutOfRange:
		return "out of range"
	default:
		panic(fmt.Sprintf("invalid level: %d", level))
	}
}

type indirection struct {
	level    level
	indirect [levelOutOfRange]indirect.Index
	ptr      *Block
}

func (ind *indirection) fromInodeBlock(inode *Inode, block Block) error {
	if block <= directMax {
		ind.setDirect(inode, indirect.Index(block))
		return nil
	}

	if block <= singlyIndirectMax {
		ind.setSingly(inode, indirect.Index(block-directMax))
		return nil
	}

	if block <= doublyIndirectMax {
		base := block - singlyIndirectMax
		ind.setDoubly(
			inode,
			indirect.Index(base%singlyIndirectCount),
			indirect.Index(base/singlyIndirectCount),
		)
		return nil
	}

	// (block - doublyIndirectMax) / doublyIndirectCount = triplyIndirectIndex
	// block - doublyIndirectMax = triplyIndirectIndex * doublyIndirectCount
	// block = triplyIndirectIndex * doublyIndirectCount + doublyIndirectMax
	if block <= triplyIndirectMax {
		base := block - doublyIndirectMax
		ind.setTriply(
			inode,
			indirect.Index((base%doublyIndirectCount)%singlyIndirectCount),
			indirect.Index((base%doublyIndirectCount)/singlyIndirectCount),
			indirect.Index(base/doublyIndirectCount),
		)
		return nil
	}

	return OutOfRangeErr
}

func (ind *indirection) indices() []indirect.Index {
	return ind.indirect[:ind.level+1]
}

func (ind *indirection) singly() indirect.Index {
	return ind.indirect[levelSingly]
}

func (ind *indirection) doubly() indirect.Index {
	return ind.indirect[levelDoubly]
}

func (ind *indirection) triply() indirect.Index {
	return ind.indirect[levelTriply]
}

func (ind *indirection) setDirect(inode *Inode, index indirect.Index) {
	*ind = indirection{
		level:    levelDirect,
		indirect: [levelOutOfRange]indirect.Index{},
		ptr:      &inode.DirectBlocks[index],
	}
}

func (ind *indirection) setSingly(inode *Inode, singly indirect.Index) {
	*ind = indirection{
		level:    levelSingly,
		indirect: [levelOutOfRange]indirect.Index{levelSingly: singly},
		ptr:      &inode.SinglyIndirectBlock,
	}
}

func (ind *indirection) setDoubly(
	inode *Inode,
	singly indirect.Index,
	doubly indirect.Index,
) {
	*ind = indirection{
		level: levelDoubly,
		indirect: [levelOutOfRange]indirect.Index{
			levelSingly: singly,
			levelDoubly: doubly,
		},
		ptr: &inode.DoublyIndirectBlock,
	}
}

func (ind *indirection) setTriply(
	inode *Inode,
	singly indirect.Index,
	doubly indirect.Index,
	triply indirect.Index,
) {
	*ind = indirection{
		level: levelTriply,
		indirect: [levelOutOfRange]indirect.Index{
			levelSingly: singly,
			levelDoubly: doubly,
			levelTriply: triply,
		},
		ptr: &inode.TriplyIndirectBlock,
	}
}

func (ind *indirection) setOutOfRange() {
	*ind = indirection{level: levelOutOfRange}
}

// singly
// |____
// | | |

// doubly
// |______________
// |____  |____  |____
// | | |  | | |  | | |

// triply
// |____________________________________________
// |______________       |______________       |______________
// |____  |____  |____   |____  |____  |____   |____  |____  |____
// | | |  | | |  | | |   | | |  | | |  | | |   | | |  | | |  | | |)
const (
	pointersPerBlock    = Block(BlockSize / BlockPointerSize)
	directMax           = DirectBlocksCount - 1
	singlyIndirectCount = pointersPerBlock
	singlyIndirectMax   = singlyIndirectCount + directMax
	doublyIndirectCount = singlyIndirectCount * pointersPerBlock
	doublyIndirectMax   = doublyIndirectCount + singlyIndirectMax
	triplyIndirectCount = doublyIndirectCount * pointersPerBlock
	triplyIndirectMax   = triplyIndirectCount + doublyIndirectMax
)
