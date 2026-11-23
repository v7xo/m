package server

import (
	"bytes"
	"context"
	"encoding/json"

	"google.golang.org/grpc/metadata"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	clgrpc "rn.m.vo7.io/vo7io/m/gocl/grpc"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ioext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
)

func logEventEnrichWithReqRespBody(e logext.Event, req any, resp any) logext.Event {
	wBuf := &bytes.Buffer{}
	w := ioext.NewDiscardOverlimit(logext.DefaultRequestBodyFieldLimitBytes, wBuf, false)

	if reqEncErr := json.NewEncoder(w).Encode(req); reqEncErr != nil {
		e = e.Str(attrname.RequestBody, "<"+reqEncErr.Error()+">")
	} else if w.TotalDiscarded() == 0 {
		e = e.Str(attrname.RequestBody, wBuf.String())
	} else {
		e = e.Str(attrname.RequestBody, wBuf.String()+placeholder.Overlimit)
	}

	wBuf.Reset()
	w = ioext.NewDiscardOverlimit(logext.DefaultResponseBodyFieldLimitBytes, wBuf, false)

	if respEncErr := json.NewEncoder(w).Encode(resp); respEncErr != nil {
		e = e.Str(attrname.ResponseBody, "<"+respEncErr.Error()+">")
	} else if w.TotalDiscarded() == 0 {
		e = e.Str(attrname.ResponseBody, wBuf.String())
	} else {
		e = e.Str(attrname.ResponseBody, wBuf.String()+placeholder.Overlimit)
	}

	return e
}

func logEventEnrichWithGRPCIncomingMD(e logext.Event, md metadata.MD) logext.Event {
	wBuf := &bytes.Buffer{}
	w := ioext.NewDiscardOverlimit(logext.DefaultRequestHeadersFieldLimitBytes, wBuf, false)

	if encErr := json.NewEncoder(w).Encode(md); encErr != nil {
		e = e.Str(attrname.RequestHeaders, "<"+encErr.Error()+">")
	} else if w.TotalDiscarded() == 0 {
		e = e.Str(attrname.RequestHeaders, wBuf.String())
	} else {
		e = e.Str(attrname.RequestHeaders, wBuf.String()+placeholder.Overlimit)
	}

	return e
}

func (s *impl) logEventEnrichGRPC(_ context.Context, _ level.Num, _ error, _ bool, e logext.Event) logext.Event {
	e = e.
		Str(attrname.TransportProto, "http2").
		Str(attrname.APICallStyle, clgrpc.GRPCAPICallStyle).
		Str(attrname.LogScope, LogScope).
		Str(attrname.TransportEndpointRole, "server")

	if lnPtr := s.grpcSrvLn.Load(); lnPtr != nil {
		ln := *lnPtr

		e = e.Str(attrname.Addr, ln.Addr().String())
	}

	return e
}
