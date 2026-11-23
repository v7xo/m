package filesystem_thelper

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/btrfs/filesystem"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/losetup_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/uuid_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func New(t *testing.T, ctx context.Context, opts ...filesystem.OptionFn) filesystem.FileSystem {
	t.Helper()

	res, err := filesystem.New(fopt.PrependFromContext(ctx, opts...))
	require.NoError(t, err)
	require.NotNil(t, res)

	return res
}

func Make(t *testing.T, ctx context.Context, opts ...filesystem.OptionFn) (s filesystem.FileSystem, params filesystem.MakeParams, devices model.BlockDeviceIDList) {
	t.Helper()

	s = New(t, ctx, opts...)

	fileSysLabel := "label_" + uuid.NewString()
	fileSysUUID := uuid.New()
	fileSysDevSize := DefaultFileSystemSize

	dev1 := losetup_thelper.Create(t, ctx, model.DataSize{Unit: model.DataSizeUnitB, UnitAmount: fileSysDevSize})
	dev2 := losetup_thelper.Create(t, ctx, model.DataSize{Unit: model.DataSizeUnitB, UnitAmount: fileSysDevSize})

	params = filesystem.MakeParams{
		BackingDevice:  model.BlockDeviceIDList{dev1.ToBlockDeviceID(), dev2.ToBlockDeviceID()},
		Label:          fileSysLabel,
		CSum:           model.BtrfsCSumCRC32C,
		UUID:           fileSysUUID,
		BlockGroupTree: true,
		Force:          true,
	}

	err := s.Make(ctx, params)
	require.NoError(t, err, "unexpected error while making BTRFS file system")

	devices = model.BlockDeviceIDList{dev1.ToBlockDeviceID(), dev2.ToBlockDeviceID()}

	return
}

type MakeV2Fixture struct {
	MakeParams filesystem.MakeParams
}

func MakeV2(t *testing.T, ctx context.Context, s filesystem.FileSystem, backing model.BlockDeviceIDList, force bool) MakeV2Fixture {
	t.Helper()

	require.False(t, len(backing) == 0, "backing device list can't be empty")

	res := MakeV2Fixture{
		MakeParams: filesystem.MakeParams{
			BackingDevice:  backing,
			Label:          "v7xo_" + uuid_thelper.Rand4BHex(t),
			UUID:           uuid_thelper.Rand(t),
			CSum:           model.BtrfsCSumCRC32C,
			BlockGroupTree: true,
			Force:          force,
		},
	}

	err := s.Make(ctx, res.MakeParams)
	require.NoError(t, err, "unexpected error while making BTRFS file system on device(s) '%s'", backing)

	return res
}

func MakeMount(t *testing.T, ctx context.Context, opts ...filesystem.OptionFn) (s filesystem.FileSystem, params filesystem.MakeParams, devices model.BlockDeviceIDList, autoTarget string) {
	t.Helper()

	s, params, devices = Make(t, ctx, opts...)

	autoTarget, unmount, err := s.Mount(ctx, devices[0], "", model.EmptyMountOptions())
	require.NoError(t, err, "unexpected error while mounting BTRFS file system")
	require.NotNil(t, unmount, "unexpected nil value of unmount function")
	require.NotEmpty(t, autoTarget, "unexpected empty mount point as a result of mount operation on BTRFS file system")

	thelper.Cleanup(t, func() error { return unmount(ctx) }, "failed to unmount '%s'", autoTarget)

	return
}

type MakeMountV2Fixture struct {
	Opts       model.MountOptions
	Target     string
	MakeParams filesystem.MakeParams
}

func MakeMountV2(t *testing.T, ctx context.Context, s filesystem.FileSystem, backing model.BlockDeviceIDList, forceMake bool) MakeMountV2Fixture {
	t.Helper()

	require.False(t, len(backing) == 0, "backing device list can't be empty")

	res := MakeMountV2Fixture{
		MakeParams: filesystem.MakeParams{
			BackingDevice:  backing,
			Label:          "v7xo_" + uuid_thelper.Rand4BHex(t),
			UUID:           uuid_thelper.Rand(t),
			CSum:           model.BtrfsCSumCRC32C,
			BlockGroupTree: true,
			Force:          forceMake,
		},
		Opts:   model.MustNewMountOptions(model.MountOption().NoDiscard()),
		Target: osutil_thelper.MkdirTemp(t, "v7xo-*", osutil.RemoveAll|osutil.RemoveYes),
	}

	err := s.Make(ctx, res.MakeParams)
	require.NoError(t, err, "unexpected error while making BTRFS file system on device(s) '%s'", backing)

	autoTarget, unmount, err := s.Mount(ctx, backing[0], res.Target, res.Opts)
	require.NoError(t, err, "unexpected error while mounting BTRFS file system")
	require.NotEmpty(t, autoTarget, "unexpected empty mount point as a result of mount operation on BTRFS file system")
	require.Equal(t, autoTarget, res.Target, "unexpected mismatch of mount point as a result of mount operation on BTRFS file system")
	require.NotNil(t, unmount, "unexpected nil value of unmount function as a result of mount operation on BTRFS file system")

	thelper.Cleanup(t, func() error { return unmount(ctx) }, "failed to unmount '%s'", autoTarget)

	return res
}

type MountSubvolFixture struct {
	Target string
	Subvol string
	Info   model.MountInfo
}

func MountSubvol(t *testing.T, ctx context.Context, s filesystem.FileSystem, subvol string, source model.BlockDeviceID, opts model.MountOptions) MountSubvolFixture {
	t.Helper()

	opts, err := opts.With(model.MountOption().BtrfsSubvol(subvol))
	require.NoError(t, err, "unexpected error while enriching mount options with specified BTRFS subvolume '%s'", subvol)

	mi, unmount, err := s.MountV2(ctx, source, "", opts)
	require.NoError(t, err, "unexpected error while mounting '%s' (subvol '%s', opts '%s')", source, subvol, opts)
	require.NotNil(t, unmount, "unexpected nil unmount function as a result of mounting '%s' (subvol '%s', opts '%s')", source, subvol, opts)
	require.NotEmpty(t, mi.Target, "unexpected empty mount target as a result of mounting '%s' (subvol '%s', opts '%s')", source, subvol, opts)

	thelper.Cleanup(
		t,
		func() error { return unmount(ctx) },
		"failed to unmount '%s' (mount of '%s' (subvol '%s', opts '%s'))",
		mi.Target,
		source,
		subvol,
		opts,
	)

	return MountSubvolFixture{Target: mi.Target, Subvol: subvol, Info: mi}
}
