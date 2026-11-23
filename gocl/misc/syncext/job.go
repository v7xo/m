package syncext

import "sync/atomic"

type (
	Job[TResult any] interface {
		Do() TResult
		PartitionKey() JobPartitionKey
	}

	userJob[TResult any] struct {
		f func() TResult
		k JobPartitionKey
	}

	backlogJob[TResult any] struct {
		item  Job[TResult]
		taken atomic.Int32
	}
)

var _ Job[struct{}] = userJob[struct{}]{}

func (v userJob[TResult]) Do() TResult { return v.f() }

func (v userJob[TResult]) PartitionKey() JobPartitionKey { return v.k }

func NewJob[TResult any](fn func() TResult, k JobPartitionKey) Job[TResult] {
	return userJob[TResult]{f: fn, k: k}
}

func NewJobVoid(fn func(), k JobPartitionKey) Job[struct{}] {
	return userJob[struct{}]{
		f: func() struct{} {
			fn()

			return struct{}{}
		},
		k: k,
	}
}
