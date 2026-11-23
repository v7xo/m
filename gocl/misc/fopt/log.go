package fopt

import (
	"strconv"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
)

type (
	LogSetter interface{ SetLog(l logext.Logger) error }

	LogSetterConstraint[T any] interface {
		*T
		LogSetter
	}

	LogOpt struct {
		log logext.Logger
	}
)

var _ = WithLog[LogOpt, *LogOpt, func(*LogOpt) error](nil, nil)

func (s *LogOpt) SetLog(log logext.Logger) error {
	if log == nil {
		log = logext.NopLogger
	}

	s.log = log

	return nil
}

func (s *LogOpt) SetLogScope(n string) error {
	return s.SetLog(s.Log().Tune().AddAttrs(attrname.Dict{attrname.LogScope: n}).Tuned())
}

func (s *LogOpt) AddLogAttrs(vals ...any) error {
	if len(vals) == 0 {
		return nil
	}

	dict := make(attrname.Dict, len(vals)/2+1)

	for i := 0; i < len(vals); i += 2 {
		var (
			n attrname.Name
			v any
		)

		switch name := vals[i].(type) {
		case attrname.Name:
			n = name
		case string:
			n = attrname.Name(name)
		case interface{ String() string }:
			n = attrname.Name(name.String())
		default:
			n = attrname.Name(string(attrname.Unnamed) + strconv.Itoa(i/2+1))
		}

		if i != len(vals)-1 {
			v = vals[i+1]
		}

		dict[n] = v
	}

	return s.SetLog(s.Log().Tune().AddAttrs(dict).Tuned())
}

func (s *LogOpt) Log() logext.Logger {
	if s.log == nil {
		return logext.NopLogger
	}

	return s.log
}

func (s *LogOpt) LogCoalesce(val ...logext.Logger) logext.Logger {
	var log logext.Logger

	for i := -1; i < len(val); i++ {
		if i == -1 {
			log = s.log
		} else {
			log = val[i]
		}

		if log == nil {
			log = logext.NopLogger
		}

		if !log.IsNop() {
			break
		}
	}

	return log
}

func WithLog[T any, TLogSetter LogSetterConstraint[T], TOption ~func(TLogSetter) error](o TOption, log logext.Logger) TOption {
	return func(t TLogSetter) error {
		if err := o(t); err != nil {
			return err
		}

		if log == nil {
			log = logext.NopLogger
		}

		return t.SetLog(log)
	}
}

func WithLogScope[T any, TLogSetter LogSetterConstraint[T], TOption ~func(TLogSetter) error](o TOption, log logext.Logger, scope string) TOption {
	return func(t TLogSetter) error {
		if err := o(t); err != nil {
			return err
		}

		if log == nil {
			log = logext.NopLogger
		}

		if scope == "" {
			return t.SetLog(log)
		}

		return t.SetLog(log.Tune().AddAttrs(attrname.Dict{attrname.LogScope: scope}).Tuned())
	}
}
