package backoff

import "time"

// Strategy defines the methodology for backing off after a http connection
// failure.
type Strategy interface {
	// Backoff returns the amount of time to wait before the next retry.
	Backoff(doneRetries int) time.Duration
}
