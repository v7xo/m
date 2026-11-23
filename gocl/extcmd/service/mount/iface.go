package mount

import (
	"context"

	modl "rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
)

type Mount interface {
	Mount(ctx context.Context, p MountParams) (target string, umount func(context.Context) error, err error)
	MountV2(ctx context.Context, p MountParams) (mi modl.MountInfo, umount func(context.Context) error, err error)
	Umount(ctx context.Context, p UmountParams) error
}
