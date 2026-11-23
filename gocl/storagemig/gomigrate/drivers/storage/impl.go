package storage

//nolint:depguard
import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/golang-migrate/migrate/v4/database"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/syncext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
	storage_query "rn.m.vo7.io/vo7io/m/gocl/storage/query"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig"
	mig_errs "rn.m.vo7.io/vo7io/m/gocl/storagemig/errs"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/gomigrate/drivers"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/types"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/util"
)

// nolint:containedctx,govet
type (
	implStateCond uint32

	impl struct {
		fopt.LogOpt
		fallbackCtx context.Context

		storageAPI storageapi.SQLAPI

		lockID string

		metaInfoSchemaName       string
		metaInfoDeployTableName  string
		metaInfoVersionTableName string

		deployGetQuery                 storage_query.Named
		deployRegisterUpdateQuery      storage_query.Named
		versionExistsQuery             storage_query.Named
		versionInsertQuery             storage_query.Named
		versionLatestGetQuery          storage_query.Named
		versionsUnsetQuery             storage_query.Named
		versionsNumsForDeployGetQuery  storage_query.Named
		deployUnregisterQuery          storage_query.Named
		deployLatestRegisteredGetQuery storage_query.Named
		deployRegisterInsertQuery      storage_query.Named
		registeredDeployExistsQuery    storage_query.Named

		lockAcquired atomic.Bool

		stepRunTimeout timeout.Timeout

		lockIDHashCode int64

		currentDeployFingerprint int64

		closeRequestMark    atomic.Int32
		currentDirtyVersion atomic.Int32
	}
)

const (
	inTxStateCond        implStateCond = 1
	initializedStateCond implStateCond = 1 << iota
	hasCurrentDeployFingerprintSetStateCond
	currentDeployRegisteredStateCond
	noCloseRequestStateCond
	allStateConds = inTxStateCond | initializedStateCond | hasCurrentDeployFingerprintSetStateCond | currentDeployRegisteredStateCond | noCloseRequestStateCond

	closeRequestYesMark = int32(1)

	noCurrentDirtyVersionMark = int32(0)

	inconsistentCallToVersionSetMethodErrMsg = `inconsistent call to method setting the version`
)

func (s *impl) EnsureInitialized(ctx context.Context) (err error) {
	if err = s.ensureConditions(ctx, noCloseRequestStateCond); err != nil {
		return
	}

	initialized, _, err := s.ensureInitialized(ctx)
	if err != nil {
		return
	}

	if !initialized {
		err = mig_errs.ErrNotInitialized
	}

	return
}

func (s *impl) IsInitialized(ctx context.Context) (initialized bool, err error) {
	if err = s.ensureConditions(ctx, noCloseRequestStateCond); err != nil {
		return
	}

	initialized, _, err = s.ensureInitialized(ctx)

	return
}

func (s *impl) Initialize(ctx context.Context) (err error) {
	initializedBefore := false

	defer func() {
		if err == nil {
			if initializedBefore {
				s.Log().Info(ctx).Msg("storage meta info initialized already, skip")

				return
			}

			s.Log().Info(ctx).Msg("storage meta info initialization succeeded")

			return
		}

		s.Log().Err(ctx, err).Msg("storage meta info initialization failed")
	}()

	if err = s.ensureConditions(ctx, noCloseRequestStateCond|inTxStateCond); err != nil {
		return
	}

	initializedBefore, schemeExists, err := s.ensureInitialized(ctx)
	if err != nil || initializedBefore {
		return
	}

	if err = s.Lock(); err != nil {
		return
	}

	defer func() {
		if unlockErr := s.Unlock(); unlockErr != nil {
			unlockErr = fmt.Errorf("error while unlocking DB; %w", unlockErr)

			err = errs.Wrap(unlockErr, err)
		}
	}()

	if !schemeExists {
		var query storage_query.Named

		if query, err = storage_query.NewNamed(metaInfoSchemeCreateQueryName, fmt.Sprintf(metaInfoSchemeCreateQueryTemplate, s.metaInfoSchemaName)); err != nil {
			return
		}

		if err = s.storageAPI.ExecSQLNonQuery(ctx, query); err != nil {
			return
		}
	}

	query, err := storage_query.NewNamed(metaInfoVersionTableCreateQueryName, fmt.Sprintf(metaInfoVersionTableCreateQueryTemplate, s.metaInfoSchemaName, s.metaInfoVersionTableName))
	if err != nil {
		return
	}

	if err = s.storageAPI.ExecSQLNonQuery(ctx, query); err != nil {
		return
	}

	if query, err = storage_query.NewNamed(metaInfoDeployTableCreateQueryName, fmt.Sprintf(metaInfoDeployTableCreateQueryTemplate, s.metaInfoSchemaName, s.metaInfoDeployTableName)); err != nil {
		return
	}

	err = s.storageAPI.ExecSQLNonQuery(ctx, query)

	return
}

func (s *impl) BeginTx(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to begin storage tx; %w", err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Info(ctx).Msg("storage tx began")
	}()

	if err = s.ensureConditions(ctx, noCloseRequestStateCond); err != nil {
		return err
	}

	return s.storageAPI.BeginTxIso(ctx, storage.IsolationLevelSerializable)
}

func (s *impl) CommitTx(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to commit storage tx; %w", err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Info(ctx).Msg("storage tx committed")
	}()

	if err = s.ensureConditions(ctx, noCloseRequestStateCond); err != nil {
		return err
	}

	return s.storageAPI.CommitTx(ctx)
}

func (s *impl) RollbackTx(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to roll back storage tx; %w", err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Info(ctx).Msg("storage tx rolled back")
	}()

	if err = s.ensureConditions(ctx, noCloseRequestStateCond); err != nil {
		return err
	}

	return s.storageAPI.RollbackTx(ctx)
}

func (s *impl) Open(_ string) (database.Driver, error) {
	if err := s.ensureConditions(s.fallbackCtx, noCloseRequestStateCond); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *impl) Close() error {
	s.closeRequestMark.Store(closeRequestYesMark)

	return nil
}

func (s *impl) Lock() (err error) {
	const (
		oldLockState = false
		newLockState = true
	)

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to acquire storage lock (lock ID '%d'); %w", s.lockIDHashCode, err)

			s.Log().Err(s.fallbackCtx, err).Send()

			return
		}

		s.Log().Info(s.fallbackCtx).Msgf("succeeded to acquire storage lock (lock ID '%d')", s.lockIDHashCode)
	}()

	if err = s.ensureConditions(s.fallbackCtx, noCloseRequestStateCond|inTxStateCond); err != nil {
		return err
	}

	return syncext.CasRestoreOnErr(
		&s.lockAcquired,
		oldLockState,
		newLockState,
		database.ErrLocked,
		func() error {
			if err := s.ensureConditions(s.fallbackCtx, noCloseRequestStateCond|inTxStateCond); err != nil {
				return err
			}

			if err := s.storageAPI.ExecSQLNonQuery(s.fallbackCtx, lockSetQuery, lockSetQueryParams{LockIDHashCode: s.lockIDHashCode}); err != nil {
				return err
			}

			return nil
		},
	)
}

func (s *impl) Unlock() (err error) {
	const (
		oldLockState = true
		newLockState = false
	)

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to release storage lock (lock ID '%d'); %w", s.lockIDHashCode, err)

			s.Log().Err(s.fallbackCtx, err).Send()

			return
		}

		s.Log().Info(s.fallbackCtx).Msgf("storage lock released (lock ID '%d')", s.lockIDHashCode)
	}()

	if err = s.ensureConditions(s.fallbackCtx, noCloseRequestStateCond|inTxStateCond); err != nil {
		return err
	}

	return syncext.CasRestoreOnErr(
		&s.lockAcquired,
		oldLockState,
		newLockState,
		database.ErrNotLocked,
		func() error { return s.ensureConditions(s.fallbackCtx, noCloseRequestStateCond|inTxStateCond) },
	)
}

func (s *impl) Run(migration io.Reader) error {
	if err := s.ensureConditions(s.fallbackCtx, allStateConds); err != nil {
		return err
	}

	migrationSQLTextBytes, err := io.ReadAll(migration)
	if err != nil {
		return err
	}

	migrationQuery, err := storage_query.NewNamed(migrationRunQueryName, string(migrationSQLTextBytes), storage_query.Option().Timeout(s.stepRunTimeout))
	if err != nil {
		return err
	}

	return s.storageAPI.ExecSQLNonQuery(s.fallbackCtx, migrationQuery)
}

func (s *impl) Version() (versionNum int, isDirty bool, err error) {
	if currentDirtyVersion := s.currentDirtyVersion.Load(); currentDirtyVersion != noCurrentDirtyVersionMark {
		versionNum = int(currentDirtyVersion)
		isDirty = true

		return
	}

	var version drivers.VersionDBState

	if version, err = s.LatestVersion(s.fallbackCtx); err != nil {
		return
	} else if version.Empty {
		versionNum = database.NilVersion

		return
	}

	versionNum = int(version.Num)

	return
}

func (s *impl) LatestVersion(ctx context.Context) (res drivers.VersionDBState, err error) {
	if err = s.ensureConditions(ctx, noCloseRequestStateCond|initializedStateCond); err != nil {
		return
	}

	dbResult := versionState{}
	if err = s.storageAPI.ExecSQLOneRow(ctx, s.versionLatestGetQuery, &dbResult); err != nil {
		if errs.IsNotFound(err) {
			res.Empty = true
			err = nil
		}

		return
	}

	res = drivers.VersionDBState{
		ApplyOrder:        uint(dbResult.ApplyOrder),
		CreateTs:          dbResult.CreatedAt,
		DeployAppRelease:  dbResult.DeployAppRelease,
		DeployFingerprint: types.DeployFingerprint(dbResult.DeployFingerprint),
		IsDeployDirty:     dbResult.IsDeployDirty,
		ModifyTs:          dbResult.ModifiedAt,
		Num:               uint(dbResult.Num),
	}

	return
}

func (s *impl) Drop() (err error) {
	if err = s.ensureConditions(s.fallbackCtx, allStateConds); err != nil {
		return
	}

	err = errs.ErrNotSupported

	return
}

func (s *impl) SetVersion(version int, isDirty bool) (err error) {
	defer func() {
		if err == nil {
			s.Log().Info(s.fallbackCtx).Msgf("set storage migration version as %d (dirty: %v)", version, isDirty)

			return
		}

		s.Log().Err(s.fallbackCtx, err).Msgf("failed to set storage migration version as %d (dirty: %v)", version, isDirty)
	}()

	//nolint:staticcheck
	if version == int(noCurrentDirtyVersionMark) || !(version == database.NilVersion || (version > 0 && version <= math.MaxInt32)) {
		err = fmt.Errorf("specified value of version number is out of range; version num: %d", version)

		return
	}

	if err = s.ensureConditions(s.fallbackCtx, allStateConds); err != nil {
		return
	}

	if (isDirty && !s.currentDirtyVersion.CompareAndSwap(noCurrentDirtyVersionMark, int32(version))) ||
		s.currentDirtyVersion.Load() != int32(version) {
		err = errors.New(inconsistentCallToVersionSetMethodErrMsg)

		return
	}

	if !isDirty {
		defer func() {
			if err == nil {
				s.currentDirtyVersion.CompareAndSwap(int32(version), noCurrentDirtyVersionMark)
			}
		}()
	}

	if version == database.NilVersion {
		// NilVersion denotes empty version set. do nothing, all versions related to current deployment should be deleted in storage by UnsetVersions method.
		//
		return
	}

	var (
		exists = false
		query  = s.versionInsertQuery
		ts     = time.Now()
	)

	if err = s.storageAPI.ExecSQLOneRow(s.fallbackCtx, s.versionExistsQuery, &exists, versionQueryParams{Version: int32(version), DeployFingerprint: s.currentDeployFingerprintInt64()}); err != nil {
		return
	} else if exists {
		return
	}

	err = s.storageAPI.ExecSQLNonQuery(
		s.fallbackCtx,
		query,
		versionQueryParams{
			CreatedAt:         ts,
			DeployFingerprint: s.currentDeployFingerprintInt64(),
			ModifiedAt:        ts,
			Version:           int32(version),
		},
	)

	return
}

func (s *impl) UnsetVersions(ctx context.Context, versions []uint) (err error) {
	defer func() {
		if err == nil {
			s.Log().Info(ctx).Msgf("unset storage migration versions; versions: %s", util.FormatVersionNums(versions))

			return
		}

		s.Log().Err(ctx, err).Msgf("failed to unset storage migration versions; versions: %s", util.FormatVersionNums(versions))
	}()

	if len(versions) == 0 {
		err = errors.New("version set can't be empty")

		return
	}

	versionsAsInt32 := make([]int32, len(versions))

	for i := range versions {
		if versions[i] > math.MaxInt32 {
			err = fmt.Errorf("specified value of version number at position %d is out of range; version num: %d", i, versions[i])

			return
		}

		versionsAsInt32[i] = int32(versions[i])
	}

	if err = s.ensureConditions(ctx, allStateConds); err != nil {
		return
	}

	var versionsUnsetCount int

	if err = s.storageAPI.ExecSQLOneRow(ctx, s.versionsUnsetQuery, &versionsUnsetCount, versionsUnsetQueryParams{Versions: versionsAsInt32, DeployFingerprint: s.currentDeployFingerprintInt64()}); err != nil {
		return
	}

	if versionsUnsetCount != len(versionsAsInt32) {
		err = fmt.Errorf("unexpected result of storage operation: count of unset migration versions in storage mismatches expected one; actual count: %d; expected count: %d", versionsUnsetCount, len(versionsAsInt32))

		return
	}

	return
}

func (s *impl) VersionsNumsForCurrentDeploy(ctx context.Context) ([]uint, error) {
	if err := s.ensureConditions(ctx, noCloseRequestStateCond|initializedStateCond|hasCurrentDeployFingerprintSetStateCond); err != nil {
		return nil, err
	}

	var versions []int

	if err := s.storageAPI.ExecSQL(ctx, s.versionsNumsForDeployGetQuery, &versions, versionsNumsForDeployGetQueryParams{DeployFingerprint: s.currentDeployFingerprintInt64()}); err != nil {
		return nil, err
	}

	versionsUint := make([]uint, len(versions))
	for i := 0; i < len(versions); i++ {
		versionsUint[i] = uint(versions[i])
	}

	return versionsUint, nil
}

func (s *impl) LatestDeploy(ctx context.Context) (res drivers.DeployDBState, err error) {
	if err = s.ensureConditions(ctx, noCloseRequestStateCond|initializedStateCond); err != nil {
		return
	}

	dbResult := []deployState{}
	if err = s.storageAPI.ExecSQL(ctx, s.deployLatestRegisteredGetQuery, &dbResult); err != nil {
		return
	}

	if len(dbResult) == 0 || dbResult[0].IsUnregistered {
		res.Empty = true

		return
	}

	res = drivers.DeployDBState{
		ApplyOrder:  uint(dbResult[0].ApplyOrder),
		AppRelease:  dbResult[0].AppRelease,
		CreateTs:    dbResult[0].CreatedAt,
		Fingerprint: types.DeployFingerprint(dbResult[0].Fingerprint),
		IsDirty:     dbResult[0].IsDirty,
		ModifyTs:    dbResult[0].ModifiedAt,
		Versions: func() []drivers.DeployDBStateVersion {
			if dbResult[0].NoVersion {
				return nil
			}

			res := make([]drivers.DeployDBStateVersion, len(dbResult))
			for i := range res {
				res[i] = drivers.DeployDBStateVersion{
					Num:        uint(dbResult[i].VersionNum),
					ApplyOrder: uint(dbResult[i].VersionApplyOrder),
				}
			}

			return res
		}(),
	}

	return
}

func (s *impl) CurrentDeployFingerprint() types.DeployFingerprint {
	return types.DeployFingerprint(s.currentDeployFingerprintInt64())
}

func (s *impl) SetCurrentDeployFingerprint(fp types.DeployFingerprint) (err error) {
	if fp == storagemig.UnassignedDeployFingerprint {
		err = fmt.Errorf("specified deploy fingerprint is not allowed to be set; fingerprint: %s", fp)

		return
	}

	if !atomic.CompareAndSwapInt64(&s.currentDeployFingerprint, int64(storagemig.UnassignedDeployFingerprint), int64(fp)) {
		err = errors.New("current deploy fingerptint can't be changed; it can be set only once")
	}

	return
}

func (s *impl) IsCurrentDeployRegistered(ctx context.Context) (exists, isRegistered, isDirty bool, err error) {
	exists, isRegistered, isDirty, err = s.fetchCurrentDeployDataFromDB(ctx)

	return
}

func (s *impl) RegisterCurrentDeploy(ctx context.Context, isDirty bool) (err error) {
	var (
		exists, alreadyRegistered bool
		deployAppRelease          = util.AppRelease()
	)

	defer func() {
		regActionMsg := "register"
		if alreadyRegistered {
			regActionMsg = "update"
		}

		dirtyMsg := "as non-dirty"
		if isDirty {
			dirtyMsg = "as dirty"
		}

		if err != nil {
			err = fmt.Errorf("failed to %s current deploy %s; deploy fingerprint: %s; %w", regActionMsg, dirtyMsg, s.CurrentDeployFingerprint(), err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Info(ctx).Msgf("succeeded to %s current deploy %s; deploy fingerprint: %s", regActionMsg, dirtyMsg, s.CurrentDeployFingerprint())
	}()

	if err = s.ensureConditions(ctx, noCloseRequestStateCond|initializedStateCond|hasCurrentDeployFingerprintSetStateCond|inTxStateCond); err != nil {
		return
	}

	exists, alreadyRegistered, _, err = s.fetchCurrentDeployDataFromDB(ctx)
	if err != nil {
		return
	}

	var (
		ts    = time.Now()
		query = s.deployRegisterInsertQuery
	)

	if exists {
		query = s.deployRegisterUpdateQuery
	}

	return s.storageAPI.ExecSQLNonQuery(
		ctx,
		query,
		deployQueryParams{
			AppRelease:  deployAppRelease,
			CreatedAt:   ts,
			Fingerprint: s.currentDeployFingerprintInt64(),
			IsDirty:     isDirty,
			ModifiedAt:  ts,
		},
	)
}

func (s *impl) UnregisterCurrentDeploy(ctx context.Context) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to unregister current deploy; deploy fingerprint: %s; %w", s.CurrentDeployFingerprint(), err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Info(ctx).Msgf("succeeded to unregister current deploy; deploy fingerprint: %s", s.CurrentDeployFingerprint())
	}()

	if err = s.ensureConditions(ctx, allStateConds); err != nil {
		return err
	}

	ts := time.Now()

	return s.storageAPI.ExecSQLNonQuery(
		ctx,
		s.deployUnregisterQuery,
		deployQueryParams{
			ModifiedAt:  ts,
			CreatedAt:   ts,
			Fingerprint: s.currentDeployFingerprintInt64(),
		},
	)
}

func (s *impl) ensureConditions(ctx context.Context, conds implStateCond) (err error) {
	if (conds & noCloseRequestStateCond) == noCloseRequestStateCond {
		if err = s.ensureNoCloseRequest(); err != nil {
			return
		}
	}

	if ctx == nil {
		ctx = s.fallbackCtx
	}

	if (conds & hasCurrentDeployFingerprintSetStateCond) == hasCurrentDeployFingerprintSetStateCond {
		if s.CurrentDeployFingerprint() == storagemig.UnassignedDeployFingerprint {
			err = errors.New("current deploy fingerprint has not set yet, it required to be set before to perform requested operation")

			return
		}
	}

	if (conds & inTxStateCond) == inTxStateCond {
		if err = s.ensureInTx(ctx); err != nil {
			return
		}
	}

	if (conds & initializedStateCond) == initializedStateCond {
		if err = s.EnsureInitialized(ctx); err != nil {
			return
		}
	}

	if (conds & currentDeployRegisteredStateCond) == currentDeployRegisteredStateCond {
		err = s.ensureCurrentDeployRegistered(ctx)
	}

	return
}

func (s *impl) ensureCurrentDeployRegistered(ctx context.Context) (err error) {
	var exists bool

	if err = s.storageAPI.ExecSQLOneRow(ctx, s.registeredDeployExistsQuery, &exists, deployQueryParams{Fingerprint: s.currentDeployFingerprintInt64()}); err != nil {
		return
	}

	if !exists {
		err = fmt.Errorf("%w; deploy fingerprint: %s", mig_errs.ErrCurrentDeployNotRegistered, s.CurrentDeployFingerprint())
	}

	return
}

func (s *impl) ensureInitialized(ctx context.Context) (initialized, schemeExists bool, err error) {
	checkResult := struct {
		TableCount   int  `db:"table_count"`
		SchemeExists bool `db:"schema_exists"`
	}{}

	metaInfoTableNames := []string{s.metaInfoVersionTableName, s.metaInfoDeployTableName}

	if err = s.storageAPI.ExecSQLOneRow(
		ctx,
		metaInfoExistenceCheckQuery,
		&checkResult,
		metaInfoExistenceCheckQueryParams{
			TableSchema: s.metaInfoSchemaName,
			TableNames:  metaInfoTableNames,
		},
	); err != nil {
		return
	}

	schemeExists = checkResult.SchemeExists

	if !schemeExists {
		return
	}

	initialized = checkResult.TableCount == len(metaInfoTableNames)

	return
}

func (s *impl) ensureInTx(ctx context.Context) (err error) {
	var txID1, txID2 int

	if err = s.storageAPI.ExecSQLOneRow(ctx, currentTxIDQuery, &txID1); err != nil {
		return
	}

	if err = s.storageAPI.ExecSQLOneRow(ctx, currentTxIDQuery, &txID2); err != nil {
		return
	}

	if txID1 == txID2 {
		return
	}

	err = errs.ErrTxRequired

	return
}

func (s *impl) ensureNoCloseRequest() error {
	if s.closeRequestMark.Load() == closeRequestYesMark {
		return fmt.Errorf("close has been requested fot this instance; instance: %v; %w", s, errs.ErrCloseRequested)
	}

	return nil
}

func (s *impl) fetchCurrentDeployDataFromDB(ctx context.Context) (exists, isRegistered, isDirty bool, err error) {
	if err = s.ensureConditions(ctx, noCloseRequestStateCond|initializedStateCond|hasCurrentDeployFingerprintSetStateCond); err != nil {
		return
	}

	deploy := deployState{}
	if err = s.storageAPI.ExecSQLOneRow(ctx, s.deployGetQuery, &deploy, deployQueryParams{Fingerprint: s.currentDeployFingerprintInt64()}); err != nil {
		if errs.IsNotFound(err) {
			err = nil
		}

		return
	}

	exists = true
	isRegistered = !deploy.IsUnregistered
	isDirty = deploy.IsDirty

	return
}

func (s *impl) currentDeployFingerprintInt64() int64 {
	return atomic.LoadInt64(&s.currentDeployFingerprint)
}

func (s *impl) setupLockIDHashCode() (err error) {
	const lockIDDriverMaterial = `gocl_storage_migration_driver`

	defer func() {
		if err != nil {
			err = fmt.Errorf("error while computing lock ID hash code; %w", err)
		}
	}()

	if s.lockID == "" {
		err = errors.New("lock ID should be specified and can't be an empty string")

		return
	}

	lockIDHashCodeStr, err := GenerateAdvisoryLockId(s.lockID, lockIDDriverMaterial)
	if err != nil {
		return
	}

	lockIDHashCode, err := strconv.ParseInt(lockIDHashCodeStr, base10, bitSize64)
	if err != nil {
		return
	}

	s.lockIDHashCode = lockIDHashCode

	return
}

func (s *impl) setupQuieries() (err error) {
	defer func() {
		if recoveryObj := recover(); recoveryObj != nil {
			if recoverErr, ok := recoveryObj.(error); ok {
				err = errs.Wrap(recoverErr, err)

				return
			}

			err = errs.Wrap(fmt.Errorf("recovered from panic; %v", recoveryObj), err)
		}
	}()

	s.versionExistsQuery = storage_query.MustNewNamed(versionExistsQueryName, fmt.Sprintf(versionExistsQueryTemplate, s.metaInfoSchemaName, s.metaInfoVersionTableName))
	s.versionInsertQuery = storage_query.MustNewNamed(versionInsertQueryName, fmt.Sprintf(versionInsertQueryTemplate, s.metaInfoSchemaName, s.metaInfoVersionTableName))
	s.versionLatestGetQuery = storage_query.MustNewNamed(versionLatestGetQueryName, fmt.Sprintf(versionLatestGetQueryTemplate, s.metaInfoSchemaName, s.metaInfoVersionTableName, s.metaInfoDeployTableName))
	s.versionsNumsForDeployGetQuery = storage_query.MustNewNamed(versionsNumsForDeployGetQueryName, fmt.Sprintf(versionsNumsForDeployGetQueryTemplate, s.metaInfoSchemaName, s.metaInfoVersionTableName, s.metaInfoDeployTableName))
	s.versionsUnsetQuery = storage_query.MustNewNamed(versionsUnsetQueryName, fmt.Sprintf(versionsUnsetQueryTemplate, s.metaInfoSchemaName, s.metaInfoVersionTableName))

	s.deployGetQuery = storage_query.MustNewNamed(deployGetQueryName, fmt.Sprintf(deployGetQueryTemplate, s.metaInfoSchemaName, s.metaInfoDeployTableName))
	s.deployLatestRegisteredGetQuery = storage_query.MustNewNamed(deployLatestRegisteredGetQueryName, fmt.Sprintf(deployLatestRegisteredGetQueryTemplate, s.metaInfoSchemaName, s.metaInfoDeployTableName, s.metaInfoVersionTableName))
	s.deployRegisterInsertQuery = storage_query.MustNewNamed(deployRegisterInsertQueryName, fmt.Sprintf(deployRegisterInsertQueryTemplate, s.metaInfoSchemaName, s.metaInfoDeployTableName))
	s.deployRegisterUpdateQuery = storage_query.MustNewNamed(deployRegisterUpdateQueryName, fmt.Sprintf(deployRegisterUpdateQueryTemplate, s.metaInfoSchemaName, s.metaInfoDeployTableName))
	s.deployUnregisterQuery = storage_query.MustNewNamed(deployUnregisterQueryName, fmt.Sprintf(deployUnregisterQueryTemplate, s.metaInfoSchemaName, s.metaInfoDeployTableName))
	s.registeredDeployExistsQuery = storage_query.MustNewNamed(registeredDeployExistsQueryName, fmt.Sprintf(registeredDeployExistsQueryTemplate, s.metaInfoSchemaName, s.metaInfoDeployTableName))

	return
}

func (s *impl) initLog() error {
	log := s.Log()
	if log.IsNop() {
		log = s.storageAPI.Log()
	}

	log = log.Tune().
		NextHook(
			func(_ context.Context, _ level.Num, _ error, _ bool, e logext.Event) logext.Event {
				return e.
					Int64(attrname.StorageMigrationDeployFingerprint, s.currentDeployFingerprintInt64()).
					Str(attrname.StorageMigrationLockID, s.lockID)
			},
		).
		Tuned()

	return s.SetLog(log)
}

func New(opts ...OptionFn) (_ drivers.StorageDriver, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{}
	if err = fopt.Apply(impl, opts...); err != nil {
		return nil, err
	}

	if impl.stepRunTimeout.IsZero() {
		return nil, errors.New("migration step run timeout can't be specified as zero")
	}

	if impl.storageAPI == nil {
		return nil, errors.New("storage API should be specified")
	}

	if !impl.storageAPI.StorageType().IsPostgreSQL() {
		return nil, errors.New("specified storage API is not PostgreSQL storage API")
	}

	if err = impl.setupLockIDHashCode(); err != nil {
		return nil, err
	}

	if err = impl.setupQuieries(); err != nil {
		return nil, err
	}

	if impl.fallbackCtx == nil {
		impl.fallbackCtx = context.Background()
	}

	if err = impl.initLog(); err != nil {
		return nil, err
	}

	return impl, nil
}

func MustNew(opts ...OptionFn) drivers.StorageDriver {
	res, err := New(opts...)
	if err != nil {
		panic(err)
	}

	return res
}
