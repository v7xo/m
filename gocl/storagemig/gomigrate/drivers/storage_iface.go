package drivers

import (
	"context"

	"github.com/golang-migrate/migrate/v4/database"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/types"
)

//nolint:interfacebloat
type StorageDriver interface {
	database.Driver

	UnsetVersions(ctx context.Context, versions []uint) error
	LatestVersion(ctx context.Context) (VersionDBState, error)
	LatestDeploy(ctx context.Context) (DeployDBState, error)

	EnsureInitialized(ctx context.Context) error
	Initialize(ctx context.Context) (err error)
	IsInitialized(ctx context.Context) (bool, error)

	BeginTx(ctx context.Context) error
	CommitTx(ctx context.Context) error
	RollbackTx(ctx context.Context) error

	// CurrentDeployFingerprint returns fingerprint (check sum) for currently running deploy.
	CurrentDeployFingerprint() types.DeployFingerprint
	IsCurrentDeployRegistered(ctx context.Context) (exists, isRegistered, isDirty bool, err error)
	SetCurrentDeployFingerprint(fp types.DeployFingerprint) error

	RegisterCurrentDeploy(ctx context.Context, isDirty bool) error
	UnregisterCurrentDeploy(ctx context.Context) error

	// VersionsNumsForCurrentDeploy returns versions scoped with current deploy in ascending order.
	VersionsNumsForCurrentDeploy(ctx context.Context) (versions []uint, err error)

	// Log returns logger this service instance writes to.
	// Immutable.
	// Can't be nil.
	Log() logext.Logger
}
