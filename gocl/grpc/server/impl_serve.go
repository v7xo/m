package server

import (
	"context"
	"errors"
	"fmt"
	"net"

	"google.golang.org/grpc"

	"rn.m.vo7.io/vo7io/m/gocl/misc/ioext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

func (s *impl) Serve(stop context.Context) error {
	return s.serve(nil, stop)
}

func (s *impl) ServeWithListener(ln net.Listener, stop context.Context) (err error) {
	return s.serve(ln, stop)
}

func (s *impl) serve(grpcLn net.Listener, stop context.Context) error {
	if err := s.initializeGRPC(stop); err != nil {
		return err
	}

	return s.serveGRPC(grpcLn, stop)
}

func (s *impl) serveGRPC(ln net.Listener, stop context.Context) (err error) {
	if ln == nil {
		if ln, err = createListener(stop, s.grpcListenHost, s.listenPort); err != nil {
			err = fmt.Errorf("error while creating listener for gRPC-server; %w", err)

			return
		}

		defer ioext.CloseWithLog(context.Background(), ln, s.Log())
	}

	if !s.grpcSrvLn.CompareAndSwap(nil, &ln) {
		err = fmt.Errorf("other listener already assigned to gRPC server; %w", errs.ErrInvalidOperation)

		return
	}

	localStop, localStopSig := context.WithCancel(stop)

	go func() {
		<-localStop.Done()
		s.grpcSrvReadiness.Store(false)

		s.Log().Info(stop).Msgf("stop context signal received; attempt to stop gRPC-server gracefully on '%s' ...", ln.Addr())

		s.grpcSrv.GracefulStop()
	}()

	s.Log().Info(stop).Msgf("start serving gRPC-server on '%s' ...", ln.Addr())
	s.grpcSrvReadiness.Store(true)

	if serveErr := s.grpcSrv.Serve(ln); serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
		err = fmt.Errorf("error while serving GRPC-server on %s; %w", ln.Addr(), serveErr)
	}

	localStopSig()

	return
}
