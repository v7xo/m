package cryptsetup

import (
	"errors"
	"fmt"

	"github.com/google/uuid"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type FormatParams interface {
	isFormatParams()

	Validate() error
	String() string
}

type FormatParamsBase struct {
	BackingDevice        model.BlockDeviceID
	KeySlot              *uint8
	PBKDFForceIterations *uint32
	PBKDFParallelism     *uint8
	Label                string
	Cipher               model.CryptCipherParams
	KeyFile              string
	UUID                 uuid.UUID
}

var _ FormatParams = FormatParamsBase{}

func (v FormatParamsBase) isFormatParams() {}

func (v FormatParamsBase) Validate() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w; format parameters invalid; err: %w", errs.ErrBadArgs, err)
		}
	}()

	if v.BackingDevice == nil {
		return errors.New("no backing device specified")
	}

	if v.Label == "" {
		return errors.New("no label specified")
	}

	if v.UUID == uuid.Nil {
		return errors.New("no UUID specified")
	}

	if v.KeyFile == "" {
		return errors.New("no key file specified")
	}

	if v.Cipher.Empty() {
		return errors.New("no cipher params specified")
	}

	return
}

func (v FormatParamsBase) String() string {
	backingDevice := placeholder.Nil
	if v.BackingDevice != nil {
		backingDevice = v.BackingDevice.String()
	}

	return "(label: " + placeholder.UnspecifiedIfEmpty(v.Label) + ", UUID: " + v.UUID.String() + ", backing device: " + backingDevice + ")"
}
