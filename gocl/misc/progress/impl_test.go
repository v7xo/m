//go:build test_unit

package progress_test

import (
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/progress"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
)

func Test_Progress(t *testing.T) {
	synctest.Test(
		t,
		func(t *testing.T) {
			coll := make([]int, 0, 100)

			handler := func(val int) {
				synctest.Sleep(timeout.MustJitteredDefault(time.Second * 3))

				coll = append(coll, val)
			}

			prog, err := progress.New(progress.Option[int]().Capacity(100).Handler(handler))
			require.NoError(t, err)
			require.NotNil(t, prog)

			wg := sync.WaitGroup{}

			wg.Go(
				func() {
					for i := range 1000 {
						synctest.Sleep(timeout.MustJitteredDefault(time.Second * 5))

						prog.Report(i)
					}
				})

			wg.Wait()

			prog.Stop(true)
		})
}
