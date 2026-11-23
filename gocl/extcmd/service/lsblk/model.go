package lsblk

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/misc/uuidext"
)

//nolint:tagliatelle
type commandOutput struct {
	Devices []struct {
		FileSystemUUID  *string `json:"uuid"`
		FileSystemType  *string `json:"fstype"`
		FileSystemLabel *string `json:"label"`
		PartLabel       *string `json:"partlabel"`
		PartitionUUID   *string `json:"partuuid"`
		Path            *string `json:"path"`
		Type            *string `json:"type"`
	} `json:"blockdevices"`
}

func (v *commandOutput) toBlockDeviceList() (res model.BlockDeviceInfoList, err error) {
	if v == nil || v.Devices == nil {
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to map command output to model of block devives info list; err: %w", err)
		}
	}()

	res = make(model.BlockDeviceInfoList, len(v.Devices))

	for i := range len(v.Devices) {
		item := v.Devices[i]
		resItem := model.BlockDeviceInfo{}

		if item.Path == nil || *item.Path == "" {
			err = fmt.Errorf("no device path specified for item at '%d' position", i)

			return
		}

		resItem.Path = *item.Path

		if item.Type != nil && *item.Type != "" {
			if err = resItem.Type.FromLsblkOutputString(*item.Type); err != nil {
				return
			}
		}

		if item.FileSystemLabel != nil && *item.FileSystemLabel != "" {
			resItem.FileSystemLabel = item.FileSystemLabel
		}

		if item.PartLabel != nil && *item.PartLabel != "" {
			resItem.PartitionLabel = item.FileSystemLabel
		}

		if item.FileSystemType != nil && *item.FileSystemType != "" {
			resItem.FileSystemType = new(model.FileSystemTypeCodeUnspecified)

			if err = resItem.FileSystemType.FromString(*item.FileSystemType); err != nil {
				return
			}
		}

		if item.PartitionUUID != nil && *item.PartitionUUID != "" {
			if uuidParsed, uuidParseErr := uuid.Parse(*item.PartitionUUID); uuidParseErr == nil {
				resItem.PartitionUUID = &uuidParsed
			}

			resItem.PartitionUUIDString = item.PartitionUUID
		}

		if item.FileSystemUUID != nil && *item.FileSystemUUID != "" {
			if uuidParsed, uuidParseErr := uuid.Parse(*item.FileSystemUUID); uuidParseErr == nil {
				resItem.FileSystemUUID = &uuidParsed
			}

			resItem.FileSystemUUIDString = item.FileSystemUUID
		}

		res[i] = resItem
	}

	return
}

type Filter struct {
	PartitionLabel *[]string
	PartitionUUID  *uuidext.List
	Type           *model.BlockDeviceTypeCodeList
}

func (filt Filter) Match(val model.BlockDeviceInfo) bool {
	res := true

	if filt.PartitionLabel != nil && *filt.PartitionLabel != nil {
		m := false

		if val.Type.IsPartition() && val.PartitionLabel != nil {
			for i := 0; !m && i < len(*filt.PartitionLabel); i++ {
				m = strings.EqualFold(*val.PartitionLabel, (*filt.PartitionLabel)[i])
			}
		}

		res = res && m
	}

	if !res {
		return false
	}

	if filt.PartitionUUID != nil && *filt.PartitionUUID != nil {
		m := false

		if val.Type.IsPartition() && val.PartitionUUID != nil {
			for i := 0; !m && i < len(*filt.PartitionUUID); i++ {
				m = *val.PartitionUUID == (*filt.PartitionUUID)[i]
			}
		}

		res = res && m
	}

	if !res {
		return false
	}

	if filt.Type != nil && *filt.Type != nil {
		m := false

		for i := 0; !m && i < len(*filt.Type); i++ {
			m = val.Type == (*filt.Type)[i]
		}

		res = res && m
	}

	return res
}

func (filt Filter) Matched(val model.BlockDeviceInfoList) model.BlockDeviceInfoList {
	if val == nil || filt.Empty() {
		return val
	}

	res := make(model.BlockDeviceInfoList, 0, len(val))

	for i := range len(val) {
		if filt.Match(val[i]) {
			res = append(res, val[i])
		}
	}

	return res
}

func (filt Filter) Empty() bool {
	return (filt.PartitionLabel == nil || *filt.PartitionLabel == nil) &&
		(filt.PartitionUUID == nil || *filt.PartitionUUID == nil) &&
		(filt.Type == nil || *filt.Type == nil)
}

func (filt Filter) String() string {
	sb := strings.Builder{}

	sb.WriteByte('(')

	if filt.PartitionLabel != nil && *filt.PartitionLabel != nil {
		sb.WriteString("partition label: [")

		sb.WriteString("count: ")
		sb.WriteString(strconv.Itoa(len(*filt.PartitionLabel)))

		for _, v := range *filt.PartitionLabel {
			sb.WriteString(", ")

			sb.WriteByte('\'')
			sb.WriteString(v)
			sb.WriteByte('\'')
		}

		sb.WriteByte(']')
	}

	if filt.PartitionUUID != nil && *filt.PartitionUUID != nil {
		if sb.Len() > 1 {
			sb.WriteString(", ")
		}

		sb.WriteString("partition UUID: ")
		sb.WriteString(filt.PartitionUUID.String())
	}

	if filt.Type != nil && *filt.Type != nil {
		if sb.Len() > 1 {
			sb.WriteString(", ")
		}

		sb.WriteString("type: ")
		sb.WriteString(filt.Type.String())
	}

	sb.WriteByte(')')

	return sb.String()
}
