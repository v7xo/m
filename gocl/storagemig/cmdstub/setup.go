package cmdstub

import (
	"context"
	"fmt"
	"io/fs"

	"rn.m.vo7.io/vo7io/m/gocl/appcmd/fluentcmd"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
)

func migrateExec(ctx context.Context, _ *Conf, log logext.Logger) error {
	return fmt.Errorf("subcommand is required: %s, %s, %s, %s", SubCmdCodeStatus, SubCmdCodeInit, SubCmdCodeUp, SubCmdCodeDown)
}

func subcmdExec(cmd SubCmdCode) func(ctx context.Context, _ *Conf, logger logext.Logger) error {
	return executor{cmd: cmd}.exec
}

// SetupWithSourcesFSSubcmdOpts creates a migration command wired to the given
// fs.FS sources and custom subcommand options (subcmdOpts). It builds the status,
// init, up and down subcommands, each configured with the provided subcmdOpts,
// and returns a factory function that produces the [fluentcmd.Cmd].
func SetupWithSourcesFSSubcmdOpts(sources fs.FS, subcmdOpts fluentcmd.OptionFn[*Conf]) func() (fluentcmd.Cmd, error) {
	return func() (fluentcmd.Cmd, error) {
		return fluentcmd.
			New(fluentcmd.
				Option[*Conf]().
				IgnoreMissOfConf().
				Use(migrateCmdName).
				Short(migrateCmdShortDesc).
				ChildFn(setupStatus("", sources, subcmdOpts)).
				ChildFn(setupInit("", sources, subcmdOpts)).
				ChildFn(setupUp("", sources, subcmdOpts)).
				ChildFn(setupDown("", sources, subcmdOpts)).
				ExecFn(migrateExec))
	}
}

// SetupWithSourcesFS creates a migration command wired to the given fs.FS sources
// with default (nil) subcommand options. It builds the status, init, up and down
// subcommands and returns a factory function that produces the [fluentcmd.Cmd].
func SetupWithSourcesFS(sources fs.FS) func() (fluentcmd.Cmd, error) {
	return func() (fluentcmd.Cmd, error) {
		return fluentcmd.
			New(fluentcmd.
				Option[*Conf]().
				IgnoreMissOfConf().
				Use(migrateCmdName).
				Short(migrateCmdShortDesc).
				ChildFn(setupStatus("", sources, nil)).
				ChildFn(setupInit("", sources, nil)).
				ChildFn(setupUp("", sources, nil)).
				ChildFn(setupDown("", sources, nil)).
				ExecFn(migrateExec))
	}
}

// SetupWithSubcmdOpts creates a migration command wired to the given sourcesDir
// on disk and custom subcommand options (subcmdOpts). It builds the status, init,
// up and down subcommands, each configured with the provided subcmdOpts, and
// returns a factory function that produces the [fluentcmd.Cmd].
func SetupWithSubcmdOpts(sourcesDir string, subcmdOpts fluentcmd.OptionFn[*Conf]) func() (fluentcmd.Cmd, error) {
	return func() (fluentcmd.Cmd, error) {
		return fluentcmd.
			New(fluentcmd.
				Option[*Conf]().
				IgnoreMissOfConf().
				Use(migrateCmdName).
				Short(migrateCmdShortDesc).
				ChildFn(setupStatus(sourcesDir, nil, subcmdOpts)).
				ChildFn(setupInit(sourcesDir, nil, subcmdOpts)).
				ChildFn(setupUp(sourcesDir, nil, subcmdOpts)).
				ChildFn(setupDown(sourcesDir, nil, subcmdOpts)).
				ExecFn(migrateExec))
	}
}
