package clientappid

import (
	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type Conf struct {
	Filter            string
	BypassMethodRegex string
	Require           bool
	Enabled           bool
}

func (c *Conf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) (err error) {
	prefix = flenv.NormalizePrefix(prefix)

	flset.BoolVar(
		&c.Require,
		prefix+"require",
		DefaultRequired,
		"indicates to require client app ID",
	)
	flset.BoolVar(
		&c.Enabled,
		prefix+"enabled",
		false,
		"enables/disables this middleware",
	)

	flset.StringVar(
		&c.Filter,
		prefix+"filter",
		DefaultFilter,
		"specifies comma separated list of authorized client app ID",
	)
	flset.StringVar(
		&c.BypassMethodRegex,
		prefix+"bypass-method-regex",
		"",
		"specifies regular expression for gRPC methods to be skipped by this middleware",
	)

	return
}

func (c *Conf) Validate(v validation.Validator) error {
	return v.Struct(c)
}

// [FlenvRootConfig] returns middleware configuration as it would be the root configuration object.
//
// Default configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvRootConfig() (conf Conf, err error) {
	err = flenv.PopulatePrefix(ConfFlenvRootFlagPrefix, &conf)

	return
}
