package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"

	"rn.m.vo7.io/vo7io/m/gocl/app/appmodule"
	"rn.m.vo7.io/vo7io/m/gocl/diag/readiness"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/mw/cors"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
)

type impl struct {
	fopt.LogOpt
	handler                   http.Handler
	corsOpts                  cors.OptionFn
	beforeStartServe          DriverFn
	listenHost                string
	listenPort                int
	shutdownTimeout           timeout.Timeout
	readiness                 atomic.Bool
	preventErrDetailedMessage bool
}

var (
	_ readiness.Collector
	_ appmodule.Module
)

// Run implements [appmodule.Module.Run] app module run.
func (s *impl) Run(ctx context.Context) error {
	return s.Serve(ctx)
}

// Name implements [appmodule.Module.Name] - app module name.
func (s *impl) Name() string {
	return AppModuleName
}

// CollectReadinessConsolidated implements [readiness.Collector].
func (s *impl) CollectReadinessConsolidated(_ context.Context) (bool, error) {
	return s.readiness.Load(), nil
}

func New(opts ...OptionFn) (_ Server, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{}
	if err = fopt.Apply(impl, opts...); err != nil {
		return
	}

	if impl.handler == nil {
		return nil, errors.New("no HTTP-handler specified")
	}

	return impl, nil
}
