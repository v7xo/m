//nolint:gci
package gomigrate

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	gomig "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source"
	_ "github.com/golang-migrate/migrate/v4/source/file" // need for registration of FS migration source type
	"github.com/huandu/xstrings"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	clerrs "rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/errs"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/gomigrate/drivers"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/types"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/util"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/mockflow"
)

type (
	implInitState struct {
		sourceAllUpVersions   types.SourceVersionList
		sourceAllDownVersions types.SourceVersionList
		sourceFingerprint     types.SourceFingerprint

		currentDeployFingerprint types.DeployFingerprint
	}

	impl struct {
		fopt.LogOpt
		sourceDriver              source.Driver
		storageDriver             drivers.StorageDriver
		initErr                   error
		initState                 *implInitState
		storageDriverName         types.StorageDriverName
		deployJobID               string
		sourceDriverName          types.SourceDriverName
		init                      sync.Once
		sourceGoTemplatingEnabled bool
	}

	conditionsForUp struct {
		upVersionsNums []uint
		nothingToDo    bool
	}
)

func (s *impl) Create(ctx context.Context, dryRun bool) (err error) {
	err = clerrs.ErrNotImplemented

	return
}

func (s *impl) Initialize(ctx context.Context, dryRun bool) (err error) {
	defer func() {
		if err == nil {
			s.Log().
				Info(ctx).
				Bool(attrname.DryRun, dryRun).
				Msg("DB initialization for migrations succeeded")

			return
		}

		s.Log().
			Err(ctx, err).
			Bool(attrname.DryRun, dryRun).
			Msg("DB initialization for migrations failed")
	}()

	err = s.runInTx(
		ctx,
		dryRun,
		func(ctx context.Context) error {
			return s.storageDriver.Initialize(ctx)
		},
	)

	return
}

func (s *impl) Up(ctx context.Context, dryRun bool) (err error) {
	var upVersionsNums []uint

	defer func() {
		if err != nil {
			err = fmt.Errorf("step UP migration failed; tried to apply versions' num(s): %v; %w", util.FormatVersionNums(upVersionsNums), err)

			s.Log().Err(ctx, err).Bool(attrname.DryRun, dryRun).Send()

			return
		}

		s.Log().
			Info(ctx).
			Bool(attrname.DryRun, dryRun).
			Msgf("step UP migration succeeded; applied versions' num(s): %v", util.FormatVersionNums(upVersionsNums))
	}()

	return s.runInTx(
		ctx,
		dryRun,
		func(ctx context.Context) error {
			var (
				isInitialized bool
				localErr      error
			)

			if isInitialized, localErr = s.storageDriver.IsInitialized(ctx); localErr != nil {
				return localErr
			}

			switch {
			case !isInitialized:
				if localErr = s.storageDriver.Initialize(ctx); localErr != nil {
					return fmt.Errorf("error while initializing DB for migrations; %w", localErr)
				}
			}

			conditions, localErr := s.ensureConditionsForUp(ctx)
			if localErr != nil {
				return localErr
			}

			if conditions.nothingToDo {
				return nil
			}

			upVersionsNums = conditions.upVersionsNums

			if localErr = s.storageDriver.RegisterCurrentDeploy(ctx, true); localErr != nil {
				return localErr
			}

			if len(upVersionsNums) == 0 {
				// unregister empty migration deploy to allow corresponding DOWN-operation to be succeeded (as if it would be executed immediately after current UP-operation).
				//
				return s.storageDriver.UnregisterCurrentDeploy(ctx)
			}

			upSources, downSources, localErr := s.Sources(ctx)
			if localErr != nil {
				return localErr
			}

			gomigsvc, localErr := gomig.NewWithInstance(s.sourceDriverName.String(), &inmemsource{up: upSources, down: downSources}, s.storageDriverName.String(), s.storageDriver)
			if localErr != nil {
				return fmt.Errorf("error while creating instance of Go-migrate; %w", localErr)
			}

			gomigsvc.Log = newLogAdapter(s.Log(), ctx)

			if localErr = gomigsvc.Steps(len(upVersionsNums)); localErr != nil {
				return localErr
			}

			return s.storageDriver.RegisterCurrentDeploy(ctx, false)
		},
	)
}

func (s *impl) Down(ctx context.Context, dryRun bool) (err error) {
	var downVersionsNums []uint

	defer func() {
		if err != nil {
			err = fmt.Errorf("step DOWN migration failed; tried to roll back versions num(s): %v; %w", util.FormatVersionNums(downVersionsNums), err)

			s.Log().Err(ctx, err).Bool(attrname.DryRun, dryRun).Send()

			return
		}

		s.Log().
			Info(ctx).
			Bool(attrname.DryRun, dryRun).
			Msgf("step DOWN migration succeeded; rolled back versions num(s): %v", util.FormatVersionNums(downVersionsNums))
	}()

	return s.runInTx(
		ctx,
		dryRun,
		func(ctx context.Context) error {
			var localErr error

			if _, _, downVersionsNums, localErr = s.ensureConditionsForDown(ctx); localErr != nil || len(downVersionsNums) == 0 {
				return localErr
			}

			if localErr = s.storageDriver.RegisterCurrentDeploy(ctx, true); localErr != nil {
				return localErr
			}

			upSources, downSources, localErr := s.Sources(ctx)
			if localErr != nil {
				return localErr
			}

			gomigsvc, localErr := gomig.NewWithInstance(s.sourceDriverName.String(), &inmemsource{up: upSources, down: downSources}, s.storageDriverName.String(), s.storageDriver)
			if localErr != nil {
				return fmt.Errorf("error while creating instance of Go-migrate component; %w", localErr)
			}

			gomigsvc.Log = newLogAdapter(s.Log(), ctx)

			if localErr = gomigsvc.Steps(-len(downVersionsNums)); localErr != nil {
				return localErr
			}

			if localErr = s.storageDriver.UnsetVersions(ctx, downVersionsNums); localErr != nil {
				return fmt.Errorf("error while unsetting rolled back migrations in DB' versions table; %w", localErr)
			}

			return s.storageDriver.UnregisterCurrentDeploy(ctx)
		},
	)
}

func (s *impl) Status(ctx context.Context) (res types.Status, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error while getting current status; %w", err)
		}
	}()

	var (
		statusDB        types.StorageStatus
		latestVersionDB drivers.VersionDBState
		latestDeployDB  drivers.DeployDBState
	)

	statusDB.LatestDeploy.Empty = true
	statusDB.LatestVersion.Empty = true

	deployFp, sourceFp, up, down, err := s.computeFingerprints(ctx)
	if err != nil {
		return
	}

	if latestVersionDB, err = s.storageDriver.LatestVersion(ctx); err != nil {
		if !errors.Is(err, errs.ErrNotInitialized) {
			return
		}

		err = nil
		statusDB.NotInitialized = true
		statusDB.UpgradableTo = up
	}

	if !statusDB.NotInitialized {
		if latestDeployDB, err = s.storageDriver.LatestDeploy(ctx); err != nil {
			return
		}

		if latestVersionDB.Empty {
			statusDB.LatestVersion.Empty = true
			statusDB.UpgradableTo = up
		} else {
			statusDB.LatestVersion = types.LatestVersionStorageStatus{
				ApplyOrder:        latestVersionDB.ApplyOrder,
				DeployAppRelease:  latestVersionDB.DeployAppRelease,
				DeployFingerprint: latestVersionDB.DeployFingerprint,
				DeployTimestamp:   latestVersionDB.ModifyTs,
				Num:               latestVersionDB.Num,
			}

			for i := range up {
				if up[i].Num <= latestVersionDB.Num {
					continue
				}

				statusDB.UpgradableTo = append(statusDB.UpgradableTo, up[i])
			}
		}

		if latestDeployDB.Empty {
			statusDB.LatestDeploy.Empty = true
		} else {
			statusDB.LatestDeploy = types.LatestDeployStorageStatus{
				ApplyOrder:  latestDeployDB.ApplyOrder,
				AppRelease:  latestDeployDB.AppRelease,
				Fingerprint: latestDeployDB.Fingerprint,
				IsDirty:     latestDeployDB.IsDirty,
				Timestamp:   latestDeployDB.CreateTs,
				Versions: func() []types.LatestDeployStorageStatusVersion {
					res := make([]types.LatestDeployStorageStatusVersion, len(latestDeployDB.Versions))

					for i := range res {
						res[i] = types.LatestDeployStorageStatusVersion{
							Num:        latestDeployDB.Versions[i].Num,
							ApplyOrder: latestDeployDB.Versions[i].ApplyOrder,
						}
					}

					return res
				}(),
			}
		}
	}

	res = types.Status{
		Storage: statusDB,
		Source: types.SourceStatus{
			Fingerprint: sourceFp,
			Up:          up,
			Down:        down,
		},
		CurrentDeployFingerprint: deployFp,
	}

	return
}

func (s *impl) Sources(context.Context) (up types.SourceVersionList, down types.SourceVersionList, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error while reading migrations' sources; %w", err)
		}
	}()

	srcPrevVersNum := -1

	for {
		var (
			srcVersNum  uint
			srcReader   io.ReadCloser
			srcVersName string
		)

		if srcPrevVersNum == -1 {
			if srcVersNum, err = s.sourceDriver.First(); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					err = nil

					break
				}
			}
		} else if srcVersNum, err = s.sourceDriver.Next(uint(srcPrevVersNum)); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				err = nil

				break
			}
		}

		if err != nil {
			return
		}

		readUp := true

	readUpDown:
		{
			var (
				direction = "\"up\""
				srcScript []byte
			)

			switch {
			case readUp:
				if srcReader, srcVersName, err = s.sourceDriver.ReadUp(srcVersNum); err != nil {
					err = fmt.Errorf("error while reading source of '%s' migration version '%d'; %w", direction, srcVersNum, err)

					return
				}
			default:
				direction = "\"down\""

				if srcReader, srcVersName, err = s.sourceDriver.ReadDown(srcVersNum); err != nil {
					err = fmt.Errorf("error while reading source of '%s' migration version '%d'; %w", direction, srcVersNum, err)

					return
				}
			}

			if srcScript, err = io.ReadAll(srcReader); err != nil {
				err = fmt.Errorf("error while reading source script of '%s' migration version '%d'; %w", direction, srcVersNum, err)

				return
			}

			if err = srcReader.Close(); err != nil {
				err = fmt.Errorf("error while closing source reader of '%s' migration version '%d (%s)'; %w", direction, srcVersNum, srcVersName, err)

				return
			}

			switch {
			case s.sourceGoTemplatingEnabled:
				if srcScript, err = s.executeSourceAsGoTextTemplate(srcScript, fmt.Sprintf("'%s' migration version '%d (%s)'", direction, srcVersNum, srcVersName), readUp); err != nil {
					return
				}
			}

			if readUp {
				up = append(
					up,
					types.SourceVersion{
						Num:    srcVersNum,
						Name:   srcVersName,
						Script: string(srcScript),
					},
				)

				readUp = false

				goto readUpDown
			}

			down = append(
				down,
				types.SourceVersion{
					Num:    srcVersNum,
					Name:   srcVersName,
					Script: string(srcScript),
				},
			)
		}

		srcPrevVersNum = int(srcVersNum)
	}

	err = verifyUpDownSources(up, down)

	return
}

func (s *impl) LatestSourceVersionID(context.Context) (string, error) {
	// TODO: implement method.
	return "", clerrs.ErrNotSupported
}

func (*impl) goTextTemplateFuncEnviron(n string) string { return os.Getenv(n) }

func (*impl) goTextTemplateFuncEnvironExists(n string) bool {
	_, exists := os.LookupEnv(n)

	return exists
}

func (s *impl) executeSourceAsGoTextTemplate(sourceScript []byte, sourceDistinguishedName string, upOrDown bool) (outScript []byte, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to execute source script as Go text template; source: %s; %w", sourceDistinguishedName, err)
		}
	}()

	funcMap := sprig.TxtFuncMap()
	if _, ok := funcMap["pascalcase"]; !ok {
		funcMap["pascalcase"] = xstrings.ToPascalCase
	}

	funcMap["environ"] = s.goTextTemplateFuncEnviron
	funcMap["environ_exists"] = s.goTextTemplateFuncEnvironExists

	t := template.New("").Funcs(funcMap)

	goTemplate, err := t.Parse(string(sourceScript))
	if err != nil {
		err = fmt.Errorf("failed to parse source script as Go text template; %w", err)

		return
	}

	dataContext := types.SourceGoTextTemplatingDataContext{
		Up:       upOrDown,
		Down:     !upOrDown,
		MockFlow: mockflow.MockFlow(),
	}

	outBuf := bytes.NewBuffer(make([]byte, 0, len(sourceScript)))

	if err = goTemplate.Execute(outBuf, dataContext); err != nil {
		err = fmt.Errorf("failed to execute Go text template; %w", err)

		return
	}

	outScript = outBuf.Bytes()

	return
}

func verifyUpDownSources(up, down types.SourceVersionList) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("migrations' sources verification revealed an error; %w", err)
		}
	}()

	if len(up) != len(down) {
		err = fmt.Errorf("count of \"up\" migrations not equals to count of \"down\" migrations: %d and %d", len(up), len(down))

		return
	}

	if len(up) == 0 {
		return
	}

	for i := range len(up) {
		if up[i].Num != down[i].Num {
			err = fmt.Errorf("version num of \"up\"-migration not equals to version num of corresponding \"down\" migration: %d (%s) and %d (%s)", up[i].Num, up[i].Name, down[i].Num, down[i].Name)

			return
		}

		switch i {
		case 0:
			if up[i].Num != 1 {
				err = fmt.Errorf("version num of first \"up\" migration not equals to 1: %d (%s)", up[i].Num, up[i].Name)

				return
			}

			continue
		}

		if up[i].Num-up[i-1].Num != 1 {
			err = fmt.Errorf("versions` nums of two neighboring migrations not differ by one: %d (%s) and %d (%s)", up[i-1].Num, up[i-1].Name, up[i].Num, up[i].Name)

			return
		}
	}

	return
}

func (s *impl) ensureConditionsForUpDown(ctx context.Context) (latestDeployDB drivers.DeployDBState, latestVersionDB drivers.VersionDBState, currentDeployVersionsNumsDB []uint, err error) {
	if err = s.initializeService(ctx); err != nil {
		return
	}

	if latestDeployDB, err = s.storageDriver.LatestDeploy(ctx); err != nil {
		return
	} else if latestDeployDB.IsDirty {
		err = fmt.Errorf("%w; DB latest deploy fingerprint: %s", errs.ErrDeployDirty, latestDeployDB)

		return
	}

	if currentDeployVersionsNumsDB, err = s.storageDriver.VersionsNumsForCurrentDeploy(ctx); err != nil {
		return
	}

	if latestVersionDB, err = s.storageDriver.LatestVersion(ctx); err != nil {
		return
	} else if !latestVersionDB.Empty && (latestVersionDB.DeployFingerprint != latestDeployDB.Fingerprint || latestDeployDB.Empty) {
		err = errors.New("DB latest migration version is not part of latest deploy; it might be caused by serious problem related to consistency of migrations metadata in DB")

		return
	}

	return
}

func (s *impl) ensureConditionsForDown(ctx context.Context) (latestDeployDB drivers.DeployDBState, latestVersionDB drivers.VersionDBState, downVersionsNumsDB []uint, err error) {
	if latestDeployDB, latestVersionDB, downVersionsNumsDB, err = s.ensureConditionsForUpDown(ctx); err != nil {
		return
	}

	deployExists, isDeployRegistered, isDeployDirty, err := s.storageDriver.IsCurrentDeployRegistered(ctx)
	if err != nil {
		return
	} else if !isDeployRegistered {
		if !deployExists {
			err = fmt.Errorf("%w; deploy fingerprint: %s", errs.ErrCurrentDeployNotRegistered, s.storageDriver.CurrentDeployFingerprint())

			return
		}

		// deploy exists but not registered; it indicates that DOWN-operation was successfully done previously OR corresponding UP-operation was empty (without new migration).
		//
		downVersionsNumsDB = nil

		return
	} else if isDeployDirty {
		err = fmt.Errorf("%w; deploy fingerprint: %s", errs.ErrCurrentDeployDirty, s.storageDriver.CurrentDeployFingerprint())

		return
	} else if latestDeployDB.Empty || latestDeployDB.Fingerprint != s.storageDriver.CurrentDeployFingerprint() {
		err = fmt.Errorf("%w; current deploy fingerprint: %s; latest deploy: %s", errs.ErrCurrentDeployNotLatest, s.storageDriver.CurrentDeployFingerprint(), latestDeployDB)

		return
	} else if latestVersionDB.Empty || latestVersionDB.DeployFingerprint != s.storageDriver.CurrentDeployFingerprint() {
		err = fmt.Errorf("%w; DB latest migration version: %s", errs.ErrLatestVersionOutOfCurrentDeploy, latestVersionDB)

		return
	} else if len(downVersionsNumsDB) == 0 || downVersionsNumsDB[len(downVersionsNumsDB)-1] != latestVersionDB.Num {
		err = errors.New("DB latest migration version is not latest among all DB versions of current deploy; it might be caused by serious problem related to consistency of migrations metadata in DB")

		return
	}

	return
}

func (s *impl) ensureConditionsForUp(ctx context.Context) (conditionsForUp, error) {
	var (
		res             conditionsForUp
		latestDeploy    drivers.DeployDBState
		latestVersionDB drivers.VersionDBState
		err             error
	)

	if latestDeploy, latestVersionDB, _, err = s.ensureConditionsForUpDown(ctx); err != nil {
		return conditionsForUp{}, err
	}

	var (
		isCurrentDeployDirty      bool
		isCurrentDeployRegistered bool
	)

	if _, isCurrentDeployRegistered, isCurrentDeployDirty, err = s.storageDriver.IsCurrentDeployRegistered(ctx); err != nil {
		return conditionsForUp{}, err
	} else if isCurrentDeployDirty {
		return conditionsForUp{}, fmt.Errorf("%w; deploy fingerprint: %s", errs.ErrCurrentDeployDirty, s.storageDriver.CurrentDeployFingerprint())
	}

	srcAllUpVersionsNums := make([]uint, len(s.initState.sourceAllUpVersions))
	for i := range srcAllUpVersionsNums {
		srcAllUpVersionsNums[i] = s.initState.sourceAllUpVersions[i].Num
	}

	for i := range srcAllUpVersionsNums {
		if srcAllUpVersionsNums[i] > latestVersionDB.Num {
			res.upVersionsNums = srcAllUpVersionsNums[i:]

			break
		}
	}

	srcMaxUpVersion := srcAllUpVersionsNums[len(srcAllUpVersionsNums)-1]

	switch {
	case isCurrentDeployRegistered:
		if latestDeploy.Fingerprint != s.storageDriver.CurrentDeployFingerprint() {
			// attempt to run UP migration when other later migration was applied.
			switch {
			case len(res.upVersionsNums) == 0:
				res.nothingToDo = true

				return res, nil
			}

			return conditionsForUp{},
				fmt.Errorf("%w; current deploy fingerprint: %s; latest deploy: %s",
					errs.ErrLatestVersionOutOfCurrentDeploy,
					s.storageDriver.CurrentDeployFingerprint(),
					latestDeploy)
		}

		if len(res.upVersionsNums) == 0 {
			res.nothingToDo = true

			return res, nil
		}

		return conditionsForUp{}, errors.New("unexpected condition: current deploy is latest registered in storage, but its versions not match sources")
	case latestDeploy.Empty:
		res.upVersionsNums = srcAllUpVersionsNums

		return res, nil
	}

	if len(res.upVersionsNums) == 0 && srcMaxUpVersion < latestVersionDB.Num {
		return conditionsForUp{},
			fmt.Errorf("%w; DB latest migration version: %s; version number of current deploy: %d", errs.ErrCurrentDeployBehindLatestVersion, latestVersionDB, srcMaxUpVersion)
	}

	return res, nil
}

func (s *impl) runInTx(ctx context.Context, dryRun bool, execFn func(context.Context) error) (err error) {
	if err = s.storageDriver.BeginTx(ctx); err != nil {
		return err
	}

	defer func() {
		if dryRun || err != nil {
			if rollbackErr := s.storageDriver.RollbackTx(ctx); rollbackErr != nil {
				err = clerrs.Wrap(rollbackErr, err)

				return
			}

			return
		}

		if commitErr := s.storageDriver.CommitTx(ctx); commitErr != nil {
			err = commitErr

			return
		}
	}()

	return execFn(ctx)
}

func (s *impl) initializeService(ctx context.Context) error {
	s.init.Do(
		func() {
			var err error

			defer func() {
				if err != nil {
					s.initErr = fmt.Errorf("error during DB migration service initialization; %w", err)
				}
			}()

			deployFp, sourceFp, allUpVersions, allDownVersions, err := s.computeFingerprints(ctx)
			if err != nil {
				return
			}

			s.initState = &implInitState{
				currentDeployFingerprint: deployFp,

				sourceAllDownVersions: allDownVersions,
				sourceAllUpVersions:   allUpVersions,
				sourceFingerprint:     sourceFp,
			}

			s.Log().
				Trace(ctx).
				Msgf(
					"migration service initialized:\n\tcurrent deploy fingerprint: %s\n\tsource fingerprint: %s\n\tsource \"up\" versions': %s\n\tsource \"down\" versions': %s;",
					deployFp,
					sourceFp,
					allUpVersions,
					allDownVersions,
				)

			err = s.storageDriver.SetCurrentDeployFingerprint(deployFp)
		},
	)

	return s.initErr
}

//nolint:gocognit,maintidx
func (s *impl) computeFingerprints(ctx context.Context) (deploy types.DeployFingerprint, source types.SourceFingerprint, up types.SourceVersionList, down types.SourceVersionList, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error while computing fingerprints; %w", err)
		}
	}()

	if up, down, err = s.Sources(ctx); err != nil {
		return
	}

	hasher := sha256.New()
	if _, err = hasher.Write([]byte("nop")); err != nil {
		return
	}

	for _, v := range []*struct {
		d string
		s types.SourceVersionList
	}{
		{d: "\"up\"", s: up},
		{d: "\"down\"", s: down},
	} {
		dir, srcs := v.d, v.s

		for i := range srcs {
			src := srcs[i]

			if _, err = fmt.Fprintf(hasher, "%d, %s, %s", src.Num, dir, src.Name); err != nil {
				err = fmt.Errorf("error while computing hash of source of %s migration version %d (%s); %w", dir, src.Num, src.Name, err)

				return
			}

			if _, err = io.Copy(hasher, bytes.NewBufferString(src.Script)); err != nil {
				err = fmt.Errorf("error while computing hash of source of %s migration version %d (%s); %w", dir, src.Num, src.Name, err)

				return
			}
		}
	}

	sourceFpBin := hasher.Sum(nil)

	sourceFpL := binary.BigEndian.Uint64(sourceFpBin[:8])                  //nolint:revive
	sourceFpR := binary.BigEndian.Uint64(sourceFpBin[len(sourceFpBin)-8:]) //nolint:revive
	source = types.SourceFingerprint(sourceFpL ^ sourceFpR)

	if _, err = hasher.Write([]byte(s.deployJobID)); err != nil {
		return
	}

	deployFpBin := hasher.Sum(nil)
	deployFpL := binary.BigEndian.Uint64(deployFpBin[:8])                  //nolint:revive
	deployFpR := binary.BigEndian.Uint64(deployFpBin[len(deployFpBin)-8:]) //nolint:revive
	deploy = types.DeployFingerprint(deployFpL ^ deployFpR)

	return
}

func (s *impl) initLog() error {
	return s.SetLog(
		s.LogCoalesce(s.storageDriver.Log()).
			Tune().
			NextHook(
				func(_ context.Context, _ level.Num, _ error, _ bool, e logext.Event) logext.Event {
					if initState := s.initState; initState != nil {
						e = e.Str(attrname.StorageMigrationDeployFingerprint, initState.currentDeployFingerprint.String())
					}

					return e.Str(attrname.StorageMigrationJobID, s.deployJobID)
				},
			).
			Tuned(),
	)
}

func New(opts ...OptionFn) (_ storagemig.Service, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{}
	if err = fopt.Apply(impl, opts...); err != nil {
		return nil, err
	}

	if impl.storageDriver == nil {
		return nil, errors.New("database driver should be specified and can't be nil")
	}

	if impl.storageDriverName == "" {
		return nil, errors.New("database driver name should be specified and can't be an empty string")
	}

	if impl.sourceDriver == nil {
		return nil, errors.New("migrations source driver should be specified and can't be nil")
	}

	if impl.sourceDriverName == "" {
		return nil, errors.New("migrations source driver should be specified and can't be an empty string")
	}

	if err = impl.initLog(); err != nil {
		return nil, err
	}

	return impl, nil
}
