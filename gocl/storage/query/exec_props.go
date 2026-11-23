package query

import (
	"context"
	"fmt"

	clcontext "rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type ExecProps struct {
	name    string
	timeout timeout.Timeout
}

func (qp ExecProps) Name() string             { return qp.name }
func (qp ExecProps) Timeout() timeout.Timeout { return qp.timeout }

func ExecPropsContext(p context.Context, name string, timeout timeout.Timeout) context.Context {
	return context.WithValue(p, clcontext.KeyStorageQueryExecProps, ExecProps{name: name, timeout: timeout})
}

func ExecPropsFromContext(ctx context.Context) (res ExecProps, ok bool, err error) {
	if ctx == nil {
		return
	}

	ctxVal := ctx.Value(clcontext.KeyStorageQueryExecProps)
	if ctxVal == nil {
		return
	}

	if res, ok = ctxVal.(ExecProps); ok {
		return
	}

	err = fmt.Errorf(
		"such type of context's value is not expected; value key: %s; actual value type: %T; expected value type: %T; err: %w",
		clcontext.KeyStorageQueryExecProps,
		ctxVal,
		ExecProps{},
		errs.ErrUnexpectedTypeOfValueInContext)

	return
}
