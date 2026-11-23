package appstarter

//nolint:gci
import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/spf13/cobra"

	"rn.m.vo7.io/vo7io/m/gocl/appcmd/fluentcmd"
	"rn.m.vo7.io/vo7io/m/gocl/appmeta/info"
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logfactory"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

// Starter interface defines application starter methods.
type Starter interface {
	// ExecAndExit executes application and exits process.
	ExecAndExit()
	// Exec executes application.
	//
	// Context ctx can be nil.
	// If context is nil then [contextext.SetAppShutdown] will be called at beginning of execution.
	// Otherwise specified context will be propagated to execution.
	Exec(ctx context.Context, log logext.Logger) error
}

type starter struct {
	rootCobraCmd *cobra.Command
	execCalled   *atomic.Bool
	runTimestamp time.Time
	cmdMap       map[*cobra.Command]fluentcmd.Cmd
}

func (s *starter) ExecAndExit() {
	if err := s.Exec(nil, nil); err != nil {
		os.Exit(errs.FailExitCode)
	}

	os.Exit(0)
}

func (s *starter) Exec(ctx context.Context, log logext.Logger) error {
	if !s.execCalled.CompareAndSwap(false, true) {
		return errs.ErrAppAlreadyStarted
	}

	s.runTimestamp = time.Now().UTC()

	ctx, log, err := s.log(ctx, log)
	if err != nil {
		err = fmt.Errorf("failed to retrieve logger; %w", err)

		fmt.Fprintln(os.Stderr, "[ERR] "+err.Error())

		return err
	}

	const (
		stopSigSentYes = int32(1)
		stopSigSentNo  = int32(0)
		failExitCode   = errs.FailExitCode
	)

	var (
		stopCtx     context.Context
		stopSig     context.CancelFunc
		stopSigSent = stopSigSentNo
	)

	if ctx == nil {
		stopCtx, stopSig = contextext.Term()

		if setErr := contextext.SetAppShutdown(stopCtx, stopSig); setErr != nil {
			stopSig()

			return setErr
		}
	} else {
		stopCtx, stopSig = context.WithCancel(ctx)
	}

	sendStopSignal := func() {
		atomic.StoreInt32(&stopSigSent, stopSigSentYes)

		stopSig()
	}

	defer sendStopSignal()

	go func() {
		<-stopCtx.Done()

		if atomic.LoadInt32(&stopSigSent) == stopSigSentNo {
			switch {
			case log == nil || log.IsNop():
				fmt.Fprintf(os.Stderr, "[INFO] shutdown signal received")
			default:
				log.
					Info(stopCtx).
					Bool(attrname.AppQuitSigReceived, true).
					Msg("shutdown signal received")
			}
		}
	}()

	func() {
		defer func() {
			if err = errs.Join(errs.PanicErr(recover()), err); err != nil && !errors.Is(err, errs.ErrAppCommandFailure) {
				err = errs.AppCommandFailureJoin(err, !log.IsNop() && log.LogLevel().IsTraceOrDebug())
			}
		}()

		err = s.rootCobraCmd.ExecuteContext(stopCtx)
	}()

	if err != nil {
		switch {
		case !errors.Is(err, errs.ErrLogged):
			if log == nil || log.IsNop() {
				fmt.Fprintf(os.Stderr, "[ERR] %s\n", err)

				break
			}

			log.Err(stopCtx, err).Send()
		}
	}

	return err
}

func (s *starter) log(ctx context.Context, log logext.Logger) (context.Context, logext.Logger, error) {
	if log != nil {
		return ctx, log, nil
	}

	if ctx != nil {
		ctxLog, exists, ctxLogErr := contextext.LogExists(ctx)
		if ctxLogErr != nil || exists {
			return ctx, ctxLog, ctxLogErr
		}
	}

	log, err := logfactory.Default()(nil)
	if err != nil {
		return ctx, nil, err
	}

	if ctx != nil {
		ctx = logext.Context(ctx, log)
	}

	return ctx, log, nil
}

func (s *starter) cmdCalledAs(c *cobra.Command) string {
	n := c.CalledAs()
	if n == "" {
		n = c.Name()
	}

	if c.HasParent() {
		pn := s.cmdCalledAs(c.Parent())

		if pn != "" {
			return pn + " " + n
		}
	}

	return n
}

func New(opts ...OptionFn) (_ Starter, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[starter](), err)
		}
	}()

	impl := &starter{
		rootCobraCmd: &cobra.Command{
			Args:          cobra.NoArgs,
			Long:          info.AboutLong(),
			Short:         info.AboutShort(),
			SilenceErrors: false,
			SilenceUsage:  true,
			Use:           "",
			Version:       info.VersionLong(),
		},
		cmdMap:     map[*cobra.Command]fluentcmd.Cmd{},
		execCalled: &atomic.Bool{},
	}
	if err = fopt.Apply(impl, opts...); err != nil {
		return nil, err
	}

	return impl, nil
}

func MustNew(opts ...OptionFn) Starter {
	res, err := New(opts...)
	if err != nil {
		panic(err)
	}

	return res
}
