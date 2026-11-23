package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logadapter"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/mw/cors"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/mw/metrics"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/mw/recovery"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/mw/xrequestid"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

func (s *impl) Serve(stop context.Context) error { return s.serve(stop, nil) }

func (s *impl) ServeWithListener(ln net.Listener, stop context.Context) error {
	return s.serve(stop, ln)
}

func (s *impl) serve(stop context.Context, ln net.Listener) error {
	h := s.handler

	if s.corsOpts != nil {
		corsMW, err := cors.New(s.corsOpts)
		if err != nil {
			return fmt.Errorf("failed to construct HTTP server CORS middleware; %w", err)
		}

		h = corsMW.Wrap(h)

		s.Log().Warn(stop).Msg("HTTP-server will use CORS middleware")
	}

	h = xrequestid.NewWithLog(s.Log()).Wrap(h)
	h = recovery.NewWithLog(s.Log(), s.preventErrDetailedMessage).Wrap(h)
	h = metrics.New(s.Name()).Wrap(h)

	driver := &http.Server{
		Addr:        net.JoinHostPort(s.listenHost, strconv.Itoa(s.listenPort)),
		Handler:     h,
		BaseContext: func(net.Listener) context.Context { return contextext.WithLog(stop, s.Log()) },
	}

	if ln != nil {
		driver.Addr = ln.Addr().String()
	}

	if !s.Log().IsNop() {
		driver.ErrorLog = log.New(logadapter.Writer(s.Log(), stop, level.Error, nil), "", 0)
	}

	switch {
	case s.beforeStartServe != nil:
		if err := s.beforeStartServe(stop, driver); err != nil {
			return fmt.Errorf("error while calling HTTP-server (driver) function (before start serving); %w", err)
		}
	}

	listenAndServStop := make(chan error, 1)

	defer s.readiness.Store(false)

	go func() {
		s.Log().Info(stop).Msgf("http server: start serving on '%s' ...", driver.Addr)
		s.readiness.Store(true)

		var serveErr error

		if ln != nil {
			serveErr = driver.Serve(ln)
		} else {
			serveErr = driver.ListenAndServe()
		}

		s.readiness.Store(false)

		if errors.Is(serveErr, http.ErrServerClosed) {
			serveErr = nil
		} else if serveErr != nil {
			s.Log().Err(stop, serveErr).Msg("http server: failed to listen and serve")

			serveErr = fmt.Errorf("failed to listen and serve HTTP-server; %w", serveErr)
		}

		listenAndServStop <- serveErr
	}()

	collectedErrs := make([]error, 0, 2) //nolint

	select {
	case <-stop.Done():
		s.Log().Trace(stop).Err(stop.Err()).Msg("http server: stop context signal received")

		shutdownCtx := context.Background()

		if !s.shutdownTimeout.IsInfinite() {
			var cancel context.CancelFunc

			shutdownCtx, cancel = context.WithTimeout(shutdownCtx, s.shutdownTimeout.Dur())

			defer cancel()
		}

		shutdownErr := driver.Shutdown(shutdownCtx)
		if shutdownErr != nil {
			s.Log().
				Err(stop, shutdownErr).
				Msg("http server: shutdown failed")

			collectedErrs = append(collectedErrs, fmt.Errorf("error while shutting down HTTP server; %w", shutdownErr))
		}

		collectedErrs = append(collectedErrs, <-listenAndServStop)
	case listenAndServeErr := <-listenAndServStop:
		collectedErrs = append(collectedErrs, listenAndServeErr)
	}

	return errs.Join(collectedErrs...)
}
