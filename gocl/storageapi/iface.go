package storageapi

import (
	"context"

	logging "rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
)

type API interface {
	// Close closes this API instance.
	// Close executes any operations related to disposal of resources used by this API instance.
	Close(ctx context.Context) error

	// CloseWithLog same as [API.Close] but logs error occurred during Close.
	// If logger not specified here then error will be logged using logger returned by [API.Log].
	CloseWithLog(ctx context.Context, l logging.Logger)

	// IsOperable indicates whether this API instance can handle queries (not closed, not busy, etc).
	IsOperable() bool

	// Name returns this API instance name.
	Name() string

	// StorageType returns storage type this API instance works with.
	StorageType() storage.StorageType

	// StorageName returns name of storage this API instance configured with.
	StorageName() string

	// Ping checks connectivity with storage.
	Ping(ctx context.Context) error

	// Log returns logger this API instance writes to.
	// Can't be nil.
	Log() logging.Logger
}
