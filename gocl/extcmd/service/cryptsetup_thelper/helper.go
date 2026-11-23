package cryptsetup_thelper

import (
	"context"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	modl "rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/cryptsetup"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/losetup"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/losetup_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/uuid_thelper"
	clhlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/value/poly"
)

func New(t *testing.T, ctx context.Context, opts ...cryptsetup.OptionFn) cryptsetup.Cryptsetup {
	t.Helper()

	res, err := cryptsetup.New(fopt.PrependFromContext(ctx, opts...))
	require.NoError(t, err)
	require.NotNil(t, res)

	return res
}

func Open(t *testing.T, ctx context.Context, s cryptsetup.Cryptsetup, params cryptsetup.OpenParams) modl.BlockDeviceID {
	t.Helper()

	opened := OpenCloser(t, ctx, s, params)

	return opened.Device
}

type OpenCloserFixture struct {
	Device       modl.BlockDeviceID
	DeviceCloser func()
	MapperName   string
}

func OpenCloser(t *testing.T, ctx context.Context, s cryptsetup.Cryptsetup, params cryptsetup.OpenParams) OpenCloserFixture {
	t.Helper()

	opened, err := s.Open(ctx, params)
	require.NoError(t, err, "unexpected errror while opening LUKS device")
	require.NotNil(t, opened, "unexpected nil device ID as a result of open operation of LUKS device")
	require.True(t, opened.IsMapper(), "device ID of opened LUKS device is not a mapper device ID")

	res := OpenCloserFixture{
		Device:     opened,
		MapperName: opened.MapperName(),
	}

	cleanup := clhlp.CleanupFn(
		t,
		func() error { return s.Close(ctx, opened) },
		"error while closing LUKS device '%s' (was opened using parameters '%s')",
		opened,
		params,
	)

	closerCalled := int32(0)

	res.DeviceCloser = func() {
		if !atomic.CompareAndSwapInt32(&closerCalled, 0, 1) {
			return
		}

		cleanup()
	}

	t.Cleanup(res.DeviceCloser)

	return res
}

type FormatOpenFixture struct {
	Device       modl.BlockDeviceID
	FormatParams cryptsetup.FormatParamsBase
	OpenParams   cryptsetup.OpenParams
	Backing      losetup_thelper.CreateV2Fixture
	Size         modl.DataSize
}

func FormatOpen(t *testing.T, ctx context.Context, s cryptsetup.Cryptsetup, losetupS losetup.Losetup, size modl.DataSize) FormatOpenFixture {
	t.Helper()

	res := FormatOpenFixture{
		Size:    size,
		Backing: losetup_thelper.CreateNewV2(t, ctx, losetupS, size),
	}

	res.FormatParams = cryptsetup.FormatParamsBase{
		BackingDevice:        res.Backing.Device.ToBlockDeviceID(),
		Cipher:               cryptsetup.DefaultCipherParams(),
		KeyFile:              osutil_thelper.MakeFileTemp(t, "v7xo_*", 16), //nolint:revive,mnd
		Label:                "v7xo_" + uuid_thelper.Rand8BHex(t),
		PBKDFForceIterations: new(uint32(8192)), //nolint:revive,mnd
		PBKDFParallelism:     new(uint8(runtime.NumCPU())),
		UUID:                 uuid_thelper.Rand(t),
	}

	err := s.Format(ctx, res.FormatParams)
	require.NoError(t, err, "unexpected error while formatting LUKS device")

	res.OpenParams = cryptsetup.OpenParams{
		BackingDevice: poly.Of[modl.BlockDeviceID]{Val: res.FormatParams.BackingDevice},
		MapperName:    "v7xo_" + uuid_thelper.Rand4BHex(t),
		Options: cryptsetup.CmdOptions().
			KeyFile(res.FormatParams.KeyFile).
			KeySlot(res.FormatParams.KeySlot).
			MustOptions(),
	}

	res.Device = Open(t, ctx, s, res.OpenParams)

	return res
}

type FormatFixture struct {
	FormatParams cryptsetup.FormatParamsBase
	Backing      losetup_thelper.CreateV2Fixture
	Size         modl.DataSize
}

func Format(t *testing.T, ctx context.Context, s cryptsetup.Cryptsetup, losetupS losetup.Losetup, size modl.DataSize) FormatFixture {
	t.Helper()

	res := FormatFixture{
		Size:    size,
		Backing: losetup_thelper.CreateNewV2(t, ctx, losetupS, size),
	}

	res.FormatParams = cryptsetup.FormatParamsBase{
		BackingDevice:        res.Backing.Device.ToBlockDeviceID(),
		Cipher:               cryptsetup.DefaultCipherParams(),
		KeyFile:              osutil_thelper.MakeFileTemp(t, "v7xo-*", 16), //nolint:mnd,revive
		Label:                "v7xo_" + uuid_thelper.Rand8BHex(t),
		PBKDFForceIterations: new(uint32(8192)), //nolint:mnd,revive
		PBKDFParallelism:     new(uint8(runtime.NumCPU())),
		UUID:                 uuid_thelper.Rand(t),
	}

	err := s.Format(ctx, res.FormatParams)
	require.NoError(t, err, "unexpected error while formatting LUKS device")

	return res
}

type FormatOpenCloserFixture struct {
	Open         OpenCloserFixture
	FormatParams cryptsetup.FormatParamsBase
	OpenParams   cryptsetup.OpenParams
	Backing      losetup_thelper.CreateV2Fixture
	Size         modl.DataSize
}

func FormatOpenCloser(t *testing.T, ctx context.Context, s cryptsetup.Cryptsetup, losetupS losetup.Losetup, size modl.DataSize) FormatOpenCloserFixture {
	t.Helper()

	res := FormatOpenCloserFixture{
		Size:    size,
		Backing: losetup_thelper.CreateNewV2(t, ctx, losetupS, size),
	}

	res.FormatParams = cryptsetup.FormatParamsBase{
		BackingDevice:        res.Backing.Device.ToBlockDeviceID(),
		Cipher:               cryptsetup.DefaultCipherParams(),
		KeyFile:              osutil_thelper.MakeFileTemp(t, "v7xo-*", 16), //nolint:mnd,revive
		Label:                "v7xo_" + uuid_thelper.Rand8BHex(t),
		PBKDFForceIterations: new(uint32(8192)), //nolint:revive,mnd
		PBKDFParallelism:     new(uint8(runtime.NumCPU())),
		UUID:                 uuid_thelper.Rand(t),
	}

	err := s.Format(ctx, res.FormatParams)
	require.NoError(t, err, "unexpected error while formatting LUKS device")

	res.OpenParams = cryptsetup.OpenParams{
		BackingDevice: poly.Of[modl.BlockDeviceID]{Val: res.FormatParams.BackingDevice},
		MapperName:    "v7xo_" + uuid_thelper.Rand4BHex(t),
		Options: cryptsetup.CmdOptions().
			KeySlot(res.FormatParams.KeySlot).
			KeyFile(res.FormatParams.KeyFile).
			MustOptions(),
	}

	res.Open = OpenCloser(t, ctx, s, res.OpenParams)

	return res
}
