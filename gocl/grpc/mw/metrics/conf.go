package metrics

import (
	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type Conf struct {
	Enabled bool
}

func (c *Conf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) (err error) {
	prefix = flenv.NormalizePrefix(prefix)

	flset.BoolVar(
		&c.Enabled,
		prefix+"enabled",
		true,
		"enables/disables this middleware",
	)

	return
}

func (c *Conf) Validate(v validation.Validator) error {
	return v.Struct(c)
}

// [FlenvRootConf] returns metrics gRPC middleware configuration as it would be the root configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvRootConf() (Conf, error) {
	return FlenvMemberConf(ConfFlenvRootFlagPrefix)
}

// [FlenvMemberConf] returns metrics gRPC middleware configuration as it would be the member of parent configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvMemberConf(memberPrefix string) (conf Conf, err error) {
	err = flenv.PopulatePrefix(memberPrefix, &conf)

	return
}
