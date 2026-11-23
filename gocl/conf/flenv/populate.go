package flenv

import (
	"errors"
	"fmt"
	"os"
	"strings"

	pf "github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

// Populate populates configuration config fields with values from environment variables.
//
// Configuration conf implements method [FlagSetBinder.BindFlagSet] which binds every configuration field to dedicated flag name.
// That flag name can be translated to a name of environment variable (see [FlagNameToEnvironmentName]). Thus, this function populates configuration with values from these environment variables.
func Populate[
	T any,
	TConf interface {
		Conf
		FlagSetBinder
		*T
	}](conf TConf) (err error) {
	defer func() {
		if err = errs.Join(err, errs.PanicErr(recover())); err != nil {
			err = fmt.Errorf("failed to populate configuration '%s' from command flags and environment; %w", stringsext.TypeStringForExpr(conf), err)
		}
	}()

	return populate(conf, func(c TConf, f *pf.FlagSet) error { return c.BindFlagSet(f) })
}

// [PopulatePrefix] populates configuration config fields with values from environment variables.
//
// Configuration conf implements method [FlagSetBinderPrefix.BindFlagSetPrefix] which binds every configuration field to dedicated flag name (with specified prefix).
// That flag name can be translated to a name of environment variable (see [FlagNameToEnvironmentName]). Thus, this function populates configuration with values from these environment variables.
func PopulatePrefix[
	T any,
	TConf interface {
		Conf
		FlagSetBinderPrefix
		*T
	}](prefix string, conf TConf) (err error) {
	defer func() {
		if err = errs.Join(err, errs.PanicErr(recover())); err != nil {
			err = fmt.Errorf("failed to populate configuration '%s' from command flags and environment (prefix '%s'); %w", stringsext.TypeStringForExpr(conf), prefix, err)
		}
	}()

	prefix = NormalizePrefix(prefix)

	return populate(conf, func(c TConf, f *pf.FlagSet) error { return c.BindFlagSetPrefix(prefix, f) })
}

func populate[
	T any,
	TConf interface {
		Conf
		*T
	}](conf TConf, bindFn func(TConf, *pf.FlagSet) error) error {
	flset := newFlagSet()

	if err := bindFn(conf, flset); err != nil {
		return err
	}

	err := flset.Parse(os.Args[1:])
	switch {
	case errors.Is(err, pf.ErrHelp):
		return nil
	case err != nil:
		return err
	}

	if err = PopulateFlagSetFromEnvironment(flset); err != nil {
		return err
	}

	if err = conf.Validate(validation.DefaultValidator()); err != nil {
		return err
	}

	if t, ok := any(conf).(UsePreparer); ok {
		if err = t.PrepareUse(); err != nil {
			return err
		}
	}

	return nil
}

// PopulateFlagSetFromEnvironment populates unassigned (not set yet) flags with values from environment.
//
//nolint:gocognit,maintidx
func PopulateFlagSetFromEnvironment(flset *pf.FlagSet) error {
	if flset == nil {
		return nil
	}

	assignedValueFlagNames := make(map[string]struct{})

	flset.Visit(func(f *pf.Flag) { assignedValueFlagNames[f.Name] = struct{}{} })

	var flagErr error

	flset.VisitAll(
		func(f *pf.Flag) {
			if flagErr != nil {
				return
			}

			if _, ok := assignedValueFlagNames[f.Name]; ok || IsFlagEnvironmentNonBindable(f) {
				return
			}

			envVarName := FlagNameToEnvironmentName(f.Name)
			if envVarVal, ok := os.LookupEnv(envVarName); ok {
				switch f.Value.Type() {
				case "stringArray", "stringSlice":
					vals := strings.Split(envVarVal, ",")
					for _, v := range vals {
						if err := flset.Set(f.Name, v); err != nil {
							flagErr = fmt.Errorf("unable to bind flag '%s' to environment variable '%s'; variable value: %v; %w", f.Name, envVarName, envVarVal, err) //nolint

							return
						}
					}
				default:
					if err := flset.Set(f.Name, envVarVal); err != nil {
						flagErr = fmt.Errorf("unable to bind flag '%s' to environment variable '%s'; variable value: %v; %w", f.Name, envVarName, envVarVal, err) //nolint

						return
					}
				}
			}
		},
	)

	return flagErr
}
