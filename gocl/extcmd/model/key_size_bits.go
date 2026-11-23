package model

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type KeySizeBits uint16

func (v KeySizeBits) String() string {
	if v == 0 {
		return placeholder.Unspecified
	}

	return strconv.FormatUint(uint64(v), 10) + " bits"
}

func (v *KeySizeBits) FromCryptsetupOutput(val string) (err error) {
	if v == nil {
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to map key size (bits) from 'cryptsetup' command output; err: %w", err)
		}
	}()

	if len(val) == 0 {
		*v = 0

		return
	}

	const suffix = " bits"

	if !strings.HasSuffix(val, suffix) {
		return fmt.Errorf("unexpected format '%s'", val)
	}

	valUint64, err := strconv.ParseUint(val[:len(val)-len(suffix)], 10, 16)
	if err != nil {
		return
	}

	if valUint64 > uint64(math.MaxUint16) {
		return fmt.Errorf("too large size '%d'", valUint64)
	}

	*v = KeySizeBits(valUint64)

	return
}
