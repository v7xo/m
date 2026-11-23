package appmodule

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

// Run runs all the specified modules each in a separate goroutine and blocks until all of them are finished.
// It's not expected that any module will finish before ctx signaled.
func Run(ctx context.Context, log logext.Logger, mods ...Module) error {
	modsCopy := make([]Module, len(mods))

	for i := range mods {
		mod := mods[i]

		if mod == nil {
			return fmt.Errorf("specified module collection contains nil-module at %d position", i)
		}

		modsCopy[i] = mod
	}

	type modFinishedArgs struct {
		err     error
		modName string
	}

	modFinishedCh := make(chan modFinishedArgs, len(modsCopy))
	wg := sync.WaitGroup{}

	modRunBaseCtx := context.WithoutCancel(ctx)
	modRunBaseCtx, finishSig := context.WithCancelCause(modRunBaseCtx)
	finishSigBaseCause := errors.New("module must be finished")

	wg.Add(len(modsCopy))

	for i := range modsCopy {
		mod := modsCopy[i]
		modName := mod.Name()
		modRunCtx := contextext.WithAttrDictPairs(modRunBaseCtx, attrname.ModuleName, modName)

		go func() {
			defer wg.Done()

			log.Info(modRunCtx).Msgf("module '%s' starting ...", modName)

			var modRunErr error

			func() {
				defer func() {
					if panicErr := errs.PanicErr(recover()); panicErr != nil {
						modRunErr = errs.Join(modRunErr, fmt.Errorf("panic occurred while running module '%s'; %w", modName, panicErr))
					}
				}()

				modRunErr = mod.Run(modRunCtx)
			}()

			if modRunErr == nil {
				log.Info(modRunCtx).Msgf("module '%s' finished", modName)
			} else if errors.Is(modRunErr, finishSigBaseCause) ||
				(errors.Is(modRunErr, context.Canceled) && errors.Is(context.Cause(modRunCtx), finishSigBaseCause)) {
				log.Info(modRunCtx).Msgf("module '%s' finished; module stop caused by following reason; %s", modName, modRunErr.Error())

				modRunErr = nil
			} else {
				log.Err(modRunCtx, modRunErr).Msgf("module '%s' finished with error", modName)
			}

			modFinishedCh <- modFinishedArgs{modName: modName, err: modRunErr}
		}()
	}

	finishSigCause := error(nil)
	modsErrs := make([]error, 0, len(modsCopy))

	select {
	case <-ctx.Done():
		log.Info(ctx).Msg("app modules run control: outter context signal received")

		finishSigCause = fmt.Errorf("%w; outter context signal received", finishSigBaseCause)
	case finishedArgs := <-modFinishedCh:
		finishSigCause = fmt.Errorf("%w; unexpected stop of module '%s' caused other module(s) to stop too", finishSigBaseCause, finishedArgs.modName)

		if finishedArgs.err == nil {
			modsErrs = append(modsErrs, fmt.Errorf("%w; module '%s' unexpectedly finished, module not reported any error", errs.ErrUnexpectedStop, finishedArgs.modName))

			break
		}

		modsErrs = append(modsErrs, fmt.Errorf("%w; module '%s' unexpectedly finished with error; %w", errs.ErrUnexpectedStop, finishedArgs.modName, finishedArgs.err))
	}

	finishSig(finishSigCause)
	wg.Wait()

	for len(modFinishedCh) != 0 {
		finishedArgs := <-modFinishedCh

		modsErrs = append(modsErrs, finishedArgs.err)
	}

	return errs.Join(modsErrs...)
}
