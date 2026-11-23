//go:build test_unit

package backoff_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/backoff"
	hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_ExpBackoff(t *testing.T) {
	ctx, log := hlp.PreLog(t)

	baseDelay := time.Second * 2
	maxDelay := time.Second * 60
	jitter := float64(.2)
	mul := float64(1.6)

	bof := backoff.Exponential{
		Conf: backoff.ExponentialConfig{
			BaseDelay:  baseDelay,
			Multiplier: mul,
			Jitter:     jitter,
			MaxDelay:   maxDelay,
		},
	}

	b := make([]time.Duration, 100)

	for i := range b {
		b[i] = bof.Backoff(i)
	}

	for i := range b {
		log.Info(ctx).Msgf("backoff num. %d is %s", i+1, b[i].String())

		require.True(t, b[i] <= maxDelay, "backoff num. %d unexpectedly greater than max. %s", i, maxDelay)
	}
}
