//go:build test_unit

package mount_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/btrfs/filesystem_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/mount"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/mount_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil_thelper"
	hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_Mount_Btrfs_Happy(t *testing.T) {
	ctx := hlp.Pre(t)

	target := osutil_thelper.MkdirTemp(t, "v7xo-*", osutil.RemoveAll|osutil.RemoveYes)

	params := mount.MountParamsDeviceBase{
		Type:   model.FileSystemTypeCodeBtrfs,
		Target: target,
	}

	params.Options.MustFromString("nodev,nosuid,noexec,ro")

	s := mount_thelper.New(t, ctx)
	_, _, devices := filesystem_thelper.Make(t, ctx)

	params.Source = devices[0]

	_, umount, err := s.Mount(ctx, params)
	require.NoError(t, err, "unexpected error while mounting created BTRFS file system")
	require.NotNil(t, umount)

	err = umount(ctx)
	require.NoError(t, err, "unexpected error while unmounting created BTRFS file system")
}
