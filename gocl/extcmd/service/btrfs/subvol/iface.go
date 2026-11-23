package subvol

import (
	"context"
	"io"

	"github.com/google/uuid"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	extcmd_model "rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
)

type Subvol interface {
	Create(ctx context.Context, path string) error

	// Deprecated: use [Subvol.DeleteByIDV2] instead.
	DeleteByID(ctx context.Context, id int64, path string) error
	DeleteByIDV2(ctx context.Context, id int64, path string, onlyIfReadOnlyNoChanges bool) error

	IsReadOnly(ctx context.Context, id int64, path string) (readOnly, noChanges bool, err error)

	// Deprecated: use [Subvol.ListV2] instead.
	List(ctx context.Context, mi model.MountInfo, excludeSnapshots bool) (SummaryList, error)
	ListV2(ctx context.Context, mi extcmd_model.MountInfo, filt ListFilter) (SummaryList, error)
	ByUUID(ctx context.Context, mi model.MountInfo, uuid uuid.UUID) (Summary, error)
	ByDirList(ctx context.Context, dir string) (SummaryList, error)

	SnapshotV2(ctx context.Context, subvol, dest string) (Summary, error)

	Receive(ctx context.Context, dest string, r io.Reader) error
	Send(ctx context.Context, parent, subvol string, w io.Writer) error
	Transmit(ctx context.Context, p TransmitParams) (Summary, error)
}
