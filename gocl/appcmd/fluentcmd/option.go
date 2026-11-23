package fluentcmd

import (
	"errors"
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/conf/confacet"
	"rn.m.vo7.io/vo7io/m/gocl/conf/confset"
	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/conf/startconf"
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

type (
	OptionFn[TConf flenv.Conf] func(*impl[TConf]) error
	OptionConfSetFn            = OptionFn[confset.Set]
)

var _ CmdGenerator = OptionFn[noConf](nil)

// CLI option sets presets for this command appropriated for command line interface utility:
//
//   - default configuration variant = [startconf.ConfVariantUserProfile]
//   - default log attribute inclusion = [attrname.AccessLog], [attrname.ConfVariant], [attrname.DateTime], [attrname.Dt], [attrname.Err], [attrname.Error], [attrname.LogOrigin], [attrname.LogScope], [attrname.Message], [attrname.RPCMethod]
//   - default log format = [logext.FormatLogfmt]
//   - default log level = [level.Warn]
//   - default diagnostics endpoint disabled = [true]
func (o OptionFn[TConf]) CLI() OptionFn[TConf] {
	const comma = ","

	return fopt.Pack(
		o,
		Option[TConf]().
			cli().
			defConfVariant(startconf.ConfVariantUserProfile).
			defLogAttributeInclusion(
				string(attrname.AccessLog)+
					comma+string(attrname.ConfVariant)+
					comma+string(attrname.DateTime)+
					comma+string(attrname.Dt)+
					comma+string(attrname.Err)+
					comma+string(attrname.Error)+
					comma+string(attrname.LogOrigin)+
					comma+string(attrname.LogScope)+
					comma+string(attrname.Message)+
					comma+string(attrname.RPCMethod)).
			defLogFormat(logext.FormatLogfmt).
			defLogLevel(level.Warn).
			defDiagnosticsEndpointDisabled(true),
	)
}

func (o OptionFn[TConf]) GenerateCmd() (Cmd, error) {
	res, err := New(o)
	if err != nil {
		return nil, fmt.Errorf("failed to generate command; %w", err)
	}

	return res, nil
}

func (o OptionFn[TConf]) cli() OptionFn[TConf] {
	return fopt.Chain(o, func(t *impl[TConf]) { t.cli = true })
}

func (o OptionFn[TConf]) Use(val string) OptionFn[TConf] {
	return fopt.Chain(o, func(t *impl[TConf]) { t.use = val })
}

func (o OptionFn[TConf]) Short(val string) OptionFn[TConf] {
	return fopt.Chain(o, func(t *impl[TConf]) { t.short = val })
}

func (o OptionFn[TConf]) SilenceUsage(val bool) OptionFn[TConf] {
	return fopt.Chain(o, func(t *impl[TConf]) { t.silenceUsage = val })
}

func (o OptionFn[TConf]) SilenceError(val bool) OptionFn[TConf] {
	return fopt.Chain(o, func(t *impl[TConf]) { t.silenceError = val })
}

func (o OptionFn[TConf]) Child(val Cmd) OptionFn[TConf] {
	return fopt.ChainErr(o, func(t *impl[TConf]) error { return appendChild(t, val) })
}

func (o OptionFn[TConf]) ExecFn(val UserExecFn[TConf]) OptionFn[TConf] {
	return fopt.Chain(o, func(t *impl[TConf]) { t.userExecFn = val })
}

func (o OptionFn[TConf]) BeforeExecFn(val UserExecFn[TConf]) OptionFn[TConf] {
	return fopt.Chain(o, func(t *impl[TConf]) { t.userBeforeExecFn = val })
}

func (o OptionFn[TConf]) Conf(val TConf) OptionFn[TConf] {
	return fopt.Chain(o, func(t *impl[TConf]) { t.conf, t.confSupplied = val, true })
}

func (o OptionFn[TConf]) ConfFlenvFlagPrefix(v string) OptionFn[TConf] {
	return fopt.Chain(o, func(t *impl[TConf]) { t.confFlenvFlagPrefix = v })
}

func (o OptionFn[TConf]) defLogFormat(val logext.Format) OptionFn[TConf] {
	return fopt.ChainErr(
		o,
		func(t *impl[TConf]) error {
			if !val.IsJSON() && !val.IsLogfmt() {
				return fmt.Errorf("unknown/invalid log format '%s'", val)
			}

			t.confDef = t.confDef.Tune().Set(confacet.KeyLogFormat, val).Tuned()

			return nil
		})
}

func (o OptionFn[TConf]) defLogLevel(val level.Name) OptionFn[TConf] {
	return fopt.ChainErr(
		o,
		func(t *impl[TConf]) error {
			if !val.IsKnown() {
				return fmt.Errorf("invalid/unknown log level '%s'", val)
			}

			t.confDef = t.confDef.Tune().Set(confacet.KeyLogLevel, val).Tuned()

			return nil
		})
}

func (o OptionFn[TConf]) defLogAttributeInclusion(val string) OptionFn[TConf] {
	return fopt.Chain(
		o,
		func(t *impl[TConf]) { t.confDef = t.confDef.Tune().Set(confacet.KeyLogAttrInclusion, val).Tuned() })
}

func (o OptionFn[TConf]) defConfVariant(val startconf.ConfVariant) OptionFn[TConf] {
	return fopt.ChainErr(
		o,
		func(t *impl[TConf]) error {
			switch val {
			case
				startconf.ConfVariantAuto,
				startconf.ConfVariantDevelop,
				startconf.ConfVariantDevelopLocal,
				startconf.ConfVariantIgnore,
				startconf.ConfVariantTest,
				startconf.ConfVariantTestLocal,
				startconf.ConfVariantUserProfile:
			case startconf.ConfVariantUnspecified:
				fallthrough
			default:
				return fmt.Errorf("invalid/unknown value '%s' of configuration variant can't be set as default", val)
			}

			t.confDef = t.confDef.Tune().Set(confacet.KeyConfVariant, val).Tuned()

			return nil
		})
}

func (o OptionFn[TConf]) defDiagnosticsEndpointDisabled(val bool) OptionFn[TConf] {
	return fopt.Chain(
		o,
		func(t *impl[TConf]) {
			t.confDef = t.confDef.Tune().Set(confacet.KeyDiagnosticsEndpointDisabled, val).Tuned()
		})
}

func (o OptionFn[TConf]) IgnoreMissOfConf() OptionFn[TConf] {
	return fopt.Chain(o, func(t *impl[TConf]) { t.ignoreMissOfConf = true })
}

func (o OptionFn[TConf]) ChildFn(val func() (Cmd, error)) OptionFn[TConf] {
	return func(t *impl[TConf]) error {
		if val == nil {
			return errors.New("child command generator func can't be a nil")
		}

		if err := o(t); err != nil {
			return err
		}

		childCmd, err := val()
		if err != nil {
			return fmt.Errorf("child command generator func has returned error as a result; %w", err)
		}

		if childCmd == nil {
			return errors.New("child command generator func has returned nil as a result")
		}

		return appendChild(t, childCmd)
	}
}

// Log sets logger to be used during command execution.
// Consider, specified here nil will be replaced with Nop-logger.
func (o OptionFn[TConf]) Log(val logext.Logger) OptionFn[TConf] {
	return fopt.Chain(
		o,
		func(t *impl[TConf]) {
			if val == nil {
				t.logFactoryFn = func(logext.ConfGetter) (logext.Logger, error) { return logext.NopLogger, nil }

				return
			}

			t.logFactoryFn = func(logext.ConfGetter) (logext.Logger, error) { return val, nil }
		},
	)
}

func Option[TConf flenv.Conf]() OptionFn[TConf] {
	return fopt.Nop[impl[TConf]]
}

func OptionNoConf() OptionFn[noConf] { return Option[noConf]().IgnoreMissOfConf() }

func OptionConfSet() OptionFn[confset.Set] { return Option[confset.Set]() }

func appendChild[TConf flenv.Conf](parent *impl[TConf], child Cmd) error {
	if child == nil {
		return fmt.Errorf("%w; child command can't be a nil", errs.ErrNilValue)
	}

	if err := child.SetParent(parent); err != nil {
		return err
	}

	parent.children = append(parent.children, child)

	return nil
}
