package storageapi

import (
	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

// ClientConf is a stub for configuration of Storage API with Storage connection config.
type ClientConf struct {
	API  Conf
	Conn storage.ConnConf
}

func (c *ClientConf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	if err := c.Conn.BindFlagSetPrefix(prefix, flset); err != nil {
		return err
	}

	if err := c.API.BindFlagSetPrefix(prefix, flset); err != nil {
		return err
	}

	return nil
}

func (c *ClientConf) BindFlagSet(flset *pflag.FlagSet) error {
	return c.BindFlagSetPrefix("", flset)
}

func (c *ClientConf) Validate(v validation.Validator) error {
	return v.Struct(c)
}

// [FlenvRootClientConf] returns storage API connection configuration as it would be the root configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvRootClientConf() (ClientConf, error) {
	return FlenvMemberClientConf("")
}

// [FlenvMemberClientConf] returns storage API connection as it would be the member of parent configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvMemberClientConf(memberPrefix string) (c ClientConf, err error) {
	err = flenv.PopulatePrefix(memberPrefix, &c)

	return
}
