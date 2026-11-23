package model

import "rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"

type BtrfsCSum int32

const (
	BtrfsCSumUnspecified BtrfsCSum = iota
	BtrfsCSumCRC32C
	BtrfsCSumSHA256
	BtrfsCSumXXHash
	BtrfsCSumBlake2
)

func (v BtrfsCSum) String() string {
	switch v {
	case BtrfsCSumUnspecified:
		return placeholder.Unspecified
	case BtrfsCSumBlake2:
		return "blake2"
	case BtrfsCSumSHA256:
		return "sha256"
	case BtrfsCSumCRC32C:
		return "crc32c"
	case BtrfsCSumXXHash:
		return "xxhash"
	default:
		return placeholder.UnknownEnumInt(v)
	}
}
