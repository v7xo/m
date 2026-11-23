package flenv

import (
	"fmt"
	"os"
	"strings"

	pf "github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/confacet"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

var flNameToEnvVarNameRepl = strings.NewReplacer("-", "_", ".", "_")

func PrintFlagEnvironment(flset *pf.FlagSet) {
	if flset == nil {
		return
	}

	flset.VisitAll(
		func(f *pf.Flag) {
			if IsFlagEnvironmentNonBindable(f) {
				return
			}

			varName := FlagNameToEnvironmentName(f.Name)
			varVal := placeholder.NotAvailable

			if varValFound, ok := os.LookupEnv(varName); ok {
				varVal = "'" + varValFound + "'"
			}

			fmt.Printf("%s = %s\n", varName, varVal) //nolint
		},
	)
}

func IsFlagEnvironmentNonBindable(f *pf.Flag) bool {
	if f == nil {
		return false
	}

	return strings.EqualFold(f.Name, FlagPrintEnvironment) || strings.EqualFold(f.Name, FlagPrintUserProfileConfDir) || strings.EqualFold(f.Name, FlagHelp)
}

func FlagEnvironmentNames(flset *pf.FlagSet) (map[string]struct{}, error) {
	if flset == nil {
		return nil, nil
	}

	names := map[string]struct{}{}
	nameErr := error(nil)

	flset.VisitAll(
		func(f *pf.Flag) {
			if nameErr != nil || IsFlagEnvironmentNonBindable(f) {
				return
			}

			name := FlagNameToEnvironmentName(f.Name)

			if _, ok := names[name]; ok {
				nameErr = fmt.Errorf("unexpected duplicate of environment name '%s' occurred (source flag '%s')", name, f.Name)

				return
			}

			names[name] = struct{}{}
		},
	)

	return names, nameErr
}

func FlagEnvironmentNamesForBinderPrefix[
	T any,
	TConf interface {
		FlagSetBinderPrefix
		*T
	}](prefix string, conf TConf) (map[string]struct{}, error) {
	flset := newFlagSet()

	if err := conf.BindFlagSetPrefix(prefix, flset); err != nil {
		return nil, err
	}

	return FlagEnvironmentNames(flset)
}

func FlagNameToEnvironmentName(n string) string {
	return flNameToEnvVarNameRepl.Replace(strings.ToUpper(n))
}

func NormalizePrefix(prefix string) string {
	if prefix == "" || strings.HasSuffix(prefix, FlagNameDelimiter) {
		return prefix
	}

	return prefix + FlagNameDelimiter
}

func NormalizeSelfPrefix(prefix string, selfPrefix string) string {
	if selfPrefix == "" {
		return NormalizePrefix(prefix)
	}

	if prefix == "" {
		return selfPrefix + FlagNameDelimiter
	}

	if prefix == selfPrefix+FlagNameDelimiter ||
		strings.HasSuffix(prefix, FlagNameDelimiter+selfPrefix+FlagNameDelimiter) {
		// for ex. prefix == "self-" || prefix == "...foo-self-"
		return prefix
	}

	if prefix == selfPrefix {
		// for ex. prefix == "self"
		return prefix + FlagNameDelimiter
	}

	if strings.HasSuffix(prefix, FlagNameDelimiter+selfPrefix) {
		// for ex. prefix == "...foo-self"
		return prefix + FlagNameDelimiter
	}

	if strings.HasSuffix(prefix, FlagNameDelimiter) {
		// for ex. prefix == "foo-"
		return prefix + selfPrefix + FlagNameDelimiter
	}

	// for ex. prefix == "foo"
	return prefix + FlagNameDelimiter + selfPrefix + FlagNameDelimiter
}

func newFlagSet() *pf.FlagSet {
	flset := pf.NewFlagSet("", pf.ContinueOnError)
	flset.Usage = func() {}
	flset.ParseErrorsAllowlist = pf.ParseErrorsAllowlist{UnknownFlags: true}

	return flset
}

func CallFlagBinder[TConf Conf](conf TConf, prefix string, flset *pf.FlagSet, def confacet.Set) (err error) {
	defer func() {
		if err = errs.Join(err, errs.PanicErr(recover())); err != nil {
			err = fmt.Errorf("failed to call the flag binder for configuration of type '%s'; %w", typestring.TypeStringOfVal(conf), err)
		}
	}()

	if prefix == "" {
		if conf, ok := any(conf).(FlagSetBinderDefault); ok {
			return conf.BindFlagSetDefault(flset, def)
		}

		if conf, ok := any(conf).(FlagSetBinderPrefixDefault); ok {
			return conf.BindFlagSetPrefixDefault("", flset, def)
		}

		if conf, ok := any(conf).(FlagSetBinder); ok {
			return conf.BindFlagSet(flset)
		}

		if conf, ok := any(conf).(FlagSetBinderPrefix); ok {
			return conf.BindFlagSetPrefix("", flset)
		}

		return fmt.Errorf("configuration not implements '%s'", typestring.TypeStringFor[FlagSetBinderDefault]())
	}

	if conf, ok := any(conf).(FlagSetBinderPrefixDefault); ok {
		return conf.BindFlagSetPrefixDefault(prefix, flset, def)
	}

	if conf, ok := any(conf).(FlagSetBinderPrefix); ok {
		return conf.BindFlagSetPrefix(prefix, flset)
	}

	return fmt.Errorf("flag prefix '%s' specified, but configuration not implements '%s'", prefix, typestring.TypeStringFor[FlagSetBinderPrefixDefault]())
}
