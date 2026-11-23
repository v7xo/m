package storagemig

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/types"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type Conf struct {
	DeployJobID               string        `validate:"gt=0,lt=256"`
	LibString                 string        `validate:"gt=0"`
	MetaInfoDeployObjectName  string        `validate:"gt=0"`
	MetaInfoSchemaName        string        `validate:"gt=0"`
	MetaInfoVersionObjectName string        `validate:"gt=0"`
	StepRunStorageTimeoutDur  time.Duration `validate:"min=-1ns"`
	SourceGoTemplatingEnabled bool
}

var (
	_ flenv.Conf                = &Conf{}
	_ flenv.FlagSetBinder       = &Conf{}
	_ flenv.FlagSetBinderPrefix = &Conf{}
	_ flenv.UsePreparer         = &Conf{}
)

func (c *Conf) StepRunStorageTimeout() timeout.Timeout {
	return timeout.Timeout(c.StepRunStorageTimeoutDur)
}

func (c *Conf) SetStepRunStorageTimeout(v timeout.Timeout) { c.StepRunStorageTimeoutDur = v.Dur() }

func (c *Conf) BindFlagSet(fls *pflag.FlagSet) error { return c.BindFlagSetPrefix("", fls) }

func (c *Conf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	prefix = flenv.NormalizePrefix(flenv.NormalizePrefix(prefix) + ConfFlenvSelfFlagPrefix)

	flset.BoolVar(
		&c.SourceGoTemplatingEnabled,
		prefix+"source-go-templating-enabled",
		false,
		"indicates to treat each migration source script as Go text template",
	)

	flset.StringVar(
		&c.MetaInfoSchemaName,
		prefix+"meta-info-schema-name",
		DefaultMetaInfoSchemaName,
		"storage schema name for migration meta info objects",
	)

	flset.StringVar(
		&c.MetaInfoVersionObjectName,
		prefix+"meta-info-version-object-name",
		DefaultMetaInfoVersionObjectName,
		"storage object name which stores migration version data",
	)

	flset.StringVar(
		&c.MetaInfoDeployObjectName,
		prefix+"meta-info-deploy-object-name",
		DefaultMetaInfoDeployObjectName,
		"storage object name which stores migration deploy data",
	)

	flset.StringVar(
		&c.DeployJobID,
		prefix+"deploy-job-id",
		"",
		"ID of migration job"+
			"\n\tID required to correlate UP and DOWN migration operations (for example, within single CI/CD pipeline)"+
			"\n\tthis ID also used to form deploy fingerprint"+
			"\n\tcan be specified as any string",
	)

	flset.StringVar(
		&c.LibString,
		prefix+"lib",
		string(DefaultLib),
		"migration provider lib"+
			"\n\tone of: '"+string(types.GoMigrateLib)+"'",
	)

	flset.DurationVar(
		&c.StepRunStorageTimeoutDur,
		prefix+"step-run-storage-timeout",
		DefaultStepRunStorageTimeout.Dur(),
		"time to wait for execution of single storage migration step execution"+
			"\n\t"+fmt.Sprintf("use '%v' for infinite waiting", timeout.Infinite.Dur().String()),
	)

	return nil
}

func (c *Conf) Lib() types.Lib                                { return types.Lib(c.LibString) }
func (c *Conf) PrepareUse() error                             { return nil }
func (c *Conf) Validate(validator validation.Validator) error { return validator.Struct(c) }

// [FlenvRootConf] returns default common (base) storage migration service configuration constructed from command flags and environment vars.
func FlenvRootConf() (conf Conf, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to populate configuration '%s' from command flags and environment; %w", stringsext.TypeStringFor[Conf](), err)
		}
	}()

	err = flenv.PopulatePrefix("", &conf)

	return
}
