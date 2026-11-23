package appmodule

import (
	"context"
)

// Module represents application module. Module is a domain specific work unit: http-server, kafka-consumer, etc.
type Module interface {
	// Name returns the module name.
	Name() string
	// Run runs the module. Call is blocking.
	Run(ctx context.Context) error
}
