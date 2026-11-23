//go:build test_unit

package filesystem_test

import (
	"testing"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/btrfs/filesystem"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/btrfs/filesystem_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/findmnt_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/mount"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/mount_thelper"
	hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_Make_Happy(t *testing.T) {
	ctx := hlp.Pre(t)
	findmntS := findmnt_thelper.New(t, ctx)
	mountS := mount_thelper.New(t, ctx, mount.Option().Findmnt(findmntS))

	filesystem_thelper.Make(t, ctx, filesystem.Option().Mount(mountS))
}

func Test_Mount_Happy(t *testing.T) {
	ctx := hlp.Pre(t)
	findmntS := findmnt_thelper.New(t, ctx)
	mountS := mount_thelper.New(t, ctx, mount.Option().Findmnt(findmntS))

	filesystem_thelper.MakeMount(t, ctx, filesystem.Option().Mount(mountS))
}
