package syncext

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type jobCtrl[TResult any] struct {
	in                   <-chan Job[TResult]
	done                 chan struct{}
	resultsCh            chan TResult
	partitionInactivated chan *jobPartition[TResult]
	workerIdleTimeout    time.Duration
	voidResult           bool
}

func (r *jobCtrl[TResult]) procItem(i Job[TResult]) {
	switch {
	case r.voidResult:
		i.Do()

		return
	}

	r.resultsCh <- i.Do()
}

func (r *jobCtrl[TResult]) doWorker(partition *jobPartition[TResult]) {
	first := partition.first
	if first != nil {
		partition.first = nil

		r.procItem(first)
	}

	var item *backlogJob[TResult]

	for {
		select {
		case item = <-partition.backlog:
		default:
			partition.idle.Reset(r.workerIdleTimeout)

			select {
			case item = <-partition.backlog:
				partition.idle.Stop()
			case <-r.done:
				r.partitionInactivated <- partition

				return
			case <-partition.idle.C:
				r.partitionInactivated <- partition

				return
			}
		}

		if item.taken.Swap(1) != 0 {
			continue
		}

		r.procItem(item.item)
	}
}

// nolint:gocognit
func (r *jobCtrl[TResult]) coordinatorLoop() {
	var (
		inactivatedPartitions = make([]*jobPartition[TResult], len(r.partitionInactivated))
		partitions            = make([]*jobPartition[TResult], len(r.partitionInactivated))

		isEOF            bool
		item             Job[TResult]
		itemPartitionNum int
		wg               sync.WaitGroup
	)

readPartitions:
	for {
		select {
		case p := <-r.partitionInactivated:
			inactivatedPartitions[p.num] = p
			partitions[p.num] = p
		default:
			break readPartitions
		}
	}

	for {
		item = nil

		if isEOF {
			inactivatedPartitionsCount := 0

			for _, p := range inactivatedPartitions {
				if p != nil {
					inactivatedPartitionsCount++
				}
			}

			switch {
			case len(partitions) == inactivatedPartitionsCount:
				close(r.resultsCh)
				close(r.partitionInactivated)

				for _, p := range partitions {
					close(p.backlog)

					p.idle.Stop()
				}

				return
			}

			inactivatedPartition := <-r.partitionInactivated
			inactivatedPartitions[inactivatedPartition.num] = inactivatedPartition

			continue
		}

		i, ok := <-r.in
		switch ok {
		case true:
			if i == nil {
				continue
			}

			item = i
		default:
			isEOF = true

			close(r.done)

			continue
		}

		if item.PartitionKey().Any() {
			itemPartitionNum = -1
		} else {
			itemPartitionNum = int(item.PartitionKey().HashUint64() % uint64(len(partitions)))
		}

		backlogItem := &backlogJob[TResult]{item: item}

	pushLoop:
		for {
			if itemPartitionNum == -1 {
				for _, p := range partitions {
					select {
					case p.backlog <- backlogItem:
						break pushLoop
					default:
					}
				}

				for _, inactivatedPartition := range inactivatedPartitions {
					if inactivatedPartition == nil {
						continue
					}

					inactivatedPartitions[inactivatedPartition.num] = nil

					inactivatedPartition.first = backlogItem.item

					go r.doWorker(inactivatedPartition)

					break pushLoop
				}

				done := make(chan struct{})
				doneMark := int32(0)
				pushed := int32(0)

				for j := range len(partitions) + 1 {
					if j == 0 {
						wg.Go(
							func() {
								defer func() {
									if atomic.SwapInt32(&doneMark, 1) == 0 {
										close(done)
									}
								}()

								select {
								case inactivatedPartition := <-r.partitionInactivated:
									inactivatedPartitions[inactivatedPartition.num] = inactivatedPartition
								case <-done:
								}
							},
						)

						continue
					}

					partition := partitions[j-1]

					wg.Go(
						func() {
							defer func() {
								if atomic.SwapInt32(&doneMark, 1) == 0 {
									close(done)
								}
							}()

							select {
							case partition.backlog <- backlogItem:
								atomic.SwapInt32(&pushed, 1)
							case <-done:
							}
						},
					)
				}

				wg.Wait()

				if pushed == 1 {
					break
				}

				continue
			}

			if inactivatedPartition := inactivatedPartitions[itemPartitionNum]; inactivatedPartition != nil {
				inactivatedPartitions[itemPartitionNum] = nil

				inactivatedPartition.first = backlogItem.item

				go r.doWorker(inactivatedPartition)

				break
			}

			select {
			case partitions[itemPartitionNum].backlog <- backlogItem:
				break pushLoop
			default:
			}

			select {
			case partitions[itemPartitionNum].backlog <- backlogItem:
				break pushLoop
			case inactivatedPartition := <-r.partitionInactivated:
				inactivatedPartitions[inactivatedPartition.num] = inactivatedPartition
			}
		}
	}
}

func sched[TResult any](in <-chan Job[TResult], maxPartitions MaxDOP, idleTimeout time.Duration, outBufSize int, voidResult bool) chan TResult {
	maxPartitionsInt := maxPartitions.MustEffective()

	if idleTimeout < 1 {
		panic(fmt.Errorf("invalid idle timeout '%s', can't be less than '%s'", idleTimeout, time.Duration(1)))
	}

	ctrl := &jobCtrl[TResult]{
		done:                 make(chan struct{}),
		in:                   in,
		partitionInactivated: make(chan *jobPartition[TResult], maxPartitionsInt),
		voidResult:           voidResult,
		workerIdleTimeout:    idleTimeout,
	}

	if ctrl.voidResult {
		ctrl.resultsCh = make(chan TResult)
	} else {
		ctrl.resultsCh = make(chan TResult, outBufSize)
	}

	for i := range maxPartitionsInt {
		partition := &jobPartition[TResult]{
			backlog: make(chan *backlogJob[TResult]),
			idle:    time.NewTimer(idleTimeout),
			num:     i,
		}

		partition.idle.Stop()

		ctrl.partitionInactivated <- partition
	}

	go ctrl.coordinatorLoop()

	return ctrl.resultsCh
}
