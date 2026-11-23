package consumer

import (
	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/kafka"
	"rn.m.vo7.io/vo7io/m/gocl/misc/backoff"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type Conf struct {
	ConsumerGroupID   string `validate:"min=0"`
	Topic             string `validate:"min=0"`
	Client            kafka.ClientConf
	FetchBackoff      backoff.ExponentialConfig
	FetchMaxRetries   int64 `validate:"gte=0"`
	EmitReaderInfoLog bool
}

var (
	_ flenv.Conf                = &Conf{}
	_ flenv.FlagSetBinderPrefix = &Conf{}
	_ flenv.UsePreparer         = &Conf{}
)

func (c *Conf) PrepareUse() error { return nil }

func (c *Conf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	prefix = flenv.NormalizePrefix(flenv.NormalizePrefix(prefix) + ConfFlenvSelfFlagPrefix)

	if err := c.FetchBackoff.BindFlagSetPrefix(prefix+"fetch-backoff", flset); err != nil {
		return err
	}

	if err := c.Client.BindFlagSetPrefix(prefix, flset); err != nil {
		return err
	}

	flset.Int64Var(&c.FetchMaxRetries, prefix+"fetch-max-retries", 2, "max. count of retries to fetch next message before exit consumer loop")
	flset.StringVar(&c.ConsumerGroupID, prefix+"consumer-group-id", "", "special identifier of Kafka consumer group")
	flset.StringVar(&c.Topic, prefix+"topic", "", "name of Kafka default topic to consume from")

	flset.BoolVar(&c.EmitReaderInfoLog, prefix+"emit-reader-info-log", false, "indicates whether to emit Kafka reader informational log"+
		"\n\tnote, this flag not impacts error log")

	return nil
}

func (c *Conf) Validate(v validation.Validator) error { return v.Struct(c) }
