package logadapter

import (
	"context"
	"io"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

type writerNop struct{}

var writerNopSingleton = &writerNop{}

func (w *writerNop) Write(p []byte) (n int, err error) { return len(p), nil }

type writer struct {
	underlying logext.Logger
	event      context.Context
	attr       attrname.Dict
	level      level.Num
}

func (w *writer) Write(p []byte) (n int, err error) {
	w.underlying.
		Level(w.event, w.level, nil).
		Attrs(w.attr).
		Msg(string(p))

	return len(p), nil
}

func Writer(l logext.Logger, e context.Context, n level.Name, a attrname.Dict) io.Writer {
	if !n.IsKnown() {
		panic(errs.ErrLogUnknownLevel)
	}

	if l == nil || l.IsNop() {
		return writerNopSingleton
	}

	if e == nil {
		e = context.Background()
	}

	return &writer{underlying: l, level: n.Num(), attr: a, event: e}
}
