package model

import (
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
	"rn.m.vo7.io/vo7io/m/gocl/value/enum"
)

type BlockDeviceTypeCode int32

const (
	BlockDeviceTypeCodeUnspecified BlockDeviceTypeCode = iota
	BlockDeviceTypeCodeCrypt
	BlockDeviceTypeCodeDisk
	BlockDeviceTypeCodeOther
	BlockDeviceTypeCodePartition
	BlockDeviceTypeCodeLoop
	BlockDeviceTypeCodeFile
	BlockDeviceTypeCodeNBD
	BlockDeviceTypeCodeLVM
)

var _ = enum.MustRegister(
	map[BlockDeviceTypeCode]string{
		BlockDeviceTypeCodeUnspecified: "",

		BlockDeviceTypeCodeCrypt:     "crypt",
		BlockDeviceTypeCodeDisk:      "disk",
		BlockDeviceTypeCodeFile:      "file",
		BlockDeviceTypeCodeLoop:      "loop",
		BlockDeviceTypeCodeLVM:       "lvm",
		BlockDeviceTypeCodeNBD:       "nbd",
		BlockDeviceTypeCodeOther:     "other",
		BlockDeviceTypeCodePartition: "part",
	})

func (v BlockDeviceTypeCode) IsCrypt() bool     { return v == BlockDeviceTypeCodeCrypt }
func (v BlockDeviceTypeCode) IsDisk() bool      { return v == BlockDeviceTypeCodeDisk }
func (v BlockDeviceTypeCode) IsFile() bool      { return v == BlockDeviceTypeCodeFile }
func (v BlockDeviceTypeCode) IsLoop() bool      { return v == BlockDeviceTypeCodeLoop }
func (v BlockDeviceTypeCode) IsLVM() bool       { return v == BlockDeviceTypeCodeLVM }
func (v BlockDeviceTypeCode) IsNBD() bool       { return v == BlockDeviceTypeCodeNBD }
func (v BlockDeviceTypeCode) IsPartition() bool { return v == BlockDeviceTypeCodePartition }

func (v BlockDeviceTypeCode) String() string { return enum.String(v) }

func (v *BlockDeviceTypeCode) FromLsblkOutputString(val string) (err error) {
	if v == nil {
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to map '%s' from '%s' string (lsblk); %w", typestring.TypeStringForExpr(v), val, err)
		}
	}()

	switch val {
	case "":
		*v = BlockDeviceTypeCodeUnspecified
	case "crypt":
		*v = BlockDeviceTypeCodeCrypt
	case "disk":
		*v = BlockDeviceTypeCodeDisk
	case "part":
		*v = BlockDeviceTypeCodePartition
	case "loop":
		*v = BlockDeviceTypeCodeLoop
	case "nbd":
		*v = BlockDeviceTypeCodeNBD
	case "lvm":
		*v = BlockDeviceTypeCodeLVM
	default:
		*v = BlockDeviceTypeCodeOther
	}

	return nil
}

type BlockDeviceTypeCodeList []BlockDeviceTypeCode

func (v BlockDeviceTypeCodeList) String() string { return enum.StringSlice(v) }
