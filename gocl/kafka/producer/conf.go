package producer

import (
	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/kafka"
	"rn.m.vo7.io/vo7io/m/gocl/misc/backoff"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type Conf struct {
	MaxWriteRetries int64 `validate:"gte=0"`
	WriteBackoff    backoff.ExponentialConfig
}

var (
	_ flenv.Conf                = &Conf{}
	_ flenv.FlagSetBinderPrefix = &Conf{}
	_ flenv.UsePreparer         = &Conf{}
)

func (c *Conf) PrepareUse() error { return nil }

func (c *Conf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	prefix = flenv.NormalizeSelfPrefix(prefix, ConfFlenvSelfFlagPrefix)

	if err := c.WriteBackoff.BindFlagSetPrefix(prefix+"write-backoff", flset); err != nil {
		return err
	}

	flset.Int64Var(&c.MaxWriteRetries, prefix+"max-write-retries", 10, "max. count of retries to write message before failing")

	return nil
}

func (c *Conf) Validate(v validation.Validator) error {
	return v.Struct(c)
}

type ClientConf struct {
	Client   kafka.ClientConf
	Producer Conf
}

func (c *ClientConf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	prefix = flenv.NormalizeSelfPrefix(prefix, ConfFlenvSelfFlagPrefix)

	if err := c.Producer.BindFlagSetPrefix(prefix, flset); err != nil {
		return err
	}

	if err := c.Client.BindFlagSetPrefix(prefix, flset); err != nil {
		return err
	}

	return nil
}

func (c ClientConf) Validate(v validation.Validator) error {
	return v.Struct(c)
}
