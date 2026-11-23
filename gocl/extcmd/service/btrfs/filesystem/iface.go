package filesystem

import (
	"context"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
)

type FileSystem interface {
	Mount(ctx context.Context, source model.BlockDeviceID, explicitTarget string, opts model.MountOptions) (target string, unmount func(context.Context) error, err error)
	MountV2(ctx context.Context, source model.BlockDeviceID, explicitTarget string, opts model.MountOptions) (mi model.MountInfo, unmount func(context.Context) error, err error)
	Make(ctx context.Context, p MakeParams) error
	Info(ctx context.Context, source model.BlockDeviceID) (model.BtrfsInfo, error)
}
