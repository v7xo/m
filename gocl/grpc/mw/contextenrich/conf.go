package contextenrich

import (
	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type Conf struct {
	Enabled bool
}

func (c *Conf) BindFlagSetPrefix(prefix string, fs *pflag.FlagSet) (err error) {
	prefix = flenv.NormalizePrefix(prefix)

	fs.BoolVar(
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

// [FlenvRootConfig] returns context enrichment gRPC middleware configuration as it would be the root configuration object.
//
// Default configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvRootConfig() (cfg Conf, err error) {
	return FlenvMemberConfig(ConfFlenvRootFlagPrefix)
}

// [FlenvMemberConfig] returns context enrichment gRPC middleware configuration as it would be the member of parent configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvMemberConfig(memberPrefix string) (cfg Conf, err error) {
	err = flenv.PopulatePrefix(memberPrefix, &cfg)

	return
}
