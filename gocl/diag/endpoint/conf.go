package endpoint

import (
	"fmt"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/confacet"
	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/diag/prompusher"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type (
	Conf struct {
		LivenessPathPrefix  string `validate:"gt=0"`
		MetricsPathPrefix   string `validate:"gt=0"`
		PProfPathPrefix     string `validate:"gt=0"`
		ReadinessPathPrefix string `validate:"gt=0"`

		ListenHost string `validate:"gt=0"`

		MetricsPusher prompusher.Conf

		ListenPort int `validate:"gt=0,lt=65536"`

		LivenessDisabled  bool
		MetricsDisabled   bool
		MetricsPush       bool
		PProfDisabled     bool
		ReadinessDisabled bool
		Disabled          bool
	}

	ConfGetter = interface{ DiagnosticsEndpointConf() (*Conf, error) }
)

var (
	_ ConfGetter                       = &Conf{}
	_ flenv.Conf                       = &Conf{}
	_ flenv.FlagSetBinderPrefix        = &Conf{}
	_ flenv.FlagSetBinderPrefixDefault = &Conf{}
	_ flenv.UsePreparer                = &Conf{}
)

func (c *Conf) PrepareUse() error { return c.MetricsPusher.PrepareUse() }

func (c *Conf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	return c.BindFlagSetPrefixDefault(prefix, flset, confacet.Empty())
}

func (c *Conf) BindFlagSetPrefixDefault(prefix string, flset *pflag.FlagSet, def confacet.Set) error {
	defConf := DefaultConf()
	defConf.Disabled = def.Facets().Val(confacet.KeyDiagnosticsEndpointDisabled, defConf.Disabled)

	prefix = flenv.NormalizeSelfPrefix(prefix, ConfFlenvSelfFlagPrefix)

	const (
		metricsDisableFlagName  = "metrics-disabled"
		metricsPusherFlagPrefix = "metrics-" + prompusher.ConfFlenvRootFlagPrefix
	)

	if err := c.MetricsPusher.BindFlagSetPrefix(prefix+metricsPusherFlagPrefix, flset); err != nil {
		return err
	}

	flset.StringVar(
		&c.ListenHost,
		prefix+"listen-host",
		defConf.ListenHost,
		"host part of addr on which diag endpoint will listen",
	)

	flset.IntVar(
		&c.ListenPort,
		prefix+"listen-port",
		defConf.ListenPort,
		"port part of addr on which diag endpoint will listen",
	)

	flset.BoolVar(
		&c.MetricsDisabled,
		prefix+metricsDisableFlagName,
		defConf.MetricsDisabled,
		"disable metrics endpoint",
	)

	flset.BoolVar(
		&c.MetricsPush,
		prefix+"metrics-push",
		defConf.MetricsPush,
		"specifies to use Prometheus pusher to push metrics to Prometheus gateway instead of listening for scraper request(s)"+
			"\n\tif metrics endpoint disabled then listening and push disabled also"+
			fmt.Sprintf("\n\t(see also '%s', '%s-*' flags)", prefix+metricsDisableFlagName, prefix+metricsPusherFlagPrefix),
	)

	flset.StringVar(
		&c.MetricsPathPrefix,
		prefix+"metrics-path-prefix",
		defConf.MetricsPathPrefix,
		"root path whereby metrics' requests will be handled",
	)

	flset.StringVar(
		&c.ReadinessPathPrefix,
		prefix+"readiness-path-prefix",
		defConf.ReadinessPathPrefix,
		"root path whereby app readiness probe requests will be handled"+
			"\n\tin general, readiness probe indicates the readiness of app to handle network traffic",
	)

	flset.BoolVar(
		&c.ReadinessDisabled,
		prefix+"readiness-disabled",
		defConf.ReadinessDisabled,
		"disable readiness endpoint",
	)

	flset.StringVar(
		&c.LivenessPathPrefix,
		prefix+"liveness-path-prefix",
		defConf.LivenessPathPrefix,
		"root path whereby app liveness probe requests will be handled"+
			"\n\tin general, liveness probe indicates whether app in broken state",
	)

	flset.BoolVar(
		&c.LivenessDisabled,
		prefix+"liveness-disabled",
		defConf.LivenessDisabled,
		"disable liveness endpoint",
	)

	flset.StringVar(
		&c.PProfPathPrefix,
		prefix+"pprof-path-prefix",
		defConf.PProfPathPrefix,
		"root path whereby pprof' requests will be handled",
	)

	flset.BoolVar(
		&c.PProfDisabled,
		prefix+"pprof-disabled",
		defConf.PProfDisabled,
		"disable pprof endpoint",
	)

	flset.BoolVar(
		&c.Disabled,
		prefix+"disabled",
		defConf.Disabled,
		"disable diagnostics endpoint at all",
	)

	return nil
}

func (c *Conf) Validate(validator validation.Validator) error { return validator.Struct(c) }

func (c *Conf) DiagnosticsEndpointConf() (*Conf, error) { return c, nil }

// [FlenvRootConf] returns diagnostics endpoint configuration as it would be the root configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as a fallback source of configuration, if no flag value specified).
func FlenvRootConf() (Conf, error) {
	return FlenvMemberConf("")
}

// [FlenvMemberConf] returns diagnostics endpoint configuration as it would be the member of parent configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvMemberConf(memberPrefix string) (c Conf, err error) {
	err = flenv.PopulatePrefix(memberPrefix, &c)

	return
}
