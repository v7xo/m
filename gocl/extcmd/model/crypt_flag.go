package model

import (
	"fmt"
	"math/bits"
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type CryptDeviceFlag uint64

const (
	CryptDeviceFlagUnspecified = CryptDeviceFlag(iota)
	CryptDeviceFlagOther       = CryptDeviceFlag(1 << (iota - 1))
	CryptDeviceFlagDiscards
	CryptDeviceFlagNoWriteWorkQueue
	CryptDeviceFlagNoReadWorkQueue
)

func (v CryptDeviceFlag) HasOther() bool { return v&CryptDeviceFlagOther == CryptDeviceFlagOther }

func (v CryptDeviceFlag) HasDiscards() bool {
	return v&CryptDeviceFlagDiscards == CryptDeviceFlagDiscards
}

func (v CryptDeviceFlag) HasNoWriteWorkQueue() bool {
	return v&CryptDeviceFlagNoWriteWorkQueue == CryptDeviceFlagNoWriteWorkQueue
}

func (v CryptDeviceFlag) HasNoReadWorkQueue() bool {
	return v&CryptDeviceFlagNoReadWorkQueue == CryptDeviceFlagNoReadWorkQueue
}

func (v CryptDeviceFlag) String() string {
	buf := make([]byte, 0, bits.OnesCount64(uint64(v))*32) //nolint

	if v.HasOther() {
		buf = append(buf, []byte("other,")...)
	}

	if v.HasDiscards() {
		buf = append(buf, []byte("discards,")...)
	}

	if v.HasNoReadWorkQueue() {
		buf = append(buf, []byte("no_read_workqueue,")...)
	}

	if v.HasNoWriteWorkQueue() {
		buf = append(buf, []byte("no_write_workqueue,")...)
	}

	if len(buf) == 0 {
		return placeholder.Unspecified
	}

	return string(buf[:len(buf)-1])
}

func (v *CryptDeviceFlag) FromCryptsetupOutput(val string) (err error) {
	if v == nil {
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to map crypt device flags from 'cryptsetup' command output '%s'; err: %w", val, err)
		}
	}()

	cp := CryptDeviceFlagUnspecified

	for flag := range strings.SplitSeq(val, " ") {
		switch flag {
		case "discards":
			cp |= CryptDeviceFlagDiscards
		case "no_read_workqueue":
			cp |= CryptDeviceFlagNoReadWorkQueue
		case "no_write_workqueue":
			cp |= CryptDeviceFlagNoWriteWorkQueue
		default:
			cp |= CryptDeviceFlagOther
		}
	}

	*v = cp

	return
}
