package gokit

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	log "rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

type OptionFn func(*impl) error

func (o OptionFn) AppCmdName(v func() string) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.appCmdName = v })
}

func (o OptionFn) AppVersion(v func() string) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.appVersion = v })
}

func (o OptionFn) AppName(v func() string) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.appName = v })
}

func (o OptionFn) NonErrorLevelWriter(w io.Writer) OptionFn {
	return fopt.ChainErr(
		o,
		func(t *impl) error {
			if w == nil {
				return errors.New("writer can't be nil")
			}

			t.nonErrLevelWriter = w

			return nil
		},
	)
}

func (o OptionFn) ErrLevelWriter(w io.Writer) OptionFn {
	return fopt.ChainErr(
		o,
		func(t *impl) error {
			if w == nil {
				return errors.New("writer can't be nil")
			}

			t.errLevelWriter = w

			return nil
		},
	)
}

func (o OptionFn) Level(v level.Name) OptionFn {
	return fopt.ChainErr(
		o,
		func(t *impl) error {
			if !v.IsKnown() {
				return fmt.Errorf("unknown log level '%s'", v)
			}

			t.level = v.Num()

			return nil
		},
	)
}

func (o OptionFn) Caller(v bool) OptionFn {
	return fopt.Chain(
		o,
		func(t *impl) {
			if v {
				t.callerInfo = makeCallerInfo(defaultCallerFnFileDepth)
			} else {
				t.callerInfo = nil
			}
		},
	)
}

func (o OptionFn) Format(v log.Format) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.format = v })
}

func (o OptionFn) Conf(v log.Conf) OptionFn {
	return fopt.ChainErr(o, func(t *impl) error { return applyConf(t, v) })
}

func applyConf(i *impl, cfg log.Conf) (err error) {
	if !cfg.Lib().IsGoKit() {
		return fmt.Errorf("lib provider denoted by specified logging configuration is '%s' whereas '%s' expected", cfg.LibString, log.LibProviderGoKit)
	}

	i.callerInfo = makeCallerInfo(defaultCallerFnFileDepth)
	i.format = cfg.Format()
	i.level = cfg.Level().Num()

	inclExclFn := func(pattern, action string) (func(attrname.Name) bool, error) {
		if withoutPrefix, ok := strings.CutPrefix(pattern, log.AttrInclExclPatternREDesignator); ok {
			//nolint:revive
			if pattern = strings.TrimSpace(withoutPrefix); pattern == "" {
				return nil, fmt.Errorf("attribute %s regular expression can't be an empty string", action)
			} else if r, compileErr := regexp.Compile(pattern); compileErr != nil {
				return nil, fmt.Errorf("failed to compile attribute %s regular expression '%s'; err: %w", action, pattern, compileErr)
			} else {
				return func(n attrname.Name) bool { return r.MatchString(string(n)) }, nil
			}
		}

		splitted := strings.Split(pattern, ",")
		dict := make(map[string]struct{}, len(splitted))

		for i := range len(splitted) {
			if n := strings.TrimSpace(splitted[i]); n == "" {
				continue
			} else {
				dict[n] = struct{}{}
			}
		}

		if len(dict) == 0 {
			return nil, nil //nolint
		}

		return func(n attrname.Name) bool {
				_, ok := dict[string(n)]

				return ok
			},
			nil
	}

	if i.excludeAttrFn, err = inclExclFn(cfg.AttributeExclude, "exclusion"); err != nil {
		return
	}

	i.includeAttrFn, err = inclExclFn(cfg.AttributeInclude, "inclusion")

	return
}

func Option() OptionFn { return fopt.Nop[impl] }
