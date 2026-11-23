package kafka

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type ClientConf struct {
	ClientID       string        `validate:"min=0"`
	TLSCACertPath  string        `validate:"required_with=UseTLS"`
	TLSCertPath    string        `validate:"required_with=UseTLS"`
	TLSCertKeyPath string        `validate:"required_with=UseTLS"`
	Brokers        []string      `validate:"min=1"`
	DialTimeout    time.Duration `validate:"min=0ns"`
	IdleTimeout    time.Duration `validate:"min=1ns"`
	RequestTimeout time.Duration `validate:"min=0ns"`
	TLSInsecure    bool
	UseTLS         bool
}

func (c *ClientConf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	prefix = flenv.NormalizePrefix(prefix)

	flset.DurationVar(&c.DialTimeout, prefix+"dial-timeout", time.Second*15, "maximum amount of time a dial will wait for a connect to complete")
	flset.DurationVar(&c.IdleTimeout, prefix+"idle-timeout", time.Second*30, "maximum amount of time that connections will remain open and unused")
	flset.DurationVar(&c.RequestTimeout, prefix+"request-timeout", time.Second*0, "time limit for requests sent by Kafka client")
	flset.StringSliceVar(&c.Brokers, prefix+"brokers", nil, "comma separated list of Kafka brokers")
	flset.StringVar(&c.ClientID, prefix+"client-id", "", "unique identifier for client connections to Kafka")

	flset.BoolVar(&c.UseTLS, prefix+"use-tls", true, "indicates whether TLS enabled")
	flset.BoolVar(&c.TLSInsecure, prefix+"tls-insecure", true, "indicates whether TLS insecured (without server certificate verification)")
	flset.StringVar(&c.TLSCACertPath, prefix+"tls-ca-cert-path", "", "path to TLS CA cerificate (required when TLS enabled)")
	flset.StringVar(&c.TLSCertKeyPath, prefix+"tls-cert-key-path", "", "path to TLS client cerificate key (required when TLS enabled)")
	flset.StringVar(&c.TLSCertPath, prefix+"tls-cert-path", "", "path to TLS client cerificate (required when TLS enabled)")

	return nil
}

func (c *ClientConf) Validate(v validation.Validator) error {
	return v.Struct(c)
}

// [FlenvRootClientConf] returns Kafka client configuration as it would be the root configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvRootClientConf() (ClientConf, error) {
	return FlenvMemberClientConf(ClientConfFlenvRootFlagPrefix)
}

// [FlenvMemberClientConf] returns Kafka client configuration as it would be the member of parent configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvMemberClientConf(memberPrefix string) (conf ClientConf, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get configuration of Kafka client (type '%s'); %w", stringsext.TypeStringFor[ClientConf](), err)
		}
	}()

	err = flenv.PopulatePrefix(memberPrefix, &conf)

	return
}
