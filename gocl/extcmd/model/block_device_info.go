package model

import (
	"strings"

	"github.com/google/uuid"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type BlockDeviceInfo struct {
	FileSystemType       *FileSystemTypeCode
	FileSystemUUID       *uuid.UUID
	FileSystemUUIDString *string
	FileSystemLabel      *string
	PartitionLabel       *string
	PartitionUUID        *uuid.UUID
	PartitionUUIDString  *string
	Path                 string
	Type                 BlockDeviceTypeCode
}

func (v BlockDeviceInfo) ToID() BlockDeviceID { return NewBlockDeviceID(v.Path, v.Type, nil) }

func (v BlockDeviceInfo) String() string {
	return "(path: " + placeholder.UnspecifiedIfEmpty(v.Path) + ")"
}

type BlockDeviceInfoList []BlockDeviceInfo

func (v BlockDeviceInfoList) String() string {
	if v == nil {
		return placeholder.Nil
	} else if len(v) == 0 {
		return placeholder.Empty
	}

	sb := &strings.Builder{}

	sb.WriteByte('[')

	sb.WriteString("count: ")
	sb.WriteString(stringsext.FormatInt(len(v)))

	for i := range len(v) {
		sb.WriteString(", ")

		if i == 32 {
			sb.WriteRune(stringsext.TruncateElipsisSym)

			break
		}

		sb.WriteString(v[i].String())
	}

	sb.WriteByte(']')

	return sb.String()
}

func (v BlockDeviceInfoList) ID() BlockDeviceIDList {
	if v == nil {
		return nil
	}

	res := make(BlockDeviceIDList, len(v))

	for i := range len(v) {
		res[i] = v[i].ToID()
	}

	return res
}
