package physical

import (
	"fmt"
	"log"
	"testing"

	"github.com/weberc2/mono/fs/pkg/alloc"
	"github.com/weberc2/mono/fs/pkg/inode/data/block/indirect"
	"github.com/weberc2/mono/fs/pkg/inode/store"
	"github.com/weberc2/mono/fs/pkg/io"
	. "github.com/weberc2/mono/fs/pkg/types"
)

var testCases = []testCase{{
	name:  "direct-exists",
	state: defaultState(),
	inputInode: Inode{
		DirectBlocks: [...]Block{0, 100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	},
	wantedInode: Inode{
		DirectBlocks: [...]Block{0, 100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	},
	inputBlock:  1,
	wantedBlock: 100,
}, {
	name:  "direct-nil",
	state: defaultState(),
	wantedInode: Inode{
		DirectBlocks: [...]Block{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	},
	inputBlock:  0,
	wantedBlock: 1,
	hook: func() {
		log.Println()
	},
}, func() testCase {
	indirect := ind{outer: 10, index: 5, inner: 6}
	return testCase{
		name:        "singly-indirect-exists-physical-exists",
		state:       defaultState(),
		inputInode:  Inode{SinglyIndirectBlock: indirect.outer},
		wantedInode: Inode{SinglyIndirectBlock: indirect.outer},
		inputBlock:  DirectBlocksCount + Block(indirect.index) - 1,
		wantedBlock: indirect.inner,
	}.givenIndirects(indirect)
}(), func() testCase {
	const (
		outer = 10
		index = 5
		inner = 1 // first free block when allocator is invoked
	)
	return testCase{
		name:        "singly-indirect-exists-physical-invalid",
		state:       defaultState(),
		inputInode:  Inode{SinglyIndirectBlock: outer},
		wantedInode: Inode{SinglyIndirectBlock: outer},
		inputBlock:  DirectBlocksCount + index - 1,
		wantedBlock: inner,
	}
}(), func() testCase {
	const (
		outer = 1
		index = 5
		inner = 2
	)
	state := defaultState()
	return testCase{
		name:        "singly-indirect-invalid",
		state:       state,
		wantedInode: Inode{SinglyIndirectBlock: outer},
		inputBlock:  directMax + index,
		wantedBlock: inner,
	}.wantedIndirects(ind{outer: outer, index: index, inner: inner})
}(), func() testCase {
	const (
		doublyIndirectBlock = 3
		doublyIndirectIndex = 5
		singlyIndirectBlock = 2
		singlyIndirectIndex = 10
		physicalBlock       = 1
	)
	return testCase{
		name:        "doubly-indirect-all-valid",
		state:       defaultState(),
		inputInode:  Inode{DoublyIndirectBlock: doublyIndirectBlock},
		wantedInode: Inode{DoublyIndirectBlock: doublyIndirectBlock},
		inputBlock:  singlyIndirectMax + (doublyIndirectIndex * singlyIndirectCount) + singlyIndirectIndex,
		wantedBlock: physicalBlock,
	}.givenIndirects(ind{
		outer: doublyIndirectBlock,
		index: doublyIndirectIndex,
		inner: singlyIndirectBlock,
	}, ind{
		outer: singlyIndirectBlock,
		index: singlyIndirectIndex,
		inner: physicalBlock,
	})
}(), func() testCase {
	const (
		doublyIndirectBlock = 3
		doublyIndirectIndex = 5
		singlyIndirectBlock = 2
		singlyIndirectIndex = 10
		physicalBlock       = 1
	)
	state := defaultState()
	return testCase{
		name:        "doubly-indirect-and-singly-indirect-valid-physical-invalid",
		state:       state,
		inputInode:  Inode{DoublyIndirectBlock: doublyIndirectBlock},
		wantedInode: Inode{DoublyIndirectBlock: doublyIndirectBlock},
		inputBlock:  singlyIndirectMax + (doublyIndirectIndex * singlyIndirectCount) + singlyIndirectIndex,
		wantedBlock: physicalBlock,
	}.givenIndirects(ind{
		outer: doublyIndirectBlock,
		index: doublyIndirectIndex,
		inner: singlyIndirectBlock,
	}).wantedIndirects(ind{
		outer: singlyIndirectBlock,
		index: singlyIndirectIndex,
		inner: physicalBlock,
	})
}(), func() testCase {
	const (
		doublyIndirectBlock = 1
		doublyIndirectIndex = 5
		singlyIndirectBlock = 2
		singlyIndirectIndex = 10
		physicalBlock       = 3
	)

	state := defaultState()

	return testCase{
		name:        "doubly-indirect-valid-singly-indirect-and-physical-invalid",
		state:       state,
		inputInode:  Inode{DoublyIndirectBlock: doublyIndirectBlock},
		wantedInode: Inode{DoublyIndirectBlock: doublyIndirectBlock},
		inputBlock:  singlyIndirectMax + (doublyIndirectIndex * singlyIndirectCount) + singlyIndirectIndex,
		wantedBlock: physicalBlock,
	}.wantedIndirects(ind{
		outer: singlyIndirectBlock,
		index: singlyIndirectIndex,
		inner: physicalBlock,
	}, ind{
		outer: doublyIndirectBlock,
		index: doublyIndirectIndex,
		inner: singlyIndirectBlock,
	})
}(), func() testCase {
	const (
		doublyIndirectBlock = 1
		doublyIndirectIndex = 5
		singlyIndirectBlock = 2
		singlyIndirectIndex = 10
		physicalBlock       = 3
	)
	state := defaultState()
	return testCase{
		name:        "doubly-indirect-none-valid",
		state:       state,
		inputInode:  Inode{},
		wantedInode: Inode{DoublyIndirectBlock: doublyIndirectBlock},
		inputBlock:  singlyIndirectMax + (doublyIndirectIndex * singlyIndirectCount) + singlyIndirectIndex,
		wantedBlock: physicalBlock,
	}.wantedIndirects(ind{
		outer: singlyIndirectBlock,
		index: singlyIndirectIndex,
		inner: physicalBlock,
	}, ind{
		outer: doublyIndirectBlock,
		index: doublyIndirectIndex,
		inner: singlyIndirectBlock,
	})
}(), func() testCase {
	const (
		triplyIndirectBlock = 4
		triplyIndirectIndex = 0
		doublyIndirectBlock = 3
		doublyIndirectIndex = 5
		singlyIndirectBlock = 2
		singlyIndirectIndex = 10
		physicalBlock       = 1
	)
	state := defaultState()
	return testCase{
		name:        "triply-indirect-all-valid",
		state:       state,
		inputInode:  Inode{TriplyIndirectBlock: triplyIndirectBlock},
		wantedInode: Inode{TriplyIndirectBlock: triplyIndirectBlock},
		inputBlock:  doublyIndirectMax + singlyIndirectIndex + (doublyIndirectIndex * singlyIndirectCount) + (triplyIndirectIndex * doublyIndirectCount),
		wantedBlock: physicalBlock,
	}.givenIndirects(ind{
		outer: triplyIndirectBlock,
		index: triplyIndirectIndex,
		inner: doublyIndirectBlock,
	}, ind{
		outer: doublyIndirectBlock,
		index: doublyIndirectIndex,
		inner: singlyIndirectBlock,
	}, ind{
		outer: singlyIndirectBlock,
		index: singlyIndirectIndex,
		inner: physicalBlock,
	})
}(), func() testCase {
	const (
		triplyIndirectBlock = 4
		triplyIndirectIndex = 0
		doublyIndirectBlock = 3
		doublyIndirectIndex = 5
		singlyIndirectBlock = 2
		singlyIndirectIndex = 10
		physicalBlock       = 1
	)
	state := defaultState()
	return testCase{
		name:        "triply-doubly-and-singly-indirects-valid-physical-invalid",
		state:       state,
		inputInode:  Inode{TriplyIndirectBlock: triplyIndirectBlock},
		wantedInode: Inode{TriplyIndirectBlock: triplyIndirectBlock},
		inputBlock:  doublyIndirectMax + singlyIndirectIndex + (doublyIndirectIndex * singlyIndirectCount) + (triplyIndirectIndex * doublyIndirectCount),
		wantedBlock: physicalBlock,
	}.givenIndirects(ind{
		outer: triplyIndirectBlock,
		index: triplyIndirectIndex,
		inner: doublyIndirectBlock,
	}, ind{
		outer: doublyIndirectBlock,
		index: doublyIndirectIndex,
		inner: singlyIndirectBlock,
	}).wantedIndirects(ind{
		outer: singlyIndirectBlock,
		index: singlyIndirectIndex,
		inner: physicalBlock,
	})
}(), func() testCase {
	const (
		triplyIndirectBlock = 1
		triplyIndirectIndex = 6
		doublyIndirectBlock = 2
		doublyIndirectIndex = 5
		singlyIndirectBlock = 3
		singlyIndirectIndex = 10
		physicalBlock       = 4
	)
	state := defaultState()
	return testCase{
		name:        "triply-and-doubly-indirects-valid-singly-and-physical-invalid",
		state:       state,
		inputInode:  Inode{TriplyIndirectBlock: triplyIndirectBlock},
		wantedInode: Inode{TriplyIndirectBlock: triplyIndirectBlock},
		inputBlock:  doublyIndirectMax + singlyIndirectIndex + (doublyIndirectIndex * singlyIndirectCount) + (triplyIndirectIndex * doublyIndirectCount),
		wantedBlock: physicalBlock,
	}.givenIndirects(ind{
		outer: triplyIndirectBlock,
		index: triplyIndirectIndex,
		inner: doublyIndirectBlock,
	}).wantedIndirects(ind{
		outer: doublyIndirectBlock,
		index: doublyIndirectIndex,
		inner: singlyIndirectBlock,
	}, ind{
		outer: singlyIndirectBlock,
		index: singlyIndirectIndex,
		inner: physicalBlock,
	})
}(), func() testCase {
	const (
		triplyIndirectBlock = 1
		triplyIndirectIndex = 6
		doublyIndirectBlock = 2
		doublyIndirectIndex = 5
		singlyIndirectBlock = 3
		singlyIndirectIndex = 10
		physicalBlock       = 4
	)
	state := defaultState()
	return testCase{
		name:        "triply-indirect-valid-doubly-singly-and-physical-invalid",
		state:       state,
		inputInode:  Inode{TriplyIndirectBlock: triplyIndirectBlock},
		wantedInode: Inode{TriplyIndirectBlock: triplyIndirectBlock},
		inputBlock:  doublyIndirectMax + singlyIndirectIndex + (doublyIndirectIndex * singlyIndirectCount) + (triplyIndirectIndex * doublyIndirectCount),
		wantedBlock: physicalBlock,
	}.wantedIndirects(ind{
		outer: triplyIndirectBlock,
		index: triplyIndirectIndex,
		inner: doublyIndirectBlock,
	}, ind{
		outer: doublyIndirectBlock,
		index: doublyIndirectIndex,
		inner: singlyIndirectBlock,
	}, ind{
		outer: singlyIndirectBlock,
		index: singlyIndirectIndex,
		inner: physicalBlock,
	})
}(), func() testCase {
	const (
		triplyIndirectBlock = 1
		triplyIndirectIndex = 6
		doublyIndirectBlock = 2
		doublyIndirectIndex = 5
		singlyIndirectBlock = 3
		singlyIndirectIndex = 10
		physicalBlock       = 4
	)
	state := defaultState()
	return testCase{
		name:        "triply-indirect-none-valid",
		state:       state,
		inputInode:  Inode{},
		wantedInode: Inode{TriplyIndirectBlock: triplyIndirectBlock},
		inputBlock:  doublyIndirectMax + singlyIndirectIndex + (doublyIndirectIndex * singlyIndirectCount) + (triplyIndirectIndex * doublyIndirectCount),
		wantedBlock: physicalBlock,
	}.wantedIndirects(ind{
		outer: triplyIndirectBlock,
		index: triplyIndirectIndex,
		inner: doublyIndirectBlock,
	}, ind{
		outer: doublyIndirectBlock,
		index: doublyIndirectIndex,
		inner: singlyIndirectBlock,
	}, ind{
		outer: singlyIndirectBlock,
		index: singlyIndirectIndex,
		inner: physicalBlock,
	})
}()}

func TestReadOrAlloc(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantedError == nil {
				tc.wantedError = func(err error) error {
					if err != nil {
						return fmt.Errorf("unexpected err: %w", err)
					}
					return nil
				}
			}

			// reserve any blocks referenced directly by the inode
			reserveAll(tc.state.allocator, &tc.inputInode)

			if tc.wantedState == nil {
				tc.wantedState = func() error { return nil }
			}
			if tc.hook != nil {
				tc.hook()
			}
			actual, err := tc.state.readWriter.ReadAlloc(
				&tc.inputInode,
				tc.inputBlock,
			)
			if err := tc.wantedError(err); err != nil {
				t.Fatal(err)
			}
			if actual != tc.wantedBlock {
				t.Fatalf(
					"wanted block `%d`; found `%d`",
					tc.wantedBlock,
					actual,
				)
			}
			if err := tc.wantedState(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

type testCase struct {
	name        string
	state       state
	inputInode  Inode
	wantedInode Inode
	inputBlock  Block
	wantedBlock Block
	wantedError func(error) error
	wantedState func() error
	hook        func()
}

func (tc testCase) givenIndirects(indirects ...ind) testCase {
	for _, indirect := range indirects {
		if err := tc.state.indirects.WriteIndirect(
			indirect.outer,
			indirect.index,
			indirect.inner,
		); err != nil {
			log.Fatalf(
				"modifying test case `%s`: preparing indirects "+
					"`%d[%d] = %d`: %v",
				tc.name,
				indirect.outer,
				indirect.index,
				indirect.inner,
				err,
			)
		}
		tc.state.allocator.Reserve(indirect.inner)
	}
	return tc
}

type state struct {
	allocator  alloc.BlockAllocator
	inodeStore InodeStore
	indirects  indirect.ReadWriter
	readWriter ReadWriter
}

func defaultState() state {
	var state state
	buf := make([]byte, 1024*1024)
	volume := io.NewBuffer(buf[:1024*512])
	bitmap := alloc.New(1024)
	state.allocator = alloc.BlockAllocator{Allocator: &bitmap}
	state.indirects = indirect.NewReadWriter(volume)
	state.inodeStore = store.NewVolumeInodeStore(io.NewBuffer(buf[1024*512:]))
	state.readWriter = NewReadWriter(
		state.allocator,
		state.indirects,
		state.inodeStore,
	)
	return state
}

func (tc testCase) wantedIndirects(indirects ...ind) testCase {
	prevWantedState := tc.wantedState
	tc.wantedState = func() error {
		if prevWantedState != nil {
			if err := prevWantedState(); err != nil {
				return err
			}
		}
		for _, indirect := range indirects {
			actual, err := tc.state.indirects.ReadIndirect(
				indirect.outer,
				indirect.index,
			)
			if err != nil {
				return fmt.Errorf(
					"ReadIndirect(): unexpected err: %v",
					err,
				)
			}
			if actual != indirect.inner {
				return fmt.Errorf(
					"ReadIndirect(): wanted `%d`; found `%d`",
					indirect.inner,
					actual,
				)
			}
		}
		return nil
	}

	return tc
}

type ind struct {
	outer Block
	index indirect.Index
	inner Block
}

func reserveAll(a alloc.BlockAllocator, i *Inode) {
	for _, b := range i.DirectBlocks {
		if b != BlockNil {
			a.Reserve(b)
		}
	}
	if i.SinglyIndirectBlock != BlockNil {
		a.Reserve(i.SinglyIndirectBlock)
	}
	if i.DoublyIndirectBlock != BlockNil {
		a.Reserve(i.DoublyIndirectBlock)
	}
	if i.TriplyIndirectBlock != BlockNil {
		a.Reserve(i.TriplyIndirectBlock)
	}
}
