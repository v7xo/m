package logext

import (
	"context"
	"maps"
	"slices"

	attrname "rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

type tuner struct {
	base               Logger
	level              *level.Num
	addAttrs           attrname.Dict
	setAttrs           attrname.Dict
	hooks              []EventHookFn
	notIncludeBaseHook bool
}

var _ Tuner = &tuner{}

func (t *tuner) applyHooks(b Logger) Logger {
	if t.hooks == nil {
		return b
	}

	var baseHook EventHookFn

	if !t.notIncludeBaseHook {
		baseHook = b.Hook()
	}

	cp := slices.Clone(t.hooks)

	var h EventHookFn = func(ctx context.Context, l level.Num, err error, fatal bool, e Event) Event {
		r := e

		if baseHook != nil {
			r = baseHook(ctx, l, err, fatal, r)
		}

		for i := range cp {
			r = t.hooks[i](ctx, l, err, fatal, r)
		}

		return r
	}

	return Hook(b, h)
}

func (t *tuner) applyLevel(b Logger) Logger {
	if t.level == nil {
		return b
	}

	return Level(b, *t.level)
}

func (t *tuner) applyAttrs(b Logger) Logger {
	var setAttrs attrname.Dict

	switch {
	case t.setAttrs == nil:
		if len(t.addAttrs) == 0 {
			return b
		}

		setAttrs = maps.Clone(t.addAttrs)

		for n, v := range b.Attrs() {
			if _, ok := setAttrs[n]; ok {
				continue
			}

			setAttrs[n] = v
		}
	case len(t.setAttrs) != 0:
		setAttrs = maps.Clone(t.setAttrs)
	}

	return Attrs(b, setAttrs)
}

func (t *tuner) Tuned() Logger {
	if t.base.IsNop() {
		return t.base
	}

	r := t.applyAttrs(t.base)
	r = t.applyLevel(r)
	r = t.applyHooks(r)

	return r
}

func (t *tuner) Level(l level.Num) Tuner {
	if !l.IsKnown() {
		panic(errs.ErrLogUnknownLevel)
	}

	if t.base.IsNop() {
		return t
	}

	t.level = &l

	return t
}

func (t *tuner) AddAttrs(a attrname.Dict) Tuner {
	if t.base.IsNop() {
		return t
	}

	if len(a) == 0 {
		return t
	}

	if t.addAttrs == nil {
		t.addAttrs = maps.Clone(a)

		maps.Copy(t.addAttrs, t.setAttrs)

		t.setAttrs = nil

		return t
	}

	maps.Copy(t.addAttrs, a)

	return t
}

func (t *tuner) SetAttrs(a attrname.Dict) Tuner {
	if t.base.IsNop() {
		return t
	}

	if a == nil {
		a = attrname.Dict{}
	}

	t.addAttrs = nil
	t.setAttrs = a

	return t
}

func (t *tuner) NextHook(h EventHookFn) Tuner {
	if t.base.IsNop() {
		return t
	}

	if h == nil {
		return t
	}

	t.hooks = append(t.hooks, h)

	return t
}

func (t *tuner) SetHook(h EventHookFn) Tuner {
	if t.base.IsNop() {
		return t
	}

	if h == nil {
		t.hooks = []EventHookFn{}
		t.notIncludeBaseHook = true

		return t
	}

	t.hooks = []EventHookFn{h}
	t.notIncludeBaseHook = true

	return t
}

func Tune(b Logger) Tuner { return &tuner{base: b} }
