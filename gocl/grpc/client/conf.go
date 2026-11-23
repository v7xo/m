package client

import (
	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/grpc"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type Conf struct {
	TransportCredentialsPresetString string `validate:"gt=0"`
	Host                             string `validate:"gt=0"`
	ClientSSLCertPEM                 string
	ClientSSLKeyPEM                  string
	MaxRecvMsgSize                   int64 `validate:"gt=0"`

	Port int `validate:"gt=0,lt=65536"`
}

func (c *Conf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	prefix = flenv.NormalizeSelfPrefix(prefix, grpc.ClientConfFlenvSelfFlagPrefix)

	flset.StringVar(
		&c.Host,
		prefix+"host",
		"",
		"host part of addr to connect to",
	)
	flset.StringVar(
		&c.TransportCredentialsPresetString,
		prefix+"transport-credentials-preset",
		DefaultTransportCredentialsPreset.String(),
		"specifies one of presets to be for gRPC client connection transport credentials"+
			"\n\t"+grpc.TransportCredentialsPresetPlaintext.String()+" - plain HTTP/2 without any credentials"+
			"\n\t"+grpc.TransportCredentialsPresetTLSInsecure.String()+" - TLS without server certificate verification and without client certificate use"+
			"\n\t"+grpc.TransportCredentialsPresetTLSSecureClient.String()+" - TLS without server certificate verification and with client certificate use"+
			"\n\t"+grpc.TransportCredentialsPresetTLSSecureServer.String()+" - TLS with server certificate verification and without client certificate use",
	)

	flset.IntVar(
		&c.Port,
		prefix+"port",
		grpc.DefaultListenPort,
		"port part of addr to connect to",
	)

	flset.StringVar(
		&c.ClientSSLCertPEM,
		prefix+"ssl-client-cert",
		"",
		"SSL client certificate (PEM format)",
	)

	flset.StringVar(
		&c.ClientSSLKeyPEM,
		prefix+"ssl-client-key",
		"",
		"SSL client certificate key (PEM format)",
	)

	flset.Int64Var(
		&c.MaxRecvMsgSize,
		prefix+"max-recv-msg-size",
		grpc.DefaultMaxRecvMsgSize,
		"message size in bytes the client can receive",
	)

	return nil
}

func (c *Conf) TransportCredentialsPreset() grpc.TransportCredentialsPreset {
	return grpc.TransportCredentialsPreset(c.TransportCredentialsPresetString)
}

func (c *Conf) Validate(v validation.Validator) (err error) {
	return v.Struct(c)
}

// [FlenvRootConf] returns gRPC-client configuration as it would be the root configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvRootConf() (Conf, error) {
	return FlenvMemberConf("")
}

// [FlenvMemberConf] returns gRPC-client configuration as it would be the member of parent configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvMemberConf(memberPrefix string) (conf Conf, err error) {
	err = flenv.PopulatePrefix(memberPrefix, &conf)

	return
}
