package contextext

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

// nolint:containedctx
type appContexts struct {
	shutdown         context.Context
	shutdownSig      context.CancelFunc
	gracefulEnd      context.Context
	gracefulEndSig   context.CancelCauseFunc
	gracefulEndCause error
}

func (v *appContexts) doGracefulEndSig() { v.gracefulEndSig(v.gracefulEndCause) }

func (v *appContexts) makeGracefulEnd(p context.Context) context.Context {
	if p == nil {
		p = context.Background()
	}

	if v == nil {
		return p
	}

	p = context.WithoutCancel(p)

	n, sig := context.WithCancelCause(p)

	context.AfterFunc(v.gracefulEnd, func() { sig(v.gracefulEndCause) })

	return n
}

const DefaultAppGracefulEndAfter = timeout.Timeout(time.Second * 15)

var appShutdownStorage atomic.Pointer[appContexts]

func TermSignals() []os.Signal {
	return []os.Signal{syscall.SIGABRT, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM}
}

// Term creates a new context that will be signaled on any of [TermSignals].
func Term() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), TermSignals()...)
}

// TermAfter creates a new context that will be signaled on any of [TermSignals] or after timeout elapsed.
func TermAfter(t timeout.Timeout) (context.Context, context.CancelFunc) {
	timeout, timeoutSig := t.Context(context.Background())
	term, termSig := signal.NotifyContext(timeout, TermSignals()...)

	return term, func() { termSig(); timeoutSig() }
}

// TrySetAppShutdownAsTerm sets a new current application shutdown context if it doesn't exist yet.
//
// If new context required then it will be created using [Term] func.
func TrySetAppShutdownAsTerm() (context.Context, context.CancelFunc) {
	curr := appShutdownStorage.Load()
	if curr != nil {
		return curr.shutdown, curr.shutdownSig
	}

	term, termCancel := Term()

	if setErr := SetAppShutdown(term, termCancel); errors.Is(setErr, errs.ErrAlreadyExists) {
		termCancel()

		curr = appShutdownStorage.Load()

		return curr.shutdown, curr.shutdownSig
	} else if setErr != nil {
		panic(setErr)
	}

	return term, termCancel
}

// MustSetAppShutdown sets current application shutdown context.
//
// Panic occurrs on attempt to replace existing context or when ctx is nil or cancel is nil.
func MustSetAppShutdown(ctx context.Context, cancel context.CancelFunc) {
	if err := SetAppShutdown(ctx, cancel); err != nil {
		panic(err)
	}
}

// SetAppShutdown sets current application shutdown context.
//
// Returns error on attempt to replace existing context or when ctx is nil or cancel is nil.
func SetAppShutdown(ctx context.Context, sig context.CancelFunc) error {
	return SetAppShutdown2(ctx, sig, DefaultAppGracefulEndAfter)
}

// SetAppShutdown2 sets current application shutdown context.
//
// Ruturns error:
//   - [errs.ErrAlreadyExists] on attempt to replace existing context
//   - [errs.ErrBadArgs] ctx is nil
//   - [errs.ErrBadArgs] cancel is nil
//   - [errs.ErrBadArgs] graceful timeout is not valid or [timeout.Infinite].
func SetAppShutdown2(ctx context.Context, sig context.CancelFunc, graceful timeout.Timeout) error {
	const errMsgPrologue = "unable to set application shutdown context"

	if ctx == nil {
		return fmt.Errorf(errMsgPrologue+"; %w; specified context is nil", errs.ErrBadArgs)
	} else if sig == nil {
		return fmt.Errorf(errMsgPrologue+"; %w; specified context cancel func is nil", errs.ErrBadArgs)
	} else if !graceful.IsValid() {
		return fmt.Errorf(errMsgPrologue+"; %w; specified graceful timeout not valid", errs.ErrBadArgs)
	} else if graceful.IsInfinite() {
		return fmt.Errorf(errMsgPrologue+"; %w; specified graceful timeout is infinite", errs.ErrBadArgs)
	}

	curr := appShutdownStorage.Load()
	if curr != nil {
		return fmt.Errorf(errMsgPrologue+"; %w", errs.ErrAlreadyExists)
	}

	newContexts := &appContexts{shutdown: ctx, shutdownSig: sig}

	switch {
	case graceful == 0:
		newContexts.gracefulEnd = ctx
		newContexts.gracefulEndSig = func(_ error) {}
	default:
		newContexts.gracefulEndCause = fmt.Errorf("app graceful shutdown timeout '%s' elapsed; %w", graceful, context.DeadlineExceeded)
		newContexts.gracefulEnd, newContexts.gracefulEndSig = context.WithCancelCause(context.WithoutCancel(ctx))
	}

	if swaped := appShutdownStorage.CompareAndSwap(curr, newContexts); !swaped {
		newContexts.doGracefulEndSig()

		return fmt.Errorf(errMsgPrologue+"; %w", errs.ErrAlreadyExists)
	}

	if graceful == 0 {
		return nil
	}

	go func() {
		<-newContexts.shutdown.Done()
		<-time.After(graceful.Dur())

		newContexts.doGracefulEndSig()
	}()

	return nil
}

// AppShutdown returns current application context.
//
// If context not set yet, returns [errs.ErrNotFound].
func AppShutdown() (context.Context, error) {
	curr := appShutdownStorage.Load()
	if curr == nil {
		return nil, fmt.Errorf("no application shutdown context; %w", errs.ErrNotFound)
	}

	return curr.shutdown, nil
}

// MustAppShutdown returns current application context (see also [AppShutdown]).
func MustAppShutdown() context.Context {
	res, err := AppShutdown()
	if err != nil {
		panic(err)
	}

	return res
}

func AppShutdownOrFb(fb context.Context) context.Context {
	curr := appShutdownStorage.Load()
	if curr == nil {
		return fb
	}

	return curr.shutdown
}

func AppShutdownOrBg() context.Context { return AppShutdownOrFb(context.Background()) }

func AppGracefulEnd(p context.Context) context.Context {
	return appShutdownStorage.Load().makeGracefulEnd(p)
}
