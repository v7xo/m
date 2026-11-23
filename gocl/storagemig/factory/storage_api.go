package factory

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	gomig_source "github.com/golang-migrate/migrate/v4/source"
	gomig_iofs "github.com/golang-migrate/migrate/v4/source/iofs"

	logging "rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi/factory"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig"
	gomig "rn.m.vo7.io/vo7io/m/gocl/storagemig/gomigrate"
	cl_mig_gomig_drvs "rn.m.vo7.io/vo7io/m/gocl/storagemig/gomigrate/drivers"
	stordrv "rn.m.vo7.io/vo7io/m/gocl/storagemig/gomigrate/drivers/storage"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/util"
)

func NewServiceSQLAPI(ctx context.Context, sources fs.FS, conf storagemig.Conf, api storageapi.SQLAPI) (svc storagemig.Service, closeSvc func(ctx context.Context) error, err error) {
	var (
		sourcesDriver gomig_source.Driver
		storDriver    cl_mig_gomig_drvs.StorageDriver
	)

	closeSvc = func(context.Context) error {
		var closeErr error

		if storDriver != nil {
			closeErr = errs.Wrap(storDriver.Close(), closeErr)
		}

		if sourcesDriver != nil {
			closeErr = errs.Wrap(sourcesDriver.Close(), closeErr)
		}

		return closeErr
	}

	defer func() {
		if err != nil {
			err = errs.Wrap(closeSvc(ctx), err)
		}
	}()

	switch {
	case conf.Lib().IsUnspecified():
		err = errors.New("migration provider lib not specified by config")

		return
	case !conf.Lib().IsGoMigrate():
		err = errs.Wrap(errs.ErrNotSupported, fmt.Errorf("specified migration provider lib not supported by this migration service; provider lib: %s", conf.Lib()))

		return
	}

	if !api.StorageType().IsPostgreSQL() {
		err = errs.Wrap(errs.ErrNotSupported, fmt.Errorf("specified by storage API storage type not supported; type: %s", api.StorageType()))

		return
	}

	sourcesDriver, err = gomig_iofs.New(sources, ".")
	if err != nil {
		return
	}

	storDriver, err = stordrv.New(
		stordrv.Option().
			StorageAPI(api).
			FallbackContext(ctx).
			LockID(util.CreateStorageLockIDFromStorageName(api.StorageName())).
			MetaInfoSchemaName(conf.MetaInfoSchemaName).
			MetaInfoVersionTableName(conf.MetaInfoVersionObjectName).
			MetaInfoDeployTableName(conf.MetaInfoDeployObjectName).
			StepRunTimeout(conf.StepRunStorageTimeout()).
			Log(api.Log()),
	)
	if err != nil {
		return
	}

	svc, err = gomig.New(
		gomig.
			Option().
			DeployJobID(conf.DeployJobID).
			SourceDriver(storagemig.FileSystemSourceDriverName, sourcesDriver).
			SourceGoTemplatingEnabled(conf.SourceGoTemplatingEnabled).
			StorageDriver(storagemig.PostgreSQLDBDriverName, storDriver).
			Log(api.Log()),
	)

	return
}

func NewServiceSQLAPIConf(
	ctx context.Context,
	sources fs.FS,
	conf storagemig.Conf,
	apiConf storageapi.Conf,
	connConf storage.ConnConf,
	log logging.Logger,
) (svc storagemig.Service, svcClose func(ctx context.Context) error, err error) {
	var (
		apiFactory    factory.FactorySQL
		api           storageapi.SQLAPI
		apiCloser     func()
		svcCloseInner func(ctx context.Context) error
	)

	svcClose = func(ctx context.Context) error {
		var closeErr error

		if svcCloseInner != nil {
			closeErr = svcCloseInner(ctx)
		}

		if apiCloser != nil {
			apiCloser()
		}

		if apiFactory != nil {
			closeErr = errs.Wrap(closeErr, apiFactory.Close(ctx))
		}

		return closeErr
	}

	defer func() {
		if err != nil {
			err = errs.Wrap(err, svcClose(ctx))
		}
	}()

	apiFactory, err = factory.New(
		factory.
			OptionSQLAPI().
			APIConf(apiConf).
			APIConstructor(factory.BaseAPIConstructor).
			ConnConf(connConf).
			Log(log),
	)
	if err != nil {
		return
	}

	api, apiCloser, ctx, err = apiFactory.New(ctx) // nolint:staticcheck
	if err != nil {
		return
	}

	svc, svcCloseInner, err = NewServiceSQLAPI(ctx, sources, conf, api)

	return
}
