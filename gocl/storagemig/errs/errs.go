package errs

import "errors"

var (
	ErrCurrentDeployBehindLatestVersion = errors.New("latest storage migration is ahead of current deployment")
	ErrCurrentDeployDirty               = errors.New("state of current deployment in storage is dirty; you need to correct storage state manually")
	ErrCurrentDeployNotLatest           = errors.New("current deployment is not the latest deployment registered in storage")
	ErrCurrentDeployNotRegistered       = errors.New("current deployment not registered")
	ErrDeployDirty                      = errors.New("state of deployment in storage is dirty; you need to correct storage state manually")
	ErrLatestVersionOutOfCurrentDeploy  = errors.New("latest storage migration is not member of current deployment")
	ErrNoSources                        = errors.New("there are no any migration sources")
	ErrNotInitialized                   = errors.New("storage not initialized with migrations")
	ErrVersionDirty                     = errors.New("state of migration version in storage is dirty; you need to correct storage state manually")
)
