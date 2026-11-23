package fluentcmd

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/confacet"
	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/conf/startconf"
	"rn.m.vo7.io/vo7io/m/gocl/diag/endpointfactory"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logfactory"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/value/valonce"
)

type impl[T flenv.Conf] struct {
	conf                              T
	flagSet                           valonce.Of[*pflag.FlagSet]
	confDef                           confacet.Set
	parent                            Cmd
	userExecFn                        UserExecFn[T]
	userBeforeExecFn                  UserExecFn[T]
	logFactoryFn                      logfactory.Fn
	diagnosticsEndpointFactoryFn      endpointfactory.Fn
	use                               string
	short                             string
	confFlenvFlagPrefix               string
	children                          []Cmd
	startConf                         startconf.Conf
	silenceUsage                      bool
	silenceError                      bool
	ignoreMissOfConf                  bool
	cli                               bool
	confSupplied                      bool
	printEnvironmentAndExit           bool
	printUserProfileConfSubdirAndExit bool
}

var _ flenv.FlagSetGenerator = &impl[noConf]{}

func (s *impl[T]) ConfFlenvFlagPrefix() string { return s.confFlenvFlagPrefix }
func (s *impl[T]) Short() string               { return s.short }
func (s *impl[T]) SilenceError() bool          { return s.silenceError }
func (s *impl[T]) SilenceUsage() bool          { return s.silenceUsage }
func (s *impl[T]) Use() string                 { return s.use }

func (s *impl[T]) Parent() Cmd {
	if s == nil {
		return nil
	}

	return s.parent
}

func (s *impl[T]) Children() []Cmd {
	if s == nil || len(s.children) == 0 {
		return nil
	}

	return slices.Clone(s.children)
}

func (s *impl[T]) SetParent(parent Cmd) error {
	if s.parent != nil {
		return fmt.Errorf("unable to change parent of command '%s', parent can be set once", s.String())
	}

	s.parent = parent

	return nil
}

func (s *impl[T]) GenerateFlagSet() (_ *pflag.FlagSet, err error) {
	return s.flagSet.Eval(context.TODO())
}

func (s *impl[T]) String() string {
	if s == nil {
		return placeholder.Nil
	}

	if len(s.short) == 0 {
		return s.use
	}

	return s.use + " (" + s.short + ")"
}

func (s *impl[T]) generateNewFlagSet(_ context.Context) (_ *pflag.FlagSet, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to generate flag set for command '%s'; %w", s, err)
		}
	}()

	defer func() { err = errs.Join(err, errs.PanicErr(recover())) }()

	flset := pflag.NewFlagSet("", pflag.PanicOnError)

	flset.BoolVar(
		&s.printEnvironmentAndExit,
		flenv.FlagPrintEnvironment,
		false,
		"writes to standard output names of environment variables to be used as source of values for explicitly unset flags and exits",
	)

	flset.BoolVar(
		&s.printUserProfileConfSubdirAndExit,
		flenv.FlagPrintUserProfileConfDir,
		false,
		"writes to standard output the path(s) of directories/files wherein user profile configuration expected to be by default"+
			"\n\tapplicable for configuration variant '"+string(startconf.ConfVariantUserProfile)+"' only",
	)

	if err = s.startConf.BindFlagSetDefault(flset, s.confDef); err != nil {
		return nil, fmt.Errorf("failed to bind flag set to '%s' configuration; %w", typestring.TypeStringOfVal(s.startConf), err)
	}

	if !s.confSupplied && s.ignoreMissOfConf {
		return flset, nil
	} else if !s.confSupplied {
		return nil, fmt.Errorf("unable to generate flag set for command '%s'; no configuration supplied", s)
	}

	if err = flenv.CallFlagBinder(s.conf, s.confFlenvFlagPrefix, flset, s.confDef); err != nil {
		return nil, err
	}

	return flset, nil
}

func (s *impl[T]) cmdLogConfGetter() logext.ConfGetter {
	if getter, ok := any(s.conf).(logext.ConfGetter); ok {
		return getter
	}

	return nil
}

func (s *impl[T]) validate() error {
	if s.userExecFn == nil {
		return errors.New("exec function not specified")
	}

	if !s.confSupplied && !s.ignoreMissOfConf {
		return errors.New("no configuration supplied")
	}

	if s.confDef == nil {
		return errors.New("no set of default configuration facets specified")
	}

	return nil
}

func New[TConf flenv.Conf](opts ...OptionFn[TConf]) (_ Cmd, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl[TConf]](), err)
		}
	}()

	impl := &impl[TConf]{
		userExecFn: func(_ context.Context, _ TConf, _ logext.Logger) error {
			return errors.New("unable to execute this command; there is no exec function specified")
		},

		confDef:                      confacet.Tuner().Tuned(),
		diagnosticsEndpointFactoryFn: endpointfactory.Default(),
		logFactoryFn:                 logfactory.Default(),
		silenceError:                 true,
		silenceUsage:                 true,
	}

	impl.flagSet = valonce.New(impl.generateNewFlagSet)

	if err = fopt.Apply(impl, opts...); err != nil {
		return nil, err
	}

	if err = impl.validate(); err != nil {
		return nil, err
	}

	return impl, nil
}

func MustNew[TConf flenv.Conf](opts ...OptionFn[TConf]) Cmd {
	res, err := New[TConf](opts...)
	if err != nil {
		panic(err)
	}

	return res
}

func NewNoConf(opts ...OptionFn[noConf]) (Cmd, error) {
	return New[noConf](fopt.Pack(OptionNoConf(), opts...))
}

func MustNewNoConf(opts ...OptionFn[noConf]) Cmd {
	res, err := NewNoConf(opts...)
	if err != nil {
		panic(err)
	}

	return res
}
