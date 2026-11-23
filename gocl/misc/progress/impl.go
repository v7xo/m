package progress

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
)

type (
	Progress[T any] interface {
		Report(val T)
	}

	of[T any] struct {
		_ noCopy

		handler          func(T)
		stop             context.Context
		stopSig          context.CancelFunc
		mu               *sync.Mutex
		awakeDeliverySig *atomic.Pointer[chan struct{}]
		stopDone         *atomic.Pointer[chan struct{}]
		buf              []T
		head             int
		tail             int
		cap              int
		avail            int
	}

	noCopy struct{}
)

var _ Progress[struct{}] = &of[struct{}]{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

func (s *of[T]) Stop(wait bool) {
	s.stopSig()

	if !wait {
		return
	}

	var waitOn, newWait chan struct{}

	for {
		curr := s.stopDone.Load()
		if curr != nil {
			waitOn = *curr

			if newWait != nil {
				close(newWait)
			}

			break
		}

		if newWait == nil {
			newWait = make(chan struct{})
		}

		if s.stopDone.CompareAndSwap(curr, &newWait) {
			waitOn = newWait

			break
		}
	}

	<-waitOn
}

func (s *of[T]) Report(val T) {
	s.mu.Lock()

	if s.buf == nil {
		s.mu.Unlock()

		return
	}

	s.buf[s.tail] = val
	s.tail = (s.tail + 1) % s.cap

	if s.avail == s.cap {
		s.head = (s.head + 1) % s.cap
	} else {
		s.avail++
	}

	s.mu.Unlock()

	if awake := s.awakeDeliverySig.Swap(nil); awake != nil {
		close(*awake)
	}
}

func (s *of[T]) consume() (T, bool) {
	var val, zero T

	s.mu.Lock()

	if s.buf == nil {
		s.mu.Unlock()

		return zero, false
	}

	if s.avail == 0 {
		s.mu.Unlock()

		return zero, false
	}

	s.avail--
	val = s.buf[s.head]
	s.buf[s.head] = zero
	s.head = (s.head + 1) % s.cap

	s.mu.Unlock()

	return val, true
}

func (s *of[T]) doDelivery() {
	defer func() {
		stopDone := make(chan struct{})

		for {
			curr := s.stopDone.Load()
			if curr != nil {
				close(stopDone)
				close(*curr)

				return
			}

			if s.stopDone.CompareAndSwap(curr, &stopDone) {
				close(stopDone)

				return
			}
		}
	}()

	awake := make(chan struct{})
	stopFn := func() {
		if awakePtr := s.awakeDeliverySig.Swap(nil); awakePtr != nil {
			close(*awakePtr)
		}

		s.mu.Lock()

		s.buf = nil

		s.mu.Unlock()
	}

	s.awakeDeliverySig.Store(&awake)

	for {
		select {
		case <-s.stop.Done():
			stopFn()

			return
		default:
		}

		select {
		case <-awake:
			awake = make(chan struct{})

			s.awakeDeliverySig.Store(&awake)

			for {
				val, ok := s.consume()
				if ok {
					s.handler(val)

					select {
					case <-s.stop.Done():
						stopFn()

						return
					default:
					}

					continue
				}

				break
			}
		case <-s.stop.Done():
			stopFn()

			return
		}
	}
}

func New[T any](opts ...OptionFn[T]) (_ *of[T], err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", typestring.TypeStringFor[of[T]](), err)
		}
	}()

	res := &of[T]{
		awakeDeliverySig: &atomic.Pointer[chan struct{}]{},
		cap:              DefaultCapacity,
		mu:               &sync.Mutex{},
		stopDone:         &atomic.Pointer[chan struct{}]{},
	}
	if err = fopt.Apply(res, opts...); err != nil {
		return nil, err
	}

	if res.cap < 1 {
		return nil, errors.New("buffer capacity can't be less than one")
	}

	if res.handler == nil {
		return nil, errors.New("no handler function specified (nil)")
	}

	if res.stop == nil {
		res.stop, res.stopSig = context.WithCancel(contextext.AppGracefulEnd(nil))
	} else {
		res.stop, res.stopSig = context.WithCancel(res.stop)
	}

	res.buf = make([]T, res.cap)

	go res.doDelivery()

	return res, nil
}
