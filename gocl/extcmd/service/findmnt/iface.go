package findmnt

import (
	"context"

	"github.com/google/uuid"

	extcmd_model "rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
)

type Findmnt interface {
	BySourceUUIDList(ctx context.Context, uuid uuid.UUID) (extcmd_model.MountInfoList, error)
	List(ctx context.Context) (extcmd_model.MountInfoList, error)
	ByTarget(ctx context.Context, target string) (extcmd_model.MountInfo, bool, error)
	ByPath(ctx context.Context, path string) (extcmd_model.MountInfo, error)
	MountOptionsString(ctx context.Context, target string) (string, error)
	MountOptionValueByName(ctx context.Context, target string, name string) (string, error)
	RootFSTypeCode(context.Context) (extcmd_model.FileSystemTypeCode, error)
}
