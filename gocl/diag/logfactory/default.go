package logfactory

import (
	"errors"
	"fmt"

	attrs "rn.m.vo7.io/vo7io/m/gocl/appmeta/buildinfo"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext/gokit"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ptrof"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

// [Default] returns default factory function of [logext.Logger].
//
// Returned function will use [logext.FlenvRootConf] if no configuration getter supplied
// or supplied getter not returned log configuration.
func Default() Fn {
	return func(confGetter logext.ConfGetter) (logext.Logger, error) {
		var (
			conf *logext.Conf
			err  error
		)

		for {
			switch confGetter {
			case nil:
				if conf, err = ptrof.OfErr(logext.FlenvRootConf()); err != nil {
					return nil, err
				}
			default:
				if conf, err = confGetter.LogConf(); err != nil {
					return nil, err
				} else if conf == nil {
					confGetter = nil

					continue
				}
			}

			break
		}

		switch {
		case conf.Lib().IsUnspecified():
			return nil, errors.New("lib provider not specified by logging configuration")
		case !conf.Lib().IsGoKit():
			return nil, fmt.Errorf("%w; lib provider '%s' specified by logging configuration isn't supported or unknown", errs.ErrNotSupported, conf.Lib())
		}

		opts := gokit.Option().
			Conf(*conf).
			AppCmdName(func() string { return attrs.CmdName }).
			AppName(func() string { return attrs.Name }).
			AppVersion(func() string { return attrs.Version() })

		return gokit.New(opts)
	}
}
