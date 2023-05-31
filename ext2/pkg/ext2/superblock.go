package ext2

import (
	"encoding/binary"
	"fmt"
)

type SuperblockState uint16

type RevLevel uint32

const (
	SuperblockMagic uint16 = 0xef53

	// SuperblockSize is the size allocated for the superblock on disk.
	// The superblock doesn't actually use this much size; it seems to be more
	// of an upper-bound in case more fields were added to the superblock.
	SuperblockSize            uint16 = 1024
	SuperblockOffset          uint32 = 1024
	SupportedIncompatFeatures uint32 = 0x0002
	SupportedROCompatFeatures uint32 = 0

	StateClean SuperblockState = 1
	StateDirty SuperblockState = 2

	RevLevelStatic  RevLevel = 0
	RevLevelDynamic RevLevel = 1

	DefaultFirstIno  uint32 = 11
	DefaultInodeSize uint16 = 128
)

type Superblock struct {
	BlocksCount     uint32
	FreeBlocksCount uint32
	FreeInodesCount uint32
	FirstDataBlock  uint32
	LogBlockSize    uint32
	BlocksPerGroup  uint32
	InodesPerGroup  uint32
	State           SuperblockState
	RevLevel        RevLevel
	FirstIno        uint32
	InodeSize       uint16
	FeatureCompat   uint32
	FeatureIncompat uint32
	FeatureROCompat uint32
}

type ErrBadMagic struct {
	Found uint16
}

func (err ErrBadMagic) Error() string {
	return fmt.Sprintf(
		"bad magic: wanted `0x%2X`; found `%0#2x",
		SuperblockMagic,
		err.Found,
	)
}

type ErrBadState struct {
	Found SuperblockState
}

func (err ErrBadState) Error() string {
	return fmt.Sprintf(
		"bad state: wanted `0x%2X`; found `%0#2x`",
		StateClean,
		err.Found,
	)
}

type ErrIncompatibleFeatures struct {
	Found uint32
}

func (err ErrIncompatibleFeatures) Error() string {
	return fmt.Sprintf(
		"volume uses incompatible features: `%0#4x`",
		err.Found,
	)
}

type ErrIncompatibleFeaturesReadOnly struct {
	Found uint32
}

func (err ErrIncompatibleFeaturesReadOnly) Error() string {
	return fmt.Sprintf(
		"volume uses incompatible features; %s: `%0#4x`",
		"only reading is supported",
		err.Found,
	)
}

func DecodeSuperblock(
	b *[SuperblockSize]byte,
	readOnly bool,
) (Superblock, error) {
	var sb Superblock
	err := sb.Decode(b, readOnly)
	return sb, err
}

func (sb *Superblock) Decode(b *[SuperblockSize]byte, readOnly bool) error {
	magic := DecodeUint16(b[56], b[57])
	if magic != SuperblockMagic {
		return fmt.Errorf("decoding superblock: %w", ErrBadMagic{magic})
	}

	state := SuperblockState(DecodeUint16(b[58], b[59]))
	if state != StateClean {
		return fmt.Errorf("decoding superblock: %w", ErrBadState{state})
	}

	rev := RevLevel(DecodeUint32(b[76], b[77], b[78], b[79]))

	var featureCompat, featureIncompat, featureROCompat uint32
	if rev >= 1 {
		featureCompat = DecodeUint32(b[92], b[93], b[94], b[95])
		featureIncompat = DecodeUint32(b[96], b[97], b[98], b[99])
		featureROCompat = DecodeUint32(b[100], b[101], b[102], b[103])
	}

	if (featureIncompat & ^SupportedIncompatFeatures) != 0 {
		return fmt.Errorf(
			"decoding superblock: %w",
			ErrIncompatibleFeatures{featureIncompat},
		)
	}

	if !readOnly && (featureROCompat & ^SupportedROCompatFeatures) != 0 {
		return fmt.Errorf(
			"decoding superblock: %w",
			ErrIncompatibleFeaturesReadOnly{featureROCompat},
		)
	}

	sb.BlocksCount = DecodeUint32(b[4], b[5], b[6], b[7])
	sb.FreeBlocksCount = DecodeUint32(b[12], b[13], b[14], b[15])
	sb.FreeInodesCount = DecodeUint32(b[16], b[17], b[18], b[19])
	sb.FirstDataBlock = DecodeUint32(b[20], b[21], b[22], b[23])
	sb.LogBlockSize = DecodeUint32(b[24], b[25], b[26], b[27])
	sb.BlocksPerGroup = DecodeUint32(b[32], b[33], b[34], b[35])
	sb.InodesPerGroup = DecodeUint32(b[40], b[41], b[42], b[43])
	sb.State = state
	sb.RevLevel = rev
	if rev != RevLevelStatic {
		sb.FirstIno = DecodeUint32(b[84], b[85], b[86], b[87])
		sb.InodeSize = DecodeUint16(b[88], b[89])
	} else {
		sb.FirstIno = DefaultFirstIno
		sb.InodeSize = DefaultInodeSize
	}
	sb.FeatureCompat = featureCompat
	sb.FeatureIncompat = featureIncompat
	sb.FeatureROCompat = featureROCompat

	return nil
}

func DecodeUint16(b0, b1 byte) uint16 {
	// Little endian: first byte is least significant
	// https://en.wikipedia.org/wiki/Endianness
	return uint16(b0) + (uint16(b1) << 8)
}

func DecodeUint32(b0, b1, b2, b3 byte) uint32 {
	// Little endian: first byte is least significant
	// https://en.wikipedia.org/wiki/Endianness
	return uint32(b0) +
		(uint32(b1) << 8) +
		(uint32(b2) << 16) +
		(uint32(b3) << 24)
}

func (superblock *Superblock) Encode(b *[SuperblockSize]byte) {
	EncodeUint32(superblock.BlocksCount, b[4:])
	EncodeUint32(superblock.FreeBlocksCount, b[12:])
	EncodeUint32(superblock.FreeInodesCount, b[16:])
	EncodeUint32(superblock.FirstDataBlock, b[20:])
	EncodeUint32(superblock.LogBlockSize, b[24:])
	EncodeUint32(superblock.BlocksPerGroup, b[32:])
	EncodeUint32(superblock.InodesPerGroup, b[40:])
	EncodeUint16(SuperblockMagic, b[56:])
	EncodeUint16(uint16(superblock.State), b[58:])
	EncodeUint32(uint32(superblock.RevLevel), b[76:])

	if superblock.RevLevel != RevLevelStatic {
		EncodeUint32(superblock.FirstIno, b[84:])
		EncodeUint16(superblock.InodeSize, b[88:])
		EncodeUint32(superblock.FeatureCompat, b[92:])
		EncodeUint32(superblock.FeatureIncompat, b[96:])
		EncodeUint32(superblock.FeatureROCompat, b[100:])
	}
}

func EncodeUint16(x uint16, b []byte) {
	binary.LittleEndian.PutUint16(b, x)
}

func EncodeUint32(x uint32, b []byte) {
	binary.LittleEndian.PutUint32(b, x)
}
