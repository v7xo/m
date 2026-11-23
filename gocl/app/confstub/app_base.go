package confstub

import (
	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type AppBase struct {
	Log
	Start
	DiagEndpoint
}

var (
	_ flenv.Conf                = &AppBase{}
	_ flenv.FlagSetBinder       = &AppBase{}
	_ flenv.FlagSetBinderPrefix = &AppBase{}
	_ flenv.UsePreparer         = &AppBase{}
)

func (c *AppBase) BindFlagSet(flset *pflag.FlagSet) error {
	return c.BindFlagSetPrefix("", flset)
}

func (c *AppBase) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	prefix = flenv.NormalizePrefix(prefix)

	if err := c.Log.BindFlagSetPrefix(prefix, flset); err != nil {
		return err
	}

	if err := c.Start.BindFlagSet(flset); err != nil {
		return err
	}

	if err := c.DiagEndpoint.BindFlagSetPrefix(prefix, flset); err != nil {
		return err
	}

	return nil
}

func (c *AppBase) Validate(v validation.Validator) error { return v.Struct(c) }

func (c *AppBase) PrepareUse() error {
	if err := c.Log.PrepareUse(); err != nil {
		return err
	}

	if err := c.Start.PrepareUse(); err != nil {
		return err
	}

	if err := c.DiagEndpoint.PrepareUse(); err != nil {
		return err
	}

	return nil
}
