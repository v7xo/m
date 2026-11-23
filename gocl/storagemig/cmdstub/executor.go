package cmdstub

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/factory"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/types"
)

type executor struct {
	cmd SubCmdCode
}

func enrichLogEvent(e logext.Event, migrationDryRun bool, enrichers ...func(logext.Event) logext.Event) logext.Event {
	e = e.Bool(attrname.DryRun, migrationDryRun)

	for i := range enrichers {
		e = enrichers[i](e)
	}

	return e
}

//nolint:revive,perfsprint
func logStatus(ctx context.Context, l logext.Logger, migrationDryRun bool, st types.Status, enrichers ...func(logext.Event) logext.Event) {
	buf := []string{}

	buf = append(
		buf,
		"Current deploy (this CI pipeline):",
		fmt.Sprintf("\tFingerprint: %s", st.CurrentDeployFingerprint),
		"Storage:",
		fmt.Sprintf("\tInitialized: %v", !st.Storage.NotInitialized),
		fmt.Sprintf("\tLatest deploy:\n\t\t%s", st.Storage.LatestDeploy),
		fmt.Sprintf("\tLatest version:\n\t\t%s", st.Storage.LatestVersion),
		fmt.Sprintf(
			"\tUpgradable to:\n%s",
			func() string {
				if len(st.Storage.UpgradableTo) == 0 {
					return "\t\t" + placeholder.Empty
				}

				buf := make([]string, 0, len(st.Storage.UpgradableTo)) //nolint:govet
				for y := 0; y < len(st.Storage.UpgradableTo); y++ {
					buf = append(buf, "\t\t"+st.Storage.UpgradableTo[y].String())
				}

				return strings.Join(buf, "\n")
			}(),
		),
		"Current source (this CI pipeline):",
		fmt.Sprintf("\tFingerprint: %s", st.Source.Fingerprint),
		"\tUp versions:",
	)

	for y := 0; y < len(st.Source.Up); y++ {
		buf = append(buf, "\t\t"+st.Source.Up[y].String())
	}

	buf = append(buf, "\tDown versions:")

	for y := 0; y < len(st.Source.Down); y++ {
		buf = append(buf, "\t\t"+st.Source.Down[y].String())
	}

	enrichLogEvent(l.Info(ctx), migrationDryRun, enrichers...).
		Msgf("[ STORAGE MIGRATION STATUS ]:\n%s", strings.Join(buf, "\n"))
}

func (e executor) exec(ctx context.Context, conf *Conf, log logext.Logger) error {
	return ExecuteSubCmd(
		ctx,
		e.cmd,
		conf.Sources,
		conf.SourcesDir,
		conf.MigrationConf,
		conf.CreateStorage,
		conf.DryRun,
		conf.SkipStatus,
		conf.StorageAPIConnConfCustomizerFn,
		log,
	)
}

func ExecuteSubCmd(
	ctx context.Context,
	cmd SubCmdCode,
	sources fs.FS,
	sourcesDir string,
	migrationConf storagemig.Conf,
	createStorage bool,
	dryRun bool,
	skipStatus bool,
	storageAPIConnCfgCustomizerFn func(*storageapi.ClientConf) error,
	log logext.Logger,
) (err error) {
	if dryRun {
		log.Warn(ctx).Msgf("storage migration command '%s' will be exeuted in DRY RUN mode", cmd)
	}

	defer func() {
		if err != nil {
			if dryRun {
				err = fmt.Errorf("storage migration command '%s' failed (DRY RUN); err: %w", cmd, err)
			} else {
				err = fmt.Errorf("storage migration command '%s' failed; err: %w", cmd, err)
			}
		}
	}()

	if len(sourcesDir) != 0 {
		sources = os.DirFS(sourcesDir)
	}

	var (
		svc          storagemig.Service
		svcClose     func(context.Context) error
		logEnrichers []func(logext.Event) logext.Event
	)

	switch {
	case migrationConf.Lib().IsUnspecified():
		err = errors.New("migration lib not specified by config")
	case migrationConf.Lib().IsGoMigrate():
		var storageAPIConnConf storageapi.ClientConf

		if err = flenv.Populate(&storageAPIConnConf); err != nil {
			return
		}

		switch {
		case storageAPIConnCfgCustomizerFn != nil:
			if err = storageAPIConnCfgCustomizerFn(&storageAPIConnConf); err != nil {
				return
			}
		}

		logEnrichers = append(
			logEnrichers,
			func(e logext.Event) logext.Event {
				return storage.EnrichLogWithConnConfProps(e, storageAPIConnConf.Conn)
			},
		)

		svc, svcClose, err = factory.NewServiceSQLAPIConf(ctx, sources, migrationConf, storageAPIConnConf.API, storageAPIConnConf.Conn, log)
	default:
		err = errs.Wrap(errs.ErrNotSupported, fmt.Errorf("specified migration lib not supported; lib: %s", migrationConf.Lib()))
	}

	if err != nil {
		return
	}

	if svcClose != nil {
		defer func() {
			err = errs.Wrap(svcClose(ctx), err)
		}()
	}

	var status types.Status

	switch {
	case !skipStatus || cmd == SubCmdCodeStatus:
		if status, err = svc.Status(ctx); err != nil {
			return
		}

		logStatus(ctx, log, dryRun, status, logEnrichers...)

		if cmd == SubCmdCodeStatus {
			return
		}
	}

	switch cmd {
	case SubCmdCodeUp, SubCmdCodeInit:
		if createStorage {
			err = svc.Create(ctx, dryRun)
		}

		switch err {
		case nil:
			if cmd == SubCmdCodeUp {
				err = svc.Up(ctx, dryRun)
			} else {
				err = svc.Initialize(ctx, dryRun)
			}
		}
	case SubCmdCodeDown:
		err = svc.Down(ctx, dryRun)
	case SubCmdCodeStatus:
		fallthrough
	default:
		err = fmt.Errorf("unknown/unexpected command type code '%s'", cmd)
	}

	if err != nil {
		return
	}

	if !skipStatus {
		if status, err = svc.Status(ctx); err != nil {
			enrichLogEvent(log.Err(ctx, err), dryRun, logEnrichers...).
				Msgf("'%s' command succeeded, but error occurred while getting the current migration status (after command execution)", cmd)
		}
	}

	enrichLogEvent(log.Info(ctx), dryRun, logEnrichers...).
		Msgf("'%s' command succeeded", cmd)

	if !skipStatus {
		logStatus(ctx, log, dryRun, status, logEnrichers...)
	}

	return
}
