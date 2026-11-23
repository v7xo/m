package cors

import (
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*middleware) error

func (o OptionFn) WithConfig(val Config) OptionFn {
	return func(i *middleware) error {
		if err := o(i); err != nil {
			return err
		}

		applyCfg(i, val)

		return nil
	}
}

func applyCfg(i *middleware, val Config) {
	i.AllowCredentials = val.AllowCredentials
	i.AllowedHeaders = strings.Split(val.AllowedHeaders, ",")
	i.AllowedMethods = strings.Split(val.AllowedMethods, ",")
	i.AllowedOrigins = strings.Split(val.AllowedOrigins, ",")
	i.AllowOriginFunc = nil
	i.Debug = false
	i.ExposedHeaders = strings.Split(val.ExposedHeaders, ",")
	i.MaxAge = int(val.MaxAge.Seconds())
	i.OptionsPassthrough = val.OptionsPassthrough
}

// DefaultOption creates option with default configuration.
func DefaultOption() OptionFn {
	return func(i *middleware) error {
		cfg, err := DefaultConfig()
		if err != nil {
			return err
		}

		applyCfg(i, cfg)

		return nil
	}
}

func Option() OptionFn { return fopt.Nop[middleware] }
