package logext

import (
	"context"

	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
)

func FromContext(ctx context.Context) (Logger, error)         { return contextext.Log(ctx) }
func MustFromContext(ctx context.Context) Logger              { return contextext.MustLog(ctx) }
func Context(ctx context.Context, log Logger) context.Context { return contextext.WithLog(ctx, log) }
