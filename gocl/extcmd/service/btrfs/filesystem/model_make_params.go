package filesystem

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	modl "rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type MakeParams struct {
	Label          string
	BackingDevice  modl.BlockDeviceIDList
	CSum           modl.BtrfsCSum
	UUID           uuid.UUID
	BlockGroupTree bool
	Force          bool
}

func (v MakeParams) String() string {
	sb := strings.Builder{}

	sb.WriteByte('(')

	w := func(n, v string) {
		if sb.Len() > 1 {
			sb.WriteString(", ")
		}

		sb.WriteString(n)
		sb.WriteString(": ")
		sb.WriteString(v)
	}

	w("label", v.Label)
	w("UUID", v.UUID.String())
	w("device", v.BackingDevice.String())

	sb.WriteByte(')')

	return sb.String()
}

func (v MakeParams) Validate() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w; BTRFS make parameters invalid; err: %w", errs.ErrBadArgs, err)
		}
	}()

	const labelMaxBytes int = 256

	if len(v.Label) == 0 {
		return errors.New("file system label not specified")
	} else if len(v.Label) > labelMaxBytes {
		return fmt.Errorf("file system label length '%d' is greater than '%d'", len(v.Label), labelMaxBytes)
	} else if strings.IndexByte(v.Label, '\n') > -1 {
		return fmt.Errorf("file system label '%s' contains invalid char '%s'", v.Label, string('\n'))
	}

	if v.UUID == uuid.Nil {
		return fmt.Errorf("file system UUID not specified, value '%s' considered as not speicified", uuid.Nil.String())
	}

	if v.CSum == modl.BtrfsCSumUnspecified {
		return errors.New("file system CRC sum algorithm not specified")
	}

	if len(v.BackingDevice) == 0 {
		return errors.New("file system backing device list speifieid as zero-length list")
	}

	for i := range len(v.BackingDevice) {
		if v.BackingDevice[i] == nil {
			return fmt.Errorf("nil device specified at '%d' item", i)
		}
	}

	return
}
