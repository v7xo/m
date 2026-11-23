package clientappid

import (
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*impl) error

func (o OptionFn) Require(val bool) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.require = val })
}

func (o OptionFn) Filter(val []string) OptionFn {
	return fopt.Chain(
		o,
		func(t *impl) {
			if len(val) == 0 {
				t.filter = nil

				return
			}

			t.filter = make(map[string]struct{}, len(val))

			for y := range val {
				if item := strings.TrimSpace(val[y]); len(item) != 0 {
					t.filter[item] = struct{}{}
				}
			}

			return
		},
	)
}

func (o OptionFn) Enabled(val bool) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.enabled = val })
}

func (o OptionFn) BypassMethodRegex(val string) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.bypassMethodRegexString = val })
}

func (o OptionFn) AltGetter(val AltGetter) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.altGetter = val })
}

func (o OptionFn) Conf(val Conf) OptionFn {
	return fopt.ChainErr(
		o,
		func(t *impl) error {
			return Option().
				Enabled(val.Enabled).
				Require(val.Require).
				BypassMethodRegex(val.BypassMethodRegex).
				Filter(strings.Split(val.Filter, ","))(t)
		},
	)
}

func Option() OptionFn { return fopt.Nop[impl] }
