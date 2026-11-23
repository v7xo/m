package lsblk

import (
	"context"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
)

type Lsblk interface {
	List(ctx context.Context, filt Filter) (model.BlockDeviceInfoList, error)
	Get(ctx context.Context, filt Filter) (model.BlockDeviceInfo, error)
}
