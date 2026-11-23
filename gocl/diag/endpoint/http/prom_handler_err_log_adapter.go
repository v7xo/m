package http

import (
	"context"
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
)

type promHandlerErrorLogAdapter struct {
	logEvtCtx context.Context //nolint:containedctx
	log       logext.Logger
}

func (adp *promHandlerErrorLogAdapter) Println(a ...any) {
	adp.log.
		Err(adp.logEvtCtx, nil).
		Str(attrname.LogOrigin, "prom_http_handler").
		Msg(fmt.Sprint(a...))
}
