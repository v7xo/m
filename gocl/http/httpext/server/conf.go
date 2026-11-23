package server

import (
	"time"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/mw/cors"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type Conf struct {
	ListenHost                string `validate:"gt=0"`
	CORS                      cors.Config
	ShutdownTimeoutDur        time.Duration `validate:"min=-1ns"`
	ListenPort                int           `validate:"gt=0,lt=65536"`
	CORSEnabled               bool
	PreventErrDetailedMessage bool
}

var (
	_ flenv.Conf                = &Conf{}
	_ flenv.FlagSetBinderPrefix = &Conf{}
	_ flenv.UsePreparer         = &Conf{}
)

func (s *Conf) PrepareUse() error { return nil }

func (c *Conf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	prefix = flenv.NormalizeSelfPrefix(prefix, ConfFlenvSelfFlagPrefix)

	if err := c.CORS.BindFlagSetPrefix(prefix+cors.ConfigFlagPrefix, flset); err != nil {
		return err
	}

	flset.BoolVar(
		&c.PreventErrDetailedMessage,
		prefix+"prevent-err-detailed-message",
		true,
		"indicates to prevent responses with full backend error message;"+
			"\n\tfalse - no prevention, that leads to potential secure info disclosure"+
			"\n\ttrue - prevention enabled and server will replace error message with special placeholder",
	)

	flset.BoolVar(
		&c.CORSEnabled,
		prefix+cors.ConfigFlagPrefix+"-enabled",
		DefaultCORSEnabled,
		"indicates whether CORS enabled for HTTP server;"+
			"\n\tif enabled, then CORS config should be supplied",
	)

	flset.DurationVar(
		&c.ShutdownTimeoutDur,
		prefix+"shutdown-timeout",
		httpext.DefaultServerShutdownTimeout.Dur(),
		"shutdown wait timeout for HTTP server; time of wait for HTTP server graceful shutdown",
	)

	flset.StringVar(
		&c.ListenHost,
		prefix+"listen-host",
		httpext.DefaultListenHost,
		"host part of addr on which server will listen",
	)

	flset.IntVar(
		&c.ListenPort,
		prefix+"listen-port",
		httpext.DefaultListenPort,
		"port part of addr on which server will listen",
	)

	return nil
}

func (c *Conf) Validate(validator validation.Validator) error {
	return validator.Struct(c)
}

// [FlenvRootConfig] returns HTTP server configuration as it would be the root configuration object.
//
// Default configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvRootConfig() (Conf, error) {
	return FlenvMemberConfig("")
}

// [FlenvMemberConfig] returns HTTP server configuration as it would be the member of parent configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvMemberConfig(memberPrefix string) (conf Conf, err error) {
	err = flenv.PopulatePrefix(memberPrefix, &conf)

	return
}
