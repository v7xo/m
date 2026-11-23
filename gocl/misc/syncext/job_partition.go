package syncext

import (
	"time"
)

type jobPartition[TResult any] struct {
	first   Job[TResult]
	idle    *time.Timer
	backlog chan *backlogJob[TResult]
	num     int
}
