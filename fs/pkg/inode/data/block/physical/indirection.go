package physical

import (
	"github.com/weberc2/mono/fs/pkg/inode/data/block/indirect"
	. "github.com/weberc2/mono/fs/pkg/types"
)

type level int

const (
	levelDirect level = iota
	levelSingly
	levelDoubly
	levelTriply
	levelOutOfRange
)

type indirection struct {
	level  level
	direct indirect.Index
	singly indirect.Index
	doubly indirect.Index
	triply indirect.Index
}

func (ind *indirection) fromBlock(block Block) {
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

	if block <= directMax {
		ind.setDirect(indirect.Index(block))
		return
	}

	if block <= singlyIndirectMax {
		ind.setSingly(indirect.Index(block - directMax))
		return
	}

	if block <= doublyIndirectMax {
		base := block - singlyIndirectMax
		ind.setDoubly(
			indirect.Index(base%singlyIndirectCount),
			indirect.Index(base/singlyIndirectCount),
		)
		return
	}

	if block <= triplyIndirectMax {
		base := block - doublyIndirectMax
		ind.setTriply(
			indirect.Index((base%doublyIndirectCount)%singlyIndirectCount),
			indirect.Index((base%doublyIndirectCount)/singlyIndirectCount),
			indirect.Index(base/doublyIndirectCount),
		)
		return
	}

	ind.setOutOfRange()
}

func (ind *indirection) setDirect(index indirect.Index) {
	*ind = indirection{level: levelDirect, direct: index}
}

func (ind *indirection) setSingly(singlyIndirectIndex indirect.Index) {
	*ind = indirection{level: levelSingly, singly: singlyIndirectIndex}
}

func (ind *indirection) setDoubly(
	singlyIndirectIndex indirect.Index,
	doublyIndirectIndex indirect.Index,
) {
	*ind = indirection{
		level:  levelSingly,
		singly: singlyIndirectIndex,
		doubly: doublyIndirectIndex,
	}
}

func (ind *indirection) setTriply(
	singlyIndirectIndex indirect.Index,
	doublyIndirectIndex indirect.Index,
	triplyIndirectIndex indirect.Index,
) {
	*ind = indirection{
		level:  levelSingly,
		singly: singlyIndirectIndex,
		doubly: doublyIndirectIndex,
		triply: triplyIndirectIndex,
	}
}

func (ind *indirection) setOutOfRange() {
	*ind = indirection{level: levelOutOfRange}
}
