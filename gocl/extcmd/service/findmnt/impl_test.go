//go:build test_unit

package findmnt_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/btrfs/filesystem"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/btrfs/filesystem_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/findmnt_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/mount_thelper"
	hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_RootFSTypeCode_Happy(t *testing.T) {
	ctx := hlp.Pre(t)
	s := findmnt_thelper.New(t, ctx)

	fsType, err := s.RootFSTypeCode(ctx)
	require.NoError(t, err, "unexpected error while calling function to get root file system type")
	require.NotEqual(t, model.FileSystemTypeCodeOther.String(), fsType, "unexpected result (root file system type name)")
	require.NotEqual(t, model.FileSystemTypeCodeUnspecified.String(), fsType, "unexpected result (root file system type name)")
}

func Test_List_Happy(t *testing.T) {
	ctx := hlp.Pre(t)
	s := findmnt_thelper.New(t, ctx)

	list, err := s.List(ctx)
	require.NoError(t, err, "unexpected error while calling function to list mounted file systems")
	require.GreaterOrEqual(t, len(list), 1, "unexpected zero-length list of mounted file systems")
}

func Test_BySourceUUIDList_Happy(t *testing.T) {
	ctx := hlp.Pre(t)

	findmntService := findmnt_thelper.New(t, ctx)
	mountService := mount_thelper.New(t, ctx)
	_, fsMkParams, fsDevices, mountPoint := filesystem_thelper.MakeMount(t, ctx, filesystem.Option().Mount(mountService))

	found, err := findmntService.BySourceUUIDList(ctx, fsMkParams.UUID)
	require.NoError(t, err, "unexpected error while finding mount(s)")
	require.Len(t, found, 1, "count of found mounts doesn't equal to expected")

	require.Equal(t, fsMkParams.UUID, found[0].FileSystemUUID, "found file system UUID doesn't match to expected")
	require.Equal(t, mountPoint, found[0].Target, "found file system mount point doesn't match to expected")
	require.NotNil(t, found[0].Source, "nil source (backing device) of found file system")
	require.Contains(t, fsDevices.ToPathSlice(), found[0].Source.Path(), "source (backing device) of found file system is not in expected list")
}
