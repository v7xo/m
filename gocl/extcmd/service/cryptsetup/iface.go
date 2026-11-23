package cryptsetup

import (
	"context"

	"github.com/google/uuid"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
)

type Cryptsetup interface {
	Close(ctx context.Context, d model.BlockDeviceID) error
	Format(ctx context.Context, p FormatParams) error
	Open(ctx context.Context, p OpenParams) (d model.BlockDeviceID, err error)
	Status(ctx context.Context, d model.BlockDeviceID) (model.CryptDeviceStatus, error)
	UUID(ctx context.Context, d model.BlockDeviceID) (uuid.UUID, error)
}
