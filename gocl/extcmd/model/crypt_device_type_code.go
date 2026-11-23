package model

import (
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/value/enum"
)

type CryptDeviceTypeCode int32

const (
	CryptDeviceTypeCodeUnspecified CryptDeviceTypeCode = iota
	CryptDeviceTypeCodeLUKS2
	CryptDeviceTypeCodeOther
)

var _ = enum.MustRegister(
	map[CryptDeviceTypeCode]string{
		CryptDeviceTypeCodeLUKS2:       "luks2",
		CryptDeviceTypeCodeOther:       "other",
		CryptDeviceTypeCodeUnspecified: "",
	})

func (v CryptDeviceTypeCode) String() string { return enum.String(v) }

func (v *CryptDeviceTypeCode) FromCryptsetupStatus(val string) (err error) {
	if v == nil {
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to map crypt device type code from '%s' string (cryptsetup status); err: %w", val, err)
		}
	}()

	switch val {
	case "":
		*v = CryptDeviceTypeCodeUnspecified
	case "luks2", "LUKS2":
		*v = CryptDeviceTypeCodeLUKS2
	default:
		*v = CryptDeviceTypeCodeOther
	}

	return
}
