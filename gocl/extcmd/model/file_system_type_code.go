package model

import (
	"rn.m.vo7.io/vo7io/m/gocl/value/enum"
)

type FileSystemTypeCode int32

const (
	FileSystemTypeCodeUnspecified FileSystemTypeCode = iota
	FileSystemTypeCodeBtrfs
	FileSystemTypeCodeVfat
	FileSystemTypeCodeExt4
	FileSystemTypeCodeXfs
	FileSystemTypeCodeCryptoLUKS
	FileSystemTypeCodeOther
)

var _ = enum.MustRegister(
	map[FileSystemTypeCode]string{
		FileSystemTypeCodeBtrfs:       "btrfs",
		FileSystemTypeCodeCryptoLUKS:  "crypto_luks",
		FileSystemTypeCodeExt4:        "ext4",
		FileSystemTypeCodeOther:       "other",
		FileSystemTypeCodeUnspecified: "",
		FileSystemTypeCodeVfat:        "vfat",
		FileSystemTypeCodeXfs:         "xfs",
	})

func (v FileSystemTypeCode) ToLinuxKernelKnownName() (string, bool) {
	switch v {
	case FileSystemTypeCodeBtrfs:
		return "btrfs", true
	case FileSystemTypeCodeVfat:
		return "vfat", true
	case FileSystemTypeCodeExt4:
		return "ext4", true
	case FileSystemTypeCodeXfs:
		return "xfs", true
	case
		FileSystemTypeCodeUnspecified,
		FileSystemTypeCodeCryptoLUKS,
		FileSystemTypeCodeOther:
		fallthrough
	default:
		return "", false
	}
}

func (v FileSystemTypeCode) IsBtrfs() bool       { return v == FileSystemTypeCodeBtrfs }
func (v FileSystemTypeCode) IsExt4() bool        { return v == FileSystemTypeCodeExt4 }
func (v FileSystemTypeCode) IsOther() bool       { return v == FileSystemTypeCodeOther }
func (v FileSystemTypeCode) IsUnspecified() bool { return v == FileSystemTypeCodeUnspecified }
func (v FileSystemTypeCode) IsVfat() bool        { return v == FileSystemTypeCodeVfat }
func (v FileSystemTypeCode) IsXfs() bool         { return v == FileSystemTypeCodeXfs }

func (v FileSystemTypeCode) String() string { return enum.String(v) }

func (v *FileSystemTypeCode) FromString(val string) error {
	if v == nil {
		return nil
	}

	parsed, err := enum.ParseFallback[FileSystemTypeCode](val, FileSystemTypeCodeOther)
	if err != nil {
		return err
	}

	*v = parsed

	return nil
}
