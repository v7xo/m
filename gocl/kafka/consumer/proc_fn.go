package consumer

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type (
	procFn interface{ isProcFn() }

	YieldNextProcFn func(ctx context.Context, msg kafka.Message) (yieldNext bool)
)

var _ procFn = YieldNextProcFn(nil)

func (YieldNextProcFn) isProcFn() {}
