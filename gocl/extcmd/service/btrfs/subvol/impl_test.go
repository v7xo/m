//go:build test_unit

package subvol_test

import (
	"bytes"
	"crypto/sha256"
	"path/filepath"
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	modl "rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/btrfs/filesystem"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/btrfs/filesystem_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/btrfs/subvol"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/btrfs/subvol_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/findmnt_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/mount"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/mount_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ioext_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil_thelper"
	hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_CreateListDelete_Happy(t *testing.T) {
	ctx := hlp.Pre(t)
	findmnt := findmnt_thelper.New(t, ctx)
	s := subvol_thelper.New(t, ctx, subvol.Option().Findmnt(findmnt))
	mount := mount_thelper.New(t, ctx, mount.Option().Findmnt(findmnt))
	_, _, _, mountPoint := filesystem_thelper.MakeMount(t, ctx, filesystem.Option().Mount(mount))

	subvol1BaseName := "subvol1"
	subvol1Path := filepath.Join(mountPoint, subvol1BaseName)

	err := s.Create(ctx, subvol1Path)
	require.NoError(t, err, "unexpected error while creating subvolume '%s'", subvol1Path)

	subvol2BaseName := "subvol2"
	subvol2Path := filepath.Join(mountPoint, subvol2BaseName)

	err = s.Create(ctx, subvol2Path)
	require.NoError(t, err, "unexpected error while creating subvolume '%s'", subvol2Path)

	list, err := s.ByDirList(ctx, mountPoint)
	require.NoError(t, err, "unexpected error while listing subvolumes of '%s'", mountPoint)

	require.Len(t, list, 2, "unexpected length '%d' of created subvolumes list", len(list))

	for i, item := range list {
		require.False(t, item.UUID == uuid.Nil, "unexpected miss of UUID for subvolume '%s'", item)

		list[i] = list[i].
			ResetReachPath().
			ResetUUID().
			ResetSnapshots()
	}

	require.ElementsMatch(
		t,
		list,
		subvol.SummaryList{
			subvol.Summary{
				Generation:         9,
				ID:                 256,
				CreationGeneration: 9,
				ParentID:           5,
				ParentUUID:         uuid.Nil,
				ReceivedUUID:       uuid.Nil,
				TopLevelID:         5,
				UUID:               uuid.Nil,
			}.
				WithPath("subvol1"),
			subvol.Summary{
				Generation:         9,
				ID:                 257,
				CreationGeneration: 9,
				ParentID:           5,
				ParentUUID:         uuid.Nil,
				ReceivedUUID:       uuid.Nil,
				TopLevelID:         5,
				UUID:               uuid.Nil,
			}.
				WithPath("subvol2"),
		},
		"",
	)

	err = s.DeleteByID(ctx, list[0].ID, mountPoint)
	require.NoError(t, err, "unexpected error while deleting subvolume '%s'", list[0])

	err = s.DeleteByID(ctx, list[1].ID, mountPoint)
	require.NoError(t, err, "unexpected error while deleting subvolume '%s'", list[1])

	list, err = s.ByDirList(ctx, mountPoint)
	require.NoError(t, err, "unexpected error while listing subvolumes of '%s'", mountPoint)

	require.Len(t, list, 0, "unexpected length '%d' of subvolumes list", len(list))
}

func Test_SendReceive_Happy(t *testing.T) {
	const tmpFileSize = 1024 * 1024

	ctx := hlp.Pre(t)
	findmnt := findmnt_thelper.New(t, ctx)
	s := subvol_thelper.New(t, ctx, subvol.Option().Findmnt(findmnt))
	mountSvc := mount_thelper.New(t, ctx, mount.Option().Findmnt(findmnt))
	_, _, sourceDevs, sourceMountPoint := filesystem_thelper.MakeMount(t, ctx, filesystem.Option().Mount(mountSvc))
	_, _, recvDevs, recvMountPoint := filesystem_thelper.MakeMount(t, ctx, filesystem.Option().Mount(mountSvc))

	sourceSubvolName := "subvol1"
	sourceSubvolPath := filepath.Join(sourceMountPoint, sourceSubvolName)
	sourceSnapshotsDirPath := filepath.Join(sourceMountPoint, "snapshot")
	sourceSubvolSnapshotPath := filepath.Join(sourceSnapshotsDirPath, sourceSubvolName)

	recvSnapshotsDirPath := filepath.Join(recvMountPoint, "snapshot")

	osutil_thelper.Exec(t, ctx, osutil.ExecParamsNullOutput{Name: "mkdir", Args: []string{sourceSnapshotsDirPath}, RequiresSuperUser: true})
	osutil_thelper.Exec(t, ctx, osutil.ExecParamsNullOutput{Name: "mkdir", Args: []string{recvSnapshotsDirPath}, RequiresSuperUser: true})

	// create a source subvolume.
	err := s.Create(ctx, sourceSubvolPath)
	require.NoError(t, err, "unexpected error while creating subvolume '%s'", sourceSubvolPath)

	// mount source subvolume.
	sourceSubvolMountPoint := mount_thelper.Mount(
		t,
		ctx,
		mountSvc,
		mount.MountParamsDeviceBase{
			Options: modl.MustNewMountOptions(modl.MountOption().BtrfsSubvol(sourceSubvolName).NoDev()),
			Source:  sourceDevs[0],
			Type:    modl.FileSystemTypeCodeBtrfs,
		},
	)

	// create and copy file to source subvolume.
	transferFilePath := osutil_thelper.MakeFileTemp(t, "v7xo-*", tmpFileSize)
	osutil_thelper.Exec(t, ctx, osutil.ExecParamsNullOutput{Name: "cp", Args: []string{"-a", transferFilePath, sourceSubvolMountPoint + "/"}, RequiresSuperUser: true})

	// make snapshot of source subvolume.
	_, err = s.SnapshotV2(ctx, sourceSubvolPath, sourceSubvolSnapshotPath)
	require.NoError(t, err, "unexpected error while making snapshot of '%s' and storing it into '%s'", sourceSubvolPath, sourceSubvolSnapshotPath)

	// retrieve subvolume info of created snapshot of source subvolume.
	sourceSubvolList, err := s.ByDirList(ctx, sourceMountPoint)
	require.NoError(t, err, "unexpected error while listing subvolumes of '%s' (source)", sourceMountPoint)
	require.Len(t, sourceSubvolList, 2, "unexpected length '%d' of subvolume list of '%s' (source)", len(sourceSubvolList), sourceMountPoint)

	sourceSnapshotList := slices.Collect(sourceSubvolList.CreatedAsSnapshot())
	require.Len(t, sourceSnapshotList, 1, "unexpected length '%d' of snapshots list of '%s' (source)", len(sourceSubvolList), sourceMountPoint)
	require.False(
		t,
		sourceSnapshotList[0].UUID == uuid.Nil,
		"unexpected zero UUID of subvolume '%s' of '%s' (source)",
		sourceSnapshotList[0],
		sourceMountPoint,
	)

	sendRecvBuf := bytes.NewBuffer(make([]byte, 0, filesystem_thelper.DefaultFileSystemSize))

	// send snapshot of source subvolume to buffer.
	err = s.Send(ctx, "", sourceSubvolSnapshotPath, sendRecvBuf)
	require.NoError(t, err, "unexpected error while sending snapshot '%s'", sourceSubvolSnapshotPath)

	// receive snapshot of source subvolume from buffer.
	err = s.Receive(ctx, recvSnapshotsDirPath, sendRecvBuf)
	require.NoError(t, err, "unexpected error while receiving snapshot '%s'", sourceSubvolSnapshotPath)

	// list subvolume(s) of receiver.
	recvSubvolList, err := s.ByDirList(ctx, recvMountPoint)
	require.NoError(t, err, "unexpected error while listing subvolumes of '%s' (receiver)", recvMountPoint)
	require.Len(t, recvSubvolList, 1, "unexpected length '%d' of received subvolume list of '%s' (receiver)", len(recvSubvolList), recvMountPoint)
	require.True(
		t,
		recvSubvolList[0].ReceivedUUID == sourceSnapshotList[0].UUID,
		"unexpected mismatch of received UUID of subvolume '%s' of '%s' (receiver)",
		recvSubvolList[0],
		recvMountPoint,
	)

	// mount received snapshot.
	recvSnapshotMountPoint := mount_thelper.Mount(
		t,
		ctx,
		mountSvc,
		mount.MountParamsDeviceBase{
			Options: modl.MustNewMountOptions(modl.MountOption().BtrfsSubvolID(recvSubvolList[0].ID).NoDev().ReadOnly()),
			Source:  recvDevs[0],
			Type:    modl.FileSystemTypeCodeBtrfs,
		},
	)

	// compare.
	transferFileContent, _ := ioext_thelper.ReadAllFile(t, nil, transferFilePath, false)
	recvFileContent, _ := ioext_thelper.ReadAllFile(t, nil, filepath.Join(recvSnapshotMountPoint, filepath.Base(transferFilePath)), false)
	require.Equal(t, sha256.Sum256(transferFileContent), sha256.Sum256(recvFileContent), "unexpected mismatch of received file content to original file content")
}
