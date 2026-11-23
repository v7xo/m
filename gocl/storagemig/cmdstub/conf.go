package cmdstub

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/app/confstub"
	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

// Conf represents migration command configuration. This configuration relates to all migration operations (up, down, init, etc).
type Conf struct {
	StorageAPIConnConfCustomizerFn func(*storageapi.ClientConf) error
	Sources                        fs.FS
	SourcesDir                     string
	MigrationConf                  storagemig.Conf
	confstub.AppBase
	CreateStorage bool
	DryRun        bool
	SkipStatus    bool
}

var (
	_ flenv.Conf                = &Conf{}
	_ flenv.FlagSetBinder       = &Conf{}
	_ flenv.FlagSetBinderPrefix = &Conf{}
	_ flenv.UsePreparer         = &Conf{}
)

func (c *Conf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	if prefix != "" {
		return fmt.Errorf("non-empty prefix '%s' not supported for binding flag set", prefix)
	}

	prefix = flenv.NormalizePrefix(prefix)

	if err := c.AppBase.BindFlagSetPrefix(prefix, flset); err != nil {
		return err
	}

	// flag binding below required to support storage API flag parsing; for ex.: go run . migrate up --flag '<flag-value>' - without binding below will fail.
	if err := (&storageapi.ClientConf{}).BindFlagSetPrefix(prefix, flset); err != nil {
		return err
	}

	if err := c.MigrationConf.BindFlagSetPrefix(prefix, flset); err != nil {
		return err
	}

	additionalFlagPrefix := flenv.NormalizePrefix(flenv.NormalizePrefix(prefix) + storagemig.ConfFlenvSelfFlagPrefix)

	flset.BoolVar(
		&c.CreateStorage,
		additionalFlagPrefix+"-create-storage",
		false,
		"indicates whether to create storage (for ex., database) before run migration"+
			fmt.Sprintf("\n\tapplicable for '%s' or '%s' command only", initSubCmdName, upSubCmdName),
	)

	flset.BoolVar(
		&c.DryRun,
		additionalFlagPrefix+"-dry-run",
		true,
		"indicates whether to run migration without any changes on storage",
	)

	flset.StringVar(
		&c.SourcesDir,
		additionalFlagPrefix+"-sources-dir",
		c.SourcesDir,
		"specifies directory wherein migration sources locate"+
			"\n\tit replaces any default sources predefined by program",
	)

	flset.BoolVar(
		&c.SkipStatus,
		additionalFlagPrefix+"-skip-status",
		false,
		"skip migration status output"+
			fmt.Sprintf("\n\tapplicable for any command excepting '%s'", statusSubCmdName),
	)

	return nil
}

func (c *Conf) BindFlagSet(flset *pflag.FlagSet) error { return c.BindFlagSetPrefix("", flset) }

func (c *Conf) Validate(v validation.Validator) error {
	if c == nil {
		return nil
	}

	if err := v.Struct(c); err != nil {
		return err
	}

	if len(c.SourcesDir) == 0 && c.Sources == nil {
		return errors.New("neither sources directory (see --help) nor sources FS-object specified")
	}

	return nil
}

func (c *Conf) PrepareUse() error {
	if err := c.AppBase.PrepareUse(); err != nil {
		return err
	}

	if err := c.MigrationConf.PrepareUse(); err != nil {
		return err
	}

	return nil
}
