package losetup

import (
	"context"

	modl "rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
)

type Losetup interface {
	Attach(ctx context.Context, p AttachParams) (modl.LoopDeviceInfo, error)
	Detach(ctx context.Context, n string) error
	List(ctx context.Context) (modl.LoopDeviceInfoList, error)
	New(ctx context.Context, p NewParams) (modl.LoopDeviceInfo, error)
}
