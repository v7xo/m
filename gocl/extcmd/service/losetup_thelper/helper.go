package losetup_thelper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	modl "rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/losetup"
	clcontext "rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil_thelper"
	cl_hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func New(t *testing.T, ctx context.Context, opts ...losetup.OptionFn) losetup.Losetup {
	t.Helper()

	res, err := losetup.New(fopt.PrependFromContext(ctx, losetup.Option()))
	require.NoError(t, err)
	require.NotNil(t, res)

	return res
}

func Create(t *testing.T, ctx context.Context, size modl.DataSize) modl.LoopDeviceInfo {
	t.Helper()

	var (
		log             = clcontext.MustLog(ctx)
		s               = New(t, ctx)
		backingFileName = osutil_thelper.MakeFileTemp(t, "v7xo-*", size.Amount())

		params = losetup.NewParamsFile{
			BackingFile: backingFileName,
			Size:        size,
		}
	)

	devInfo, err := s.New(ctx, params)
	require.NoError(t, err)
	require.True(t, len(devInfo.Name) != 0, "unexpected empty name of created loop device returned")

	cl_hlp.Cleanup(
		t,
		func() error {
			detachErr := s.Detach(ctx, devInfo.Name)
			if detachErr != nil {
				log.Err(ctx, detachErr).Msgf("error occurred while detaching loop device '%s'", devInfo)
			} else {
				log.Info(ctx).Msgf("loop device '%s' attempted to detach", devInfo)
			}

			return detachErr
		},
	)

	require.Equal(t, backingFileName, devInfo.BackingFile, "unexpected mismatch of backing file name")

	return devInfo
}

type CreateV2Fixture struct {
	BackingFile string
	Device      modl.LoopDeviceInfo
	Size        int64
}

func CreateNewV2(t *testing.T, ctx context.Context, s losetup.Losetup, size modl.DataSize) CreateV2Fixture {
	t.Helper()

	res := CreateV2Fixture{
		BackingFile: osutil_thelper.MakeFileTemp(t, "v7xo-*", size.Amount()),
		Size:        size.Amount(),
	}

	var (
		log    = clcontext.MustLog(ctx)
		params = losetup.NewParamsFile{
			BackingFile: res.BackingFile,
			Size:        size,
		}
	)

	devInfo, err := s.New(ctx, params)
	require.NoError(t, err)
	require.True(t, len(devInfo.Name) != 0, "unexpected empty name of created loop device returned")

	cl_hlp.Cleanup(
		t,
		func() error {
			detachErr := s.Detach(ctx, devInfo.Name)
			if detachErr != nil {
				log.Err(ctx, detachErr).Msgf("error occurred while detaching loop device '%s'", devInfo)
			} else {
				log.Info(ctx).Msgf("loop device '%s' attempted to detach", devInfo)
			}

			return detachErr
		},
	)

	require.Equal(t, res.BackingFile, devInfo.BackingFile, "unexpected mismatch of backing file name")

	res.Device = devInfo

	return res
}
