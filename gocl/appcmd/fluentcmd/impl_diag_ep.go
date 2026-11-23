package fluentcmd

import (
	"context"
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/diag/endpoint"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
)

func (s *impl[TConf]) diagnosticsEndpointStart(log logext.Logger) (context.CancelFunc, error) {
	var stopSig context.CancelFunc

	stopSigCaller := func() {
		if stopSig != nil {
			stopSig()
		}
	}

	if s.diagnosticsEndpointFactoryFn == nil || !s.confSupplied {
		return stopSigCaller, nil
	}

	var confGetter endpoint.ConfGetter

	if t, ok := any(s.conf).(endpoint.ConfGetter); ok {
		confGetter = t
	}

	var (
		svc endpoint.Service
		err error
	)

	if svc, err = s.diagnosticsEndpointFactoryFn(confGetter, log); err != nil {
		return stopSigCaller, fmt.Errorf("failed to construct diagnostics endpoint service; %w", err)
	}

	stopCtx, stopSig := context.WithCancel(context.Background())

	defer func() {
		if err != nil {
			stopSig()
		}
	}()

	//nolint:contextcheck
	if err = svc.Start(stopCtx); err != nil {
		return stopSigCaller, fmt.Errorf("failed to start diagnostics endpoint service; %w", err)
	}

	return stopSigCaller, nil
}
