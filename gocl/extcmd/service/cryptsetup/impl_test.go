//go:build test_unit

package cryptsetup_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/cryptsetup"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/cryptsetup_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/losetup_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ptrof"
	cluuid "rn.m.vo7.io/vo7io/m/gocl/misc/uuidext"
	hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/value/poly"
)

func Test_Format_BaseParams_Happy(t *testing.T) {
	ctx := hlp.Pre(t)

	s := cryptsetup_thelper.New(t, ctx)
	loopDevice := losetup_thelper.Create(t, ctx, model.DataSize{Unit: model.DataSizeUnitB, UnitAmount: 1024 * 1024 * 64})

	backingDevice := loopDevice.ToBlockDeviceID()
	keySlot := ptrof.Of(uint8(2))
	keyFile := osutil_thelper.MakeFileTemp(t, "v7xo_*", 64)

	formatParams := cryptsetup.FormatParamsBase{
		BackingDevice:        backingDevice,
		Cipher:               model.MustDecodeCryptCipherParams("aes-xts-plain64"),
		KeyFile:              keyFile,
		KeySlot:              keySlot,
		Label:                "test_" + cluuid.MustRand4BHex(),
		PBKDFForceIterations: ptrof.Of(uint32(1500)),
		PBKDFParallelism:     ptrof.Of(uint8(2)),
		UUID:                 uuid.New(),
	}

	err := s.Format(ctx, formatParams)
	require.NoError(t, err, "unexpected error while formatting LUKS device")

	deviceMapperName := "crypt_" + cluuid.MustRand4BHex()
	openParams := cryptsetup.OpenParams{
		BackingDevice: poly.Of[model.BlockDeviceID]{Val: backingDevice},
		MapperName:    deviceMapperName,
		Options: cryptsetup.CmdOptions().
			KeyFile(keyFile).
			KeySlot(keySlot).
			MustOptions(),
	}
	deviceID := cryptsetup_thelper.Open(t, ctx, s, openParams)

	status, err := s.Status(ctx, deviceID)
	require.NoError(t, err, "unexpected error while retrieving status if LUKS device")

	require.Equal(t, deviceID, status.Device())
	require.Equal(t, deviceID.BackingDevice(), status.BackingDevice)
	require.Equal(t, formatParams.Cipher, status.Cipher)
}
