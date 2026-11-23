package losetup

import (
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
)

// nolint:tagliatelle
type commandOutput struct {
	Devices []struct {
		AutoClear         *bool   `json:"autoclear"`
		BackingFile       *string `json:"back-file"`
		DirectIO          *bool   `json:"dio"`
		LogicalSectorSize *int16  `json:"log-sec"`
		Name              *string `json:"name"`
		Offset            *int64  `json:"offset"`
		ReadOnly          *bool   `json:"ro"`
		SizeLimit         *int64  `json:"sizelimit"`
	} `json:"loopdevices"`
}

func (v *commandOutput) toLoopDeviceList() (res model.LoopDeviceInfoList, err error) {
	if v == nil || v.Devices == nil {
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to map command output to model of loop devives info list; err: %w", err)
		}
	}()

	res = make(model.LoopDeviceInfoList, len(v.Devices))

	for i := range len(v.Devices) {
		dev := v.Devices[i]
		resItem := model.LoopDeviceInfo{}

		if dev.Name == nil || *dev.Name == "" {
			return res, fmt.Errorf("no device name specified for item at '%d' position", i)
		}

		resItem.Name = *dev.Name

		if dev.AutoClear == nil {
			return res, fmt.Errorf("no auto clear flag specified for item at '%d' position", i)
		}

		resItem.AutoClear = *dev.AutoClear

		if dev.DirectIO == nil {
			return res, fmt.Errorf("no direct IO flag specified for item at '%d' position", i)
		}

		resItem.DirectIO = *dev.DirectIO

		if dev.BackingFile == nil || *dev.BackingFile == "" {
			return res, fmt.Errorf("no device backing file specified for item at '%d' position", i)
		}

		resItem.BackingFile = *dev.BackingFile

		if dev.LogicalSectorSize == nil {
			return res, fmt.Errorf("no logical sector size specified for item at '%d' position", i)
		}

		resItem.LogicalSectorSize = *dev.LogicalSectorSize

		if dev.Offset == nil {
			return res, fmt.Errorf("no offset specified for item at '%d' position", i)
		}

		resItem.Offset = *dev.Offset

		if dev.ReadOnly == nil {
			return res, fmt.Errorf("no read-only flag specified for item at '%d' position", i)
		}

		resItem.ReadOnly = *dev.ReadOnly

		if dev.SizeLimit == nil {
			return res, fmt.Errorf("no size limit specified for item at '%d' position", i)
		}

		resItem.SizeLimit = *dev.SizeLimit

		res[i] = resItem
	}

	return
}
