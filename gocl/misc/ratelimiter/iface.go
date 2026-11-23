package ratelimiter

import (
	"context"
	"time"
)

type Service interface {
	WaitNext(ctx context.Context) (ts time.Time, err error)
	TryWaitNext(ctx context.Context) (ok bool, ts time.Time, err error)

	String() string
}
