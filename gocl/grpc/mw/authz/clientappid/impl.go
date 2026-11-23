package clientappid

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type impl struct {
	filter                  map[string]struct{}
	bypassMethodRegex       *regexp.Regexp
	altGetter               AltGetter
	bypassMethodRegexString string
	require                 bool
	enabled                 bool
}

func (r *impl) unary(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if !r.enabled || (r.bypassMethodRegex != nil && r.bypassMethodRegex.MatchString(info.FullMethod)) {
		return handler(ctx, req)
	}

	id := ""
	reqMd, exists := metadata.FromIncomingContext(ctx)

	if exists {
		var vals []string

		vals, exists = reqMd[string(attrname.XUserAgent)]

		if exists {
			switch {
			case len(vals) == 0 || (len(vals) == 1 && len(vals[0]) == 0):
				return nil, errs.BadArgsWrapWith(errors.New("invalid value of metadata; metadata value can't be an empty string; metadata: " + string(attrname.XUserAgent)))
			case len(vals) == 1:
				id = vals[0]
			default:
				return nil, errs.BadArgsWrapWith(errors.New("invalid value of metadata; too many items in value; metadata: " + string(attrname.XUserAgent)))
			}
		}
	}

	switch {
	case r.altGetter != nil:
		altID, altExists, err := r.altGetter(ctx, req, info)
		if err != nil {
			return nil, fmt.Errorf("failed to get client app ID using alternative way; %w", err)
		}

		switch {
		case exists:
			if altExists && altID != id {
				return nil, errs.BadArgsWrapWith(errors.New("client app ID specified by alternative way not equals to client app ID specified in call metadata; metadata: " + string(attrname.XUserAgent)))
			}
		case altExists:
			if len(altID) == 0 {
				return nil, errs.BadArgsWrapWith(errors.New("invalid value of client app ID specified by alternative way"))
			}

			id, exists = altID, true
		}
	}

	switch {
	case exists:
		if len(r.filter) != 0 {
			if _, ok := r.filter[id]; !ok {
				return nil, errs.ErrUnauthorized
			}
		}
	case r.require:
		return nil, errs.ErrUnauthorized
	}

	return handler(ctx, req)
}

func New(opts ...OptionFn) (grpc.UnaryServerInterceptor, error) {
	const errMsgTemplate = "failed to create instance of gRPC server interceptor for incoming client app ID; %w"

	res := &impl{}

	if err := fopt.Apply(res, opts...); err != nil {
		return nil, fmt.Errorf(errMsgTemplate, err)
	}

	switch {
	case res.bypassMethodRegexString != "":
		var err error

		if res.bypassMethodRegex, err = regexp.Compile(res.bypassMethodRegexString); err != nil {
			return nil, fmt.Errorf(errMsgTemplate, fmt.Errorf("failed to compile regular expression for gRPC method(s) to be skipped; %w", err))
		}
	}

	return res.unary, nil
}
