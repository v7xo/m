package contextext

import (
	"context"

	"rn.m.vo7.io/vo7io/m/gocl/misc/uuidext"
)

func RequestID(ctx context.Context) (reqID string, ok bool, err error) {
	return Val[string](ctx, KeyRequestID)
}

func WithRequestID(ctx context.Context) (res context.Context, err error) {
	_, ok, err := RequestID(ctx)
	if err != nil {
		return
	} else if ok {
		res = ctx

		return
	}

	reqID, err := uuidext.RequestIDHex()
	if err != nil {
		return
	}

	res = WithRequestIDReplace(ctx, reqID)

	return
}

func WithRequestIDReplace(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, KeyRequestID, id)
}
