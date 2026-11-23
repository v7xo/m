package query

import (
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
)

type OptionFn func(*Named) error

func (o OptionFn) Evaluatable(val bool) OptionFn {
	return fopt.Chain(o, func(t *Named) { t.evaluatable = val })
}

func (o OptionFn) EvaluationOutputBuffer(val int) OptionFn {
	return fopt.Chain(o, func(t *Named) { t.evalOutputBufCap = val })
}

func (o OptionFn) Timeout(val timeout.Timeout) OptionFn {
	return fopt.Chain(o, func(t *Named) { t.timeout = val })
}

func (o OptionFn) ParamQualifier(val ParamQualifier) OptionFn {
	return fopt.Chain(o, func(t *Named) { t.paramQualifier = val })
}

// AtSign specifies use [AtSignParamQualifier] as param name qualifier.
func (o OptionFn) AtSign() OptionFn { return o.ParamQualifier(AtSignParamQualifier) }

func (o OptionFn) PositionalParams(val bool) OptionFn {
	return fopt.Chain(o, func(t *Named) { t.hasPositionalSyntax = val })
}

func Option() OptionFn { return fopt.Nop[Named] }
