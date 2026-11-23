package readiness

import "context"

type Collector interface {
	// CollectReadinessConsolidated returns consolidated result of readiness.
	CollectReadinessConsolidated(ctx context.Context) (bool, error)
}
