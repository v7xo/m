package fscopier

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(val logext.Logger) OptionFn { return fopt.WithLogScope(o, val, LogScope) }

func (o OptionFn) PathFilter(val PathFilterFn) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.pathFilter = val })
}

func (o OptionFn) ExistingAction(val ExistingAction) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.existingAction = val })
}

func (o OptionFn) ExistingFilter(val ExistingFilterFn) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.existingFilter = val })
}

func (o OptionFn) FilePermFilter(val PermFilter) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.filePermFilter = val })
}

func (o OptionFn) NameFilter(val NameFilterFn) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.nameFilter = val })
}

func (o OptionFn) ContentFilter(val ContentFilterFn) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.contentFilter = val })
}

func (o OptionFn) DisableDeletedSuffix(val bool) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.disableDeletedSuffix = val })
}

func (o OptionFn) DisableNameUnescaping(val bool) OptionFn {
	return fopt.Chain(o, func(t *impl) { t.disableNameUnescaping = val })
}

func Option() OptionFn { return fopt.Nop[impl] }
