//go:build test_unit

package slicesext_test

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/slicesext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/syncext"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_ProcPartitions(t *testing.T) {
	ctx := thelper.Pre(t)

	inp := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	mu := &sync.Mutex{}

	outDOPSentinel := &atomic.Int32{}
	outMaxDOP := &atomic.Int32{}

	for range len(inp) {
		outMaxDOP.Store(0)
		outDOPSentinel.Store(0)

		maxDOP := syncext.MaxDOPRand()
		partSize := rand.Int31n(int32(len(inp)*2)) + 1
		workSimulate := time.Millisecond * time.Duration(rand.Int31n(200))

		out := make([]int64, 0, len(inp))

		err := slicesext.ProcPartitions(
			ctx,
			inp,
			int(partSize),
			maxDOP,
			func(ctx context.Context, _ int, p []int64) error {
				n := outDOPSentinel.Add(1)

				defer func() { outDOPSentinel.Add(-1) }()

				for {
					if o := outMaxDOP.Load(); n > o {
						if outMaxDOP.CompareAndSwap(o, n) {
							break
						}
					} else {
						break
					}
				}

				// simulate work
				select {
				case <-ctx.Done():
					return context.Cause(ctx)
				case <-time.After(workSimulate):
				}

				mu.Lock()

				out = append(out, p...)

				mu.Unlock()

				return nil
			},
		)

		require.NoError(t, err)
		require.ElementsMatchf(t, inp, out, "actual max. DOP: %d", outMaxDOP.Load())

		if maxDOP.IsAuto() {
			require.True(t, outMaxDOP.Load() <= int32(len(inp)), "%d <= %d", outMaxDOP.Load(), len(inp))
		} else {
			require.True(t, outMaxDOP.Load() <= int32(maxDOP), "%d <= %d", outMaxDOP.Load(), maxDOP)
		}

	}
}
