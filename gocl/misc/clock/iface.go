package clock

import (
	"context"
	"time"
)

type Clock interface {
	// CurrentTimestamp returns current time instant as it defined by this service.
	CurrentTimestamp(ctx context.Context) time.Time

	// Now is shorthand of [Service.CurrentTimestamp].
	Now(ctx context.Context) time.Time
}
