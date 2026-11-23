package cryptsetup

import "rn.m.vo7.io/vo7io/m/gocl/extcmd/model"

const (
	maxCommandOutputBuffer     = model.DefaultCommandOutputLines
	luks2DeviceType            = "luks2"
	luksDeviceMapperPathPrefix = "/dev/mapper/"

	cryptsetupCmdName = "cryptsetup"

	LogScope = "gocl_extcmd_cryptsetup"
)
