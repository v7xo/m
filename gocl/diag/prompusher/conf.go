package prompusher

import (
	"time"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type Conf struct {
	GatewayAddress string        `validate:"gt=0"`
	Interval       time.Duration `validate:"min=1s"`
	Timeout        time.Duration `validate:"min=1ns"`
}

var (
	_ flenv.Conf                = &Conf{}
	_ flenv.FlagSetBinderPrefix = &Conf{}
	_ flenv.UsePreparer         = &Conf{}
)

func (c *Conf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	def := DefaultConf()

	prefix = flenv.NormalizePrefix(prefix)

	flset.StringVar(
		&c.GatewayAddress,
		prefix+"gateway-address",
		def.GatewayAddress,
		"address of Prometheus push gateway"+
			"\n\taddress must include schema, host and port (for ex., http://localhost:9001)")

	flset.DurationVar(
		&c.Interval,
		prefix+"interval",
		def.Interval,
		"push metrics periodically with specified time interval")

	flset.DurationVar(
		&c.Timeout,
		prefix+"timeout",
		def.Timeout,
		"max. timeout of single push operation")

	return nil
}

func (c *Conf) Validate(v validation.Validator) error { return v.Struct(c) }

func (c *Conf) PrepareUse() error { return nil }

// [FlenvRootConf] returns pusher configuration as it would be the root configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvRootConf() (Conf, error) {
	return FlenvMemberConf(ConfFlenvRootFlagPrefix)
}

// [FlenvMemberConf] returns pusher configuration as it would be the member of parent configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvMemberConf(memberPrefix string) (c Conf, err error) {
	err = flenv.PopulatePrefix(memberPrefix, &c)

	return
}
