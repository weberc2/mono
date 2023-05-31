package ext2

type GroupDesc struct {
	BlockBitmap     uint32
	InodeBitmap     uint32
	InodeTable      uint32
	FreeBlocksCount uint16
	FreeInodesCount uint16
	UsedDirsCount   uint16
}

// NB: The Rust implementation takes a `Superblock` argument but doesn't use
// it. Similarly, it returned a `Result<GroupDesc>` even though it only ever
// returned `Ok(...)`. These seemed vestigial, so I'm omitting here.
func DecodeGroupDesc(b *[GroupDescSize]byte) GroupDesc {
	return GroupDesc{
		BlockBitmap:     DecodeUint32(b[0], b[1], b[2], b[3]),
		InodeBitmap:     DecodeUint32(b[4], b[5], b[6], b[7]),
		InodeTable:      DecodeUint32(b[8], b[9], b[10], b[11]),
		FreeBlocksCount: DecodeUint16(b[12], b[13]),
		FreeInodesCount: DecodeUint16(b[14], b[15]),
		UsedDirsCount:   DecodeUint16(b[16], b[17]),
	}
}

// pub fn encode_group_desc(_superblock: &Superblock,
//
//	  group_desc: &GroupDesc, bytes: &mut [u8]) -> Result<()>
//	  {
//	  encode_u32(group_desc.block_bitmap, &mut bytes[0..]);
//	  encode_u32(group_desc.inode_bitmap, &mut bytes[4..]);
//	  encode_u32(group_desc.inode_table, &mut bytes[8..]);
//	  encode_u16(group_desc.free_blocks_count, &mut bytes[12..]);
//	  encode_u16(group_desc.free_inodes_count, &mut bytes[14..]);
//	  encode_u16(group_desc.used_dirs_count, &mut bytes[16..]);
//	  Ok(())
//	}
func (desc *GroupDesc) Encode(b *[GroupDescSize]byte) {
	EncodeUint32(desc.BlockBitmap, b[0:])
	EncodeUint32(desc.InodeBitmap, b[4:])
	EncodeUint32(desc.InodeTable, b[8:])
	EncodeUint16(desc.FreeBlocksCount, b[12:])
	EncodeUint16(desc.FreeInodesCount, b[14:])
	EncodeUint16(desc.UsedDirsCount, b[16:])
}

const (
	// GroupDescSize is the size of a GroupDesc on disk in bytes. This value is
	// pulled directly from the Rust implementation:
	// * https://github.com/honzasp/libext2/blob/4d54d8d/src/decode.rs#L52
	// * https://github.com/honzasp/libext2/blob/4d54d8d/src/group.rs#L28
	GroupDescSize = 32
)
