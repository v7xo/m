package readiness

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
)

type simple struct {
	collectors            []Collector
	collectorsMu          sync.Mutex
	readinessConsolidated atomic.Bool
}

var defaultRegistry atomic.Pointer[simple]

func (s *simple) Register(c Collector) error {
	if c == nil {
		return errors.New("collector arg 'c' can't be nil")
	} else if _, ok := c.(*simple); ok {
		return errors.New("invalid collector")
	}

	s.collectorsMu.Lock()

	s.collectors = append(s.collectors, c)

	s.collectorsMu.Unlock()

	return nil
}

func (s *simple) CollectReadinessConsolidated(_ context.Context) (bool, error) {
	return s.readinessConsolidated.Load(), nil
}

func (s *simple) start(stop context.Context) {
	ticker := time.NewTicker(time.Second)

	go func() {
	mainLoop:
		for {
			select {
			case <-stop.Done():
				s.readinessConsolidated.Store(false)

				return
			case <-ticker.C:
			}

			var colls []Collector

			s.collectorsMu.Lock()

			colls = make([]Collector, len(s.collectors))
			copy(colls, s.collectors)

			s.collectorsMu.Unlock()

			if len(colls) == 0 {
				s.readinessConsolidated.Store(false)

				continue mainLoop
			}

			for _, coll := range colls {
				ready, collErr := coll.CollectReadinessConsolidated(stop)
				if collErr != nil {
					// TODO: log collector error.
					s.readinessConsolidated.Store(false)

					continue mainLoop
				} else if !ready {
					s.readinessConsolidated.Store(false)

					continue mainLoop
				}
			}

			s.readinessConsolidated.Store(true)
		}
	}()
}

func Registry() *simple {
	newReg := &simple{}

	if defaultRegistry.CompareAndSwap(nil, newReg) {
		appShutdownCtx, _ := contextext.Term()

		newReg.start(appShutdownCtx)

		return newReg
	}

	return defaultRegistry.Load()
}

type alwaysReady struct{}

func (*alwaysReady) CollectReadinessConsolidated(_ context.Context) (bool, error) { return true, nil }

func AlwaysReady() Collector { return &alwaysReady{} }
