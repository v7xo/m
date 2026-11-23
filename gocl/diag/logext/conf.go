package logext

import (
	"fmt"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/confacet"
	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type Conf struct {
	LibString        string
	FormatString     string
	LevelString      string
	AttributeExclude string
	AttributeInclude string
}

var (
	_ ConfGetter                       = &Conf{}
	_ flenv.Conf                       = &Conf{}
	_ flenv.FlagSetBinderPrefix        = &Conf{}
	_ flenv.FlagSetBinderPrefixDefault = &Conf{}
	_ flenv.UsePreparer                = &Conf{}
)

func (c *Conf) PrepareUse() error { return nil }

func (c *Conf) BindFlagSetPrefixDefault(prefix string, flset *pflag.FlagSet, def confacet.Set) error {
	prefix = flenv.NormalizeSelfPrefix(prefix, ConfFlenvSelfFlagPrefix)

	const inclExclConsideration = "\n\tto designate a regular expression prepend '" + AttrInclExclPatternREDesignator + "'" +
		"\n\tnote, either inclusion or exclusion considered:" +
		"\n\t\t* if inclusion defined, then only attributes match inclusion emitted" +
		"\n\t\t* if inclusion NOT defined and exclusion defined, then only attributes NOT match exclusion emitted" +
		"\n\t\t* otherwise all attributes emitted"

	flset.StringVar(
		&c.LibString,
		prefix+"lib",
		string(def.Facets().Val(confacet.KeyLogLibProvider, DefaultLibProvider)),
		"provider lib",
	)

	flset.StringVar(
		&c.FormatString,
		prefix+"format",
		string(def.Facets().Val(confacet.KeyLogFormat, FormatJSON)),
		"output format",
	)

	flset.StringVar(
		&c.LevelString,
		prefix+"level",
		string(def.Facets().Val(confacet.KeyLogLevel, level.InfoNum.Name())),
		"min. event level to log",
	)

	flset.StringVar(
		&c.AttributeExclude,
		prefix+"attr-excl",
		"",
		"comma separated list or regular expression to match log event attributes which should be excluded from output"+inclExclConsideration,
	)

	flset.StringVar(
		&c.AttributeInclude,
		prefix+"attr-incl",
		def.Facets().Val(confacet.KeyLogAttrInclusion, ""),
		"comma separated list or regular expression to match log event attributes which should be included to output"+inclExclConsideration,
	)

	return nil
}

func (c *Conf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	return c.BindFlagSetPrefixDefault(prefix, flset, confacet.Empty())
}

func (c *Conf) Validate(validator validation.Validator) (err error) {
	if c == nil {
		return
	}

	switch {
	case !c.Level().IsKnown():
		err = fmt.Errorf("unknown logging level '%s'", c.Level())
	case !c.Format().IsKnown():
		err = fmt.Errorf("unknown logging format '%s'", c.Format())
	case !c.Lib().IsKnown():
		err = fmt.Errorf("unknown logging lib provider '%s'", c.Lib())
	}

	return
}

func (c *Conf) LogConf() (*Conf, error) { return c, nil }

func (c Conf) Lib() LibProvider  { return LibProvider(c.LibString) }
func (c Conf) Format() Format    { return Format(c.FormatString) }
func (c Conf) Level() level.Name { return level.Name(c.LevelString) }

// [FlenvRootConf] returns logging configuration as it would be the root configuration object.
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvRootConf() (Conf, error) {
	return FlenvMemberConf("")
}

// [FlenvMemberConf] returns logging configuration as it would be the member of parent configuration object.
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvMemberConf(memberPrefix string) (conf Conf, err error) {
	err = flenv.PopulatePrefix(memberPrefix, &conf)

	return
}
