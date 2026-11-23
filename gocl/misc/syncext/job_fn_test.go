package syncext_test

import (
	"math/rand/v2"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/syncext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
	"rn.m.vo7.io/vo7io/m/gocl/misc/uuid_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_JobVoid_Partitioning(t *testing.T) {
	ctx, log := thelper.PreLog(t)

	// nolint:thelper
	testFn := func(t *testing.T, maxPartitions syncext.MaxDOP) {
		dop := syncext.MaxDOP(maxPartitions.MustEffective())
		idleTimeout := time.Millisecond * 10
		itemsCount := int(dop) * 80                                           // nolint:revive
		simulateTime := time.Millisecond * 15                                 // nolint:revive
		testTimeout := simulateTime*time.Duration(itemsCount) + time.Second*5 // nolint:revive

		if thelper.IsDebuggerAttached(t) {
			testTimeout = time.Hour
		}

		inCh := make(chan syncext.Job[struct{}], itemsCount)

		type (
			partitionState struct {
				key           string
				checkpoint    atomic.Int32
				expectedOrder []int
				actualOrder   []int
			}

			item struct {
				execMarker atomic.Int32
				id         int
			}
		)

		partitions := make([]*partitionState, 0, dop)

		for range dop {
			partitions = append(
				partitions,
				&partitionState{
					key: uuid_thelper.Rand8BHex(t),
				},
			)
		}

		items := make([]*item, 0, itemsCount)

		for i := range itemsCount {
			items = append(items, &item{id: i})
		}

		var (
			totalExecutedCount int32
			anyPartitionCount  int32
		)

		for i := range items {
			item := items[i]
			itemID := item.id
			anyPartition := rand.Int()%3 == 0 // nolint:revive

			var (
				partitionState *partitionState
				partitionKey   syncext.JobPartitionKey
			)

			if anyPartition {
				anyPartitionCount++
				partitionKey = syncext.JobPartitionKeyAny
			} else {
				partitionState = partitions[rand.IntN(int(dop))]
				partitionState.expectedOrder = append(partitionState.expectedOrder, itemID)
				partitionKey = syncext.JobPartitionKey(partitionState.key)
			}

			inCh <- syncext.NewJobVoid(
				func() {
					atomic.AddInt32(&totalExecutedCount, 1)

					switch item.execMarker.Add(1) {
					case 1:
						// correct.
					default:
						require.Fail(t, "", "item '%d' executed twice as at least", itemID)

						return
					}

					if !anyPartition {
						cp := partitionState.checkpoint.Add(1)
						switch cp {
						case 1:
							// correct.
						default:
							require.Fail(t, "", "race condition check not passed", "partition key '%s'", partitionState.key)

							return
						}
					}

					time.Sleep(timeout.MustJitteredDefault(simulateTime)) // simulate work

					if !anyPartition {
						partitionState.actualOrder = append(partitionState.actualOrder, itemID)

						cp := partitionState.checkpoint.Add(-1)
						switch cp {
						case 0:
							// correct.
						default:
							require.Fail(t, "", "race condition check not passed", "partition key '%s'", partitionState.key)
						}
					}
				},
				partitionKey,
			)
		}

		close(inCh)

		log.Info(ctx).Msg("begin start of items execution")

		outCh := syncext.SchedVoid(inCh, dop, idleTimeout)

		log.Info(ctx).Msg("end start of items execution")
		log.Info(ctx).Msgf("total items: %d, any partition items: %d", len(items), anyPartitionCount)

		log.Info(ctx).Msgf("wait for completion '%s' ...", testTimeout)

		select {
		case <-ctx.Done():
			require.Fail(t, "test context canceled")
		case <-time.After(testTimeout):
			require.Fail(t, "test timeout reached")
		case <-outCh:
		}

		if t.Failed() {
			return
		}

		require.Equal(t, len(items), int(totalExecutedCount))

		for k, p := range partitions {
			require.ElementsMatch(t, p.expectedOrder, p.actualOrder, "expected set of items not matches actual one for partition '%s'", k)
			require.Equal(t, p.expectedOrder, p.actualOrder, "expected order of items not matches actual one for partition '%s'", k)
		}
	}

	t.Run(
		"max. partitions - auto",
		func(t *testing.T) {
			t.Parallel()

			testFn(t, syncext.MaxDOPAuto)
		},
	)
	t.Run(
		"max. partitions - 1",
		func(t *testing.T) {
			t.Parallel()

			testFn(t, 1)
		},
	)
}
