package ext2

type BlockPosLevel uint8

const (
	PosLevel0     BlockPosLevel = 0
	PosLevel1     BlockPosLevel = 1
	PosLevel2     BlockPosLevel = 2
	PosLevel3     BlockPosLevel = 3
	PosOutOfRange BlockPosLevel = 4
)

type BlockPos struct {
	Level BlockPosLevel
	Data  [3]uint64
}

func BlockPosLevel0(data uint64) BlockPos {
	return BlockPos{
		Level: PosLevel0,
		Data:  [3]uint64{data, 0, 0},
	}
}

func BlockPosLevel1(data uint64) BlockPos {
	return BlockPos{
		Level: PosLevel1,
		Data:  [3]uint64{data, 0, 0},
	}
}

func BlockPosLevel2(data0, data1 uint64) BlockPos {
	return BlockPos{
		Level: PosLevel2,
		Data:  [3]uint64{data0, data1, 0},
	}
}

func BlockPosLevel3(data0, data1, data2 uint64) BlockPos {
	return BlockPos{
		Level: PosLevel3,
		Data:  [3]uint64{data0, data1, data2},
	}
}

func BlockPosOutOfRange() BlockPos {
	return BlockPos{Level: PosOutOfRange}
}
