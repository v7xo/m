package storage

import (
	"fmt"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type MasterDBConf struct {
	Database string `validate:"gt=0"`
	User     string
	Password string
}

func (c *MasterDBConf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	prefix = flenv.NormalizePrefix(flenv.NormalizePrefix(prefix) + MasterDBConfFlenvSelfFlagPrefix)

	flset.StringVar(&c.Database, prefix+"database", "", "master DB name")
	flset.StringVar(&c.User, prefix+"user", "", "master DB user")
	flset.StringVar(&c.Password, prefix+"password", "", "master DB password")

	return nil
}

func (c *MasterDBConf) Validate(v validation.Validator) error {
	return v.Struct(c)
}

// [FlenvRootConf] returns master DB configuration as it would be the root configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvRootMasterDBConf() (cfg MasterDBConf, err error) {
	return FlenvMemberMasterDBConf("")
}

// [FlenvMemberConnConf] returns master DB configuration as it would be the member of parent configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvMemberMasterDBConf(memberPrefix string) (cfg MasterDBConf, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get configuration of master DB; %w", err)
		}
	}()

	err = flenv.PopulatePrefix(memberPrefix, &cfg)

	return
}
