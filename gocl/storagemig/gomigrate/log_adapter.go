package gomigrate

import (
	"context"

	gomig "github.com/golang-migrate/migrate/v4"

	attrname "rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
)

type logAdapter struct {
	log logext.Logger
	ctx context.Context //nolint:containedctx
}

func (a *logAdapter) Printf(format string, v ...any) {
	a.log.Info(a.ctx).Str(attrname.LogOrigin, "lib_gomigrate").Msgf(format, v...)
}

func (a *logAdapter) Verbose() bool {
	return a.log.LogLevel().IsTraceOrDebug()
}

func newLogAdapter(l logext.Logger, eventContext context.Context) gomig.Logger {
	if l == nil {
		l = logext.NopLogger
	}

	return &logAdapter{log: l, ctx: eventContext}
}
