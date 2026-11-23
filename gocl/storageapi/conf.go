package storageapi

import (
	"time"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type Conf struct {
	Name string `validate:"gt=0,lte=32"`

	CommandTimeoutDur time.Duration `validate:"min=-1ns"`

	MetricsDisabled             bool
	QueryNamedParameterDisabled bool
}

func (c *Conf) CommandTimeout() timeout.Timeout {
	return timeout.Timeout(c.CommandTimeoutDur)
}

func (c *Conf) BindFlagSetPrefix(prefix string, fls *pflag.FlagSet) error {
	prefix = flenv.NormalizePrefix(flenv.NormalizePrefix(prefix) + ConfFlenvSelfFlagPrefix)

	fls.BoolVar(
		&c.MetricsDisabled,
		prefix+"metrics-disabled",
		DefaultMetricsDisabled,
		"disable producing of storage API metrics",
	)
	fls.StringVar(
		&c.Name,
		prefix+"name",
		DefaultName,
		"storage API name. any meaningful name",
	)
	fls.DurationVar(
		&c.CommandTimeoutDur,
		prefix+"command-timeout",
		DefaultCommandTimeout,
		"storage API command timeout",
	)
	fls.BoolVar(
		&c.QueryNamedParameterDisabled,
		prefix+"query-named-parameter-disabled",
		DefaultQueryNamedParameterDisabled,
		"disable use of named parameters in storage API query - positional parameters used instead",
	)

	return nil
}

func (c *Conf) BindFlagSet(fs *pflag.FlagSet) error { return c.BindFlagSetPrefix("", fs) }

func (c *Conf) Validate(v validation.Validator) error { return v.Struct(c) }

// [FlenvRootConf] returns storage connection configuration as it would be the root configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvRootConf() (Conf, error) {
	return FlenvMemberConf("")
}

// [FlenvMemberConf] returns storage connection configuration as it would be the member of parent configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvMemberConf(memberPrefix string) (c Conf, err error) {
	err = flenv.PopulatePrefix(memberPrefix, &c)

	return
}
