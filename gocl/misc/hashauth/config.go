package hashauth

import (
	"fmt"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type Config struct {
	SaltHex       string `validate:"gt=0"`
	SecretHashHex string `validate:"gt=0"`
}

func (c *Config) BindFlagSetPrefix(prefix string, fs *pflag.FlagSet) error {
	prefix = flenv.NormalizePrefix(prefix)

	fs.StringVar(
		&c.SaltHex,
		prefix+"salt-hex",
		"",
		"hex-encoded salt used to compute hash of secret")

	fs.StringVar(
		&c.SecretHashHex,
		prefix+"secret-hash-hex",
		"",
		"hex-encoded hash of secret")

	return nil
}

func (c *Config) BindFlags(fs *pflag.FlagSet) error {
	return c.BindFlagSetPrefix("", fs)
}

func (c *Config) Validate(v validation.Validator) error {
	return v.Struct(c)
}

// [FlenvRootConfig] returns hash authentication configuration as it would be the root configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvRootConfig() (cfg Config, err error) {
	return FlenvMemberConfig(ConfigFlagPrefix)
}

// [FlenvMemberConfig] returns hash authentication configuration as it would be the member of parent configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvMemberConfig(memberPrefix string) (cfg Config, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get configuration of hash authentication; %w", err)
		}
	}()

	err = flenv.PopulatePrefix(memberPrefix, &cfg)

	return
}
