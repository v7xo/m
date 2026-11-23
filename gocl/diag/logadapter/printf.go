package logadapter

import (
	"context"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

type (
	printfNop struct{}

	printf struct {
		ctx        context.Context
		underlying logext.Logger
		level      level.Num
	}
)

var printfNopSingleton = &printfNop{} // nolint:forbidigo

func (p *printfNop) Printf(format string, v ...any) {}

func (p *printf) Printf(format string, v ...any) {
	ctx := p.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	p.underlying.
		Level(ctx, p.level, nil).
		Msgf(format, v...)
}

func Printf(log logext.Logger, lev level.Name) interface{ Printf(format string, v ...any) } {
	if !lev.IsKnown() {
		panic(errs.ErrLogUnknownLevel)
	}

	if log == nil || log.IsNop() {
		return printfNopSingleton // nolint:forbidigo
	}

	return &printf{underlying: log, level: lev.Num()} // nolint:forbidigo
}

func PrintfContext(ctx context.Context, log logext.Logger, lev level.Name) interface{ Printf(format string, v ...any) } {
	if !lev.IsKnown() {
		panic(errs.ErrLogUnknownLevel)
	}

	if log == nil || log.IsNop() {
		return printfNopSingleton // nolint:forbidigo
	}

	return &printf{underlying: log, level: lev.Num(), ctx: ctx} // nolint:forbidigo
}
