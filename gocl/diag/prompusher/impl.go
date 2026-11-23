package prompusher

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus/push"

	"rn.m.vo7.io/vo7io/m/gocl/appmeta/info"
	"rn.m.vo7.io/vo7io/m/gocl/diag/prommetric"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type impl struct {
	fopt.LogOpt
	closeSig       chan struct{}
	conf           Conf
	closeRequested atomic.Bool
	startRequested atomic.Bool
}

func (i *impl) Start(stop context.Context) (err error) {
	if i.closeRequested.Load() {
		err = errs.ErrCloseRequested

		return
	}

	if i.startRequested.Swap(true) {
		err = errs.ErrStartRequested

		return
	}

	go func() {
		p := push.New(i.conf.GatewayAddress, info.CmdAppQualifiedName()).
			Gatherer(prommetric.DefaultRegistry())

		ticker := time.NewTicker(i.conf.Interval)
		defer ticker.Stop()

		exit := false
		for !exit {
			select {
			case <-stop.Done():
				exit = true
			case <-i.closeSig:
				exit = true
			case <-ticker.C:
			}

			pushCtx, pushCtxCancel := context.WithTimeout(context.Background(), i.conf.Timeout)
			if pushErr := p.PushContext(pushCtx); pushErr != nil {
				i.Log().
					Err(stop, pushErr).
					Msgf("failed to perform push operation; gateway: '%s'", i.conf.GatewayAddress)
			} else {
				i.Log().
					Trace(stop).
					Msgf("push operation succeeded; gateway: '%s'", i.conf.GatewayAddress)
			}

			pushCtxCancel()
			ticker.Reset(i.conf.Interval)
		}
	}()

	return
}

func (i *impl) Close() (err error) {
	if i.closeRequested.Swap(true) {
		return
	}

	close(i.closeSig)

	return nil
}

func New(opts ...OptionFn) (_ Pusher, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{}
	if err = fopt.Apply(impl, opts...); err != nil {
		return
	}

	impl.closeSig = make(chan struct{})

	return impl, nil
}
