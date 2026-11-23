package flenv

import (
	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/confacet"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type (
	UsePreparer interface{ PrepareUse() error }

	Conf interface {
		Validate(v validation.Validator) error
	}

	FlagSetBinderPrefix interface {
		BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error
	}

	FlagSetBinderPrefixDefault interface {
		BindFlagSetPrefixDefault(prefix string, flset *pflag.FlagSet, def confacet.Set) error
	}

	FlagSetBinder interface {
		BindFlagSet(flset *pflag.FlagSet) error
	}

	FlagSetBinderDefault interface {
		BindFlagSetDefault(flset *pflag.FlagSet, def confacet.Set) error
	}

	FlagSetGenerator interface {
		GenerateFlagSet() (*pflag.FlagSet, error)
	}
)
