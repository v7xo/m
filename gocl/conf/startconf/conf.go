package startconf

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/confacet"
	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

var (
	_ flenv.Conf                 = &Conf{}
	_ flenv.FlagSetBinder        = &Conf{}
	_ flenv.FlagSetBinderDefault = &Conf{}
	_ flenv.UsePreparer          = &Conf{}
)

type Conf struct {
	confVariant         string
	precedingFileNames  string
	succeedingFileNames string
	userProfileConfDir  string

	cmdTimeout   time.Duration `validate:"min=-1ns"`
	cmdExitDelay time.Duration `validate:"min=0ns"`
}

func (c *Conf) PrepareUse() error { return nil }

// nolint:mnd,revive
func (c *Conf) BindFlagSetDefault(flset *pflag.FlagSet, def confacet.Set) error {
	const (
		variantFlag            = "conf-variant"
		precedingFlag          = "conf-preceding"
		succeedingFlag         = "conf-succeeding"
		userProfileConfDirFlag = "user-profile-conf-dir"
	)

	flsetPrefix := flenv.NormalizePrefix(ConfFlenvSelfFlagPrefix)

	flset.StringVar(
		&c.userProfileConfDir,
		flsetPrefix+userProfileConfDirFlag,
		"",
		"overrides default user profile configuration directory"+
			"\n\tthis is directory wherein configutation file(s) being searched"+
			"\n\t(see also --"+flenv.FlagPrintUserProfileConfDir+")",
	)

	flset.StringVar(
		&c.confVariant,
		flsetPrefix+variantFlag,
		def.Facets().Val(confacet.KeyConfVariant, ConfVariantAuto).String(),
		"denotes configuration variant"+
			"\n\tcurrently supported variants are:"+

			"\n\t\t'"+string(ConfVariantUnspecified)+"' (unspecified) - files '--"+flsetPrefix+precedingFlag+
			"' and '--"+flsetPrefix+succeedingFlag+"' considered as contained configuration"+

			"\n\t\t"+string(ConfVariantIgnore)+" - no any configuration files considered"+

			"\n\t\t"+string(ConfVariantDevelop)+" - files --"+flsetPrefix+precedingFlag+", '"+DevelopConfYamlFileName+"', --"+flsetPrefix+succeedingFlag+" considered as contained configuration"+
			"\n\t\t"+string(ConfVariantDevelopLocal)+" - files --"+flsetPrefix+precedingFlag+", '"+DevelopLocalConfYamlFileName+"', '"+DevelopConfYamlFileName+"', --"+flsetPrefix+succeedingFlag+" considered as contained configuration"+

			"\n\t\t"+string(ConfVariantTest)+" - files --"+flsetPrefix+precedingFlag+", '"+TestConfYamlFileName+"', --"+flsetPrefix+succeedingFlag+" considered as contained configuration"+
			"\n\t\t"+string(ConfVariantTestLocal)+" - files --"+flsetPrefix+precedingFlag+", '"+TestLocalConfYamlFileName+"', '"+TestConfYamlFileName+"', --"+flsetPrefix+succeedingFlag+" considered as contained configuration"+

			"\n\t\t"+string(ConfVariantUserProfile)+" - files in user profile directory considered as contained configuration (see also --"+flenv.FlagPrintUserProfileConfDir+")"+

			"\n\t\t"+string(ConfVariantAuto)+" - actual configuration variant determined as follows:"+
			fmt.Sprintf("\n\t\t\t%s - if working directory contains '%s' file", ConfVariantDevelopLocal, DevelopLocalConfYamlFileName)+
			fmt.Sprintf("\n\t\t\t'%s' (unspecified) - otherwise", ConfVariantUnspecified)+

			"\n\tcan be omitted",
	)

	flset.StringVar(
		&c.precedingFileNames,
		flsetPrefix+precedingFlag,
		"",
		"specifies comma separated list of file names which contain configuration"+
			"\n\tprecedes configuration determined for '"+flsetPrefix+variantFlag+"'"+
			"\n\tfile names will be handled in order this list defines"+
			"\n\t(see also '"+flsetPrefix+variantFlag+"', '"+flsetPrefix+succeedingFlag+"')",
	)

	flset.StringVar(
		&c.succeedingFileNames,
		flsetPrefix+succeedingFlag,
		"",
		"specifies comma separated list of file names which contain configuration"+
			"\n\tsucceeds configuration determined for '"+flsetPrefix+variantFlag+"'"+
			"\n\tfile names will be handled in order this list defines"+
			"\n\t(see also '"+flsetPrefix+variantFlag+"', '"+flsetPrefix+precedingFlag+"')",
	)

	flset.DurationVar(
		&c.cmdTimeout,
		flsetPrefix+"cmd-timeout",
		timeout.Infinite.Dur(),
		"time to wait command to be completed"+
			"\n\tafter specified time the command will be signaled to terminate/cancel its execution"+
			"\n\t\t"+timeout.Infinite.Dur().String()+" - infinite wait",
	)

	flset.DurationVar(
		&c.cmdExitDelay,
		flsetPrefix+"cmd-exit-delay",
		time.Duration(0),
		"time delay before exit"+
			"\n\tafter command completes its execution the delay will taken before exit"+
			"\n\t\t0 - no delay",
	)

	return nil
}

func (c *Conf) BindFlagSet(flset *pflag.FlagSet) error {
	return c.BindFlagSetDefault(flset, confacet.Empty())
}

func (c *Conf) Validate(v validation.Validator) error { return v.Struct(c) }

func (c *Conf) ConfVariant() ConfVariant { return ConfVariant(c.confVariant) }

func (c *Conf) UserProfileConfDir() string { return c.userProfileConfDir }

func (c *Conf) PrecedingFileNames() []string {
	res := strings.Split(c.precedingFileNames, ",")
	if len(res) == 0 || (len(res) == 1 && res[0] == "") {
		return nil
	}

	return res
}

func (c *Conf) SucceedingFileNames() []string {
	res := strings.Split(c.succeedingFileNames, ",")
	if len(res) == 0 || (len(res) == 1 && res[0] == "") {
		return nil
	}

	return res
}

func (c *Conf) CmdExitDelay() time.Duration { return c.cmdExitDelay }

func (c *Conf) CmdTimeout() timeout.Timeout { return timeout.Timeout(c.cmdTimeout) }

// [FlenvRootConf] returns configuration as it would be the root configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvRootConf() (c Conf, err error) {
	err = flenv.Populate(&c)

	return
}
