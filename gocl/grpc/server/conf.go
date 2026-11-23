package server

import (
	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/grpc"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type Conf struct {
	ListenHost string `validate:"gt=0"`

	ListenPort int `validate:"gt=0,lt=65536"`

	MaxRecvMsgSize            int64 `validate:"gt=0"`
	NoAccessLog               bool
	NoBodyInAccessLog         bool
	NoMetrics                 bool
	PreventErrDetailedMessage bool
}

func (c *Conf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	prefix = flenv.NormalizeSelfPrefix(prefix, ConfFlenvSelfFlagPrefix)

	flset.StringVar(
		&c.ListenHost,
		prefix+"listen-host",
		grpc.DefaultListenHost,
		"host part of addr on which server will listen",
	)

	flset.IntVar(
		&c.ListenPort,
		prefix+"listen-port",
		grpc.DefaultListenPort,
		"port part of addr on which server will listen",
	)

	flset.Int64Var(
		&c.MaxRecvMsgSize,
		prefix+"max-recv-msg-size",
		grpc.DefaultMaxRecvMsgSize,
		"message size in bytes the server can receive",
	)

	flset.BoolVar(
		&c.NoAccessLog,
		prefix+"no-access-log",
		grpc.DefaultNoAccessLog,
		"disable access log for calls",
	)

	flset.BoolVar(
		&c.NoBodyInAccessLog,
		prefix+"no-body-in-access-log",
		grpc.DefaultNoBodyInAccessLog,
		"indicates to omit request/response body in output of access log"+
			"\n\t"+"senseless if access log disabled",
	)

	flset.BoolVar(
		&c.NoMetrics,
		prefix+"no-metrics",
		grpc.DefaultNoMetrics,
		"disable metrics emitting",
	)

	flset.BoolVar(
		&c.PreventErrDetailedMessage,
		prefix+"prevent-err-detailed-message",
		grpc.DefaultPreventErrDetailedMessage,
		"prevents gRPC status message to be filled with backend error message"+
			"\n\tfalse - gRPC status message will be filled with backend error message that leads to potential secure info disclosure"+
			"\n\ttrue - gRPC status message will be filled with message translated from original backend error",
	)

	return nil
}

func (c *Conf) Validate(v validation.Validator) error {
	return v.Struct(c)
}

// [FlenvRootConf] returns gRPC-server configuration as it would be the root configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvRootConf() (Conf, error) {
	return FlenvMemberConf("")
}

// [FlenvMemberConf] returns gRPC-server configuration as it would be the member of parent configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvMemberConf(memberPrefix string) (conf Conf, err error) {
	err = flenv.PopulatePrefix(memberPrefix, &conf)

	return
}
