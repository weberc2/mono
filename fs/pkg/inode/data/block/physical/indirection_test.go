package physical

import (
	"testing"

	"github.com/weberc2/mono/fs/pkg/types"
)

func TestIndirection(t *testing.T) {
	const (
		triplyIndirectBlock = 4
		triplyIndirectIndex = 6
		doublyIndirectBlock = 3
		doublyIndirectIndex = 5
		singlyIndirectBlock = 2
		singlyIndirectIndex = 10
		physicalBlock       = 1
	)
	var ind indirection
	var inode types.Inode
	ind.fromInodeBlock(
		&inode,
		doublyIndirectMax+singlyIndirectIndex+
			(doublyIndirectIndex*singlyIndirectCount)+
			(triplyIndirectIndex*doublyIndirectCount),
	)

	if ind.triply() != triplyIndirectIndex {
		t.Fatalf(
			"triply indirect index: wanted `%d`; found `%d`",
			triplyIndirectIndex,
			ind.triply(),
		)
	}

	if ind.doubly() != doublyIndirectIndex {
		t.Fatalf(
			"doubly indirect index: wanted `%d`; found `%d`",
			doublyIndirectIndex,
			ind.doubly(),
		)
	}

	if ind.singly() != singlyIndirectIndex {
		t.Fatalf(
			"singly indirect index: wanted `%d`; found `%d`",
			singlyIndirectIndex,
			ind.singly(),
		)
	}
}
