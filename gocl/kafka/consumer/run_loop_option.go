package consumer

import (
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type runLoopOpts struct {
	procFn                    procFn
	topic                     string
	breakOnFetchNextTimeout   time.Duration
	breakOnLoopRunTimeReached time.Duration
	noCommits                 bool
}

type RunLoopOptionFn func(*runLoopOpts) error

func (o RunLoopOptionFn) ProcFn(val procFn) RunLoopOptionFn {
	return fopt.Chain(o, func(t *runLoopOpts) { t.procFn = val })
}

func (o RunLoopOptionFn) NoCommits(val bool) RunLoopOptionFn {
	return fopt.Chain(o, func(t *runLoopOpts) { t.noCommits = val })
}

func (o RunLoopOptionFn) BreakOnFetchNextTimeout(val time.Duration) RunLoopOptionFn {
	return fopt.Chain(o, func(t *runLoopOpts) { t.breakOnFetchNextTimeout = val })
}

func (o RunLoopOptionFn) BreakOnLoopRunTimeReached(val time.Duration) RunLoopOptionFn {
	return fopt.Chain(o, func(t *runLoopOpts) { t.breakOnLoopRunTimeReached = val })
}

func (o RunLoopOptionFn) Topic(val string) RunLoopOptionFn {
	return fopt.Chain(o, func(t *runLoopOpts) { t.topic = val })
}

func RunLoopOption() RunLoopOptionFn { return fopt.Nop[runLoopOpts] }
