package backoff

import (
	"math/rand"
	"time"
)

// Exponential implements exponential backoff algorithm.
type Exponential struct {
	// Conf contains all options to configure the backoff algorithm.
	Conf ExponentialConfig
}

// Backoff returns the amount of time to wait before the next retry.
//
//nolint:predeclared
func (v *Exponential) Backoff(doneRetries int) time.Duration {
	if v == nil {
		return 0
	}

	b := float64(v.Conf.BaseDelay)
	max := float64(v.Conf.MaxDelay)

	for range doneRetries {
		b = b * v.Conf.Multiplier
	}

	if b >= max {
		b = max
	}

	rand := rand.New(rand.NewSource(time.Now().UnixNano())) // nolint:gosec

	const maxIter = 7

	for range maxIter {
		j := v.Conf.Jitter * (rand.Float64()*2.0 - 1.0)
		r := b + b*j

		if r <= 0.0 {
			return 0
		} else if r < max {
			return time.Duration(r)
		}
	}

	return time.Duration(max)
}
