package ext2

type GroupID uint64

type Group struct {
	Idx         GroupID
	Desc        GroupDesc
	BlockBitmap DynamicBitmap
	InodeBitmap DynamicBitmap
	Dirty       bool
}
