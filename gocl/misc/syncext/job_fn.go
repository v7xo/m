package syncext

import (
	"time"
)

func SchedVoid(in <-chan Job[struct{}], partitions MaxDOP, idleTimeout time.Duration) <-chan struct{} {
	return sched(in, partitions, idleTimeout, 0, true)
}

func Sched[TResult any](in <-chan Job[TResult], partitions MaxDOP, idleTimeout time.Duration, outBufSize int) <-chan TResult {
	return sched(in, partitions, idleTimeout, outBufSize, false)
}

type SchedAndCollectParams struct {
	Partitions  MaxDOP
	IdleTimeout time.Duration
}

func SchedAndCollect[TResult any](in <-chan Job[TResult], p SchedAndCollectParams) <-chan []TResult {
	var collectedOutputs []TResult

	outputsCh := make(chan []TResult, 1)
	jobsOut := sched(in, p.Partitions, p.IdleTimeout, 0, false)

	go func() {
		for {
			output, ok := <-jobsOut
			if !ok {
				outputsCh <- collectedOutputs

				close(outputsCh)

				return
			}

			collectedOutputs = append(collectedOutputs, output)
		}
	}()

	return outputsCh
}
