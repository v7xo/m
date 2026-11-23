package storagemig

import (
	"context"

	"rn.m.vo7.io/vo7io/m/gocl/storagemig/types"
)

// Service represents migration service API.
// Any operation (method) of this API must conform to idempotency requirements.
type Service interface {
	// CreateStorage creates storage instance (for ex., database).
	Create(ctx context.Context, dryRun bool) error
	// Init initialize storage to support migrations.
	Initialize(ctx context.Context, dryRun bool) error
	// Up upgrades storage with migrations.
	Up(ctx context.Context, dryRun bool) error
	// Down downgrades storage with migrations.
	Down(ctx context.Context, dryRun bool) error
	// Status returns current migration status in storage.
	Status(ctx context.Context) (types.Status, error)
	// Sources gets sources of up & down migrations.
	Sources(ctx context.Context) (up types.SourceVersionList, down types.SourceVersionList, err error)
	// LatestSourceVersionNum returns latest version of migration sources.
	LatestSourceVersionID(ctx context.Context) (string, error)
}
