package consumer

import "context"

type Service interface {
	RunLoop(ctx context.Context, opts ...RunLoopOptionFn) error
}
