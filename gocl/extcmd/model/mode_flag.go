package model

import (
	"fmt"
	"math/bits"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type ModeFlag uint64

const (
	ModeFlagRead = ModeFlag(1 << iota)
	ModeFlagWrite
)

func (v ModeFlag) HasRead() bool { return v&ModeFlagRead == ModeFlagRead }

func (v ModeFlag) HasWrite() bool { return v&ModeFlagWrite == ModeFlagWrite }

func (v ModeFlag) String() string {
	buf := make([]byte, 0, bits.OnesCount64(uint64(v))*16) //nolint

	if v.HasRead() {
		buf = append(buf, []byte("read,")...)
	}

	if v.HasWrite() {
		buf = append(buf, []byte("write,")...)
	}

	if len(buf) == 0 {
		return placeholder.Unspecified
	}

	return string(buf[:len(buf)-1])
}

func (v *ModeFlag) FromCryptsetupOutput(val string) (err error) {
	if v == nil {
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to map mode flags from 'cryptsetup' command output; err: %w", err)
		}
	}()

	switch val {
	case "read/write":
		*v = ModeFlagRead | ModeFlagWrite
	case "read":
		*v = ModeFlagRead
	case "write":
		*v = ModeFlagRead
	default:
		err = fmt.Errorf("it is unknown how to map '%s' output", val)
	}

	return
}
