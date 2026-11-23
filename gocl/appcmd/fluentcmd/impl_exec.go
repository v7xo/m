package fluentcmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/appmeta/buildinfo"
	"rn.m.vo7.io/vo7io/m/gocl/conf/envext"
	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/conf/startconf"
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

func (s *impl[T]) Exec(ctx context.Context, p ExecParams) (err error) {
	if s.printEnvironmentAndExit || s.printUserProfileConfSubdirAndExit {
		return s.printAndExit()
	}

	// TODO: fields below should be in dedicated package.
	buildinfo.CmdName = p.CalledAs

	registerAppCmdInfo()

	defer func() { registerAppCmdExitStatus(err) }()

	logAttrs := s.defaultLogAttrs(p)

	log, err := s.defaultLog(ctx, logAttrs)
	if err != nil {
		return err
	}

	stdinCloser := s.stdinCloser(ctx, log)

	defer stdinCloser()

	confVariant, err := s.prepareStartConf(logAttrs)
	if err != nil {
		return err
	}

	log = log.Tune().AddAttrs(logAttrs).Tuned()

	if err = s.prepareConf(confVariant, false); err != nil {
		return err
	}

	if ctx, log, err = s.overrideWithCmdLog(ctx, log, logAttrs); err != nil {
		return err
	}

	if err = s.execCore(ctx, log); err != nil {
		return err
	}

	return nil
}

func (s *impl[T]) execCore(ctx context.Context, log logext.Logger) (err error) {
	var diagEndpointStop func()

	defer func() {
		err = errs.Join(errs.PanicErr(recover()), err)
		if err != nil {
			err = errs.AppCommandFailureJoin(err, !log.IsNop() && log.LogLevel().IsTraceOrDebug())

			log.Err(ctx, err).Send()

			err = errs.Join(errs.ErrLogged, err)
		} else {
			log.Trace(ctx).Msg("command completed without error; exit")
		}

		s.doExitDelay(ctx, log)

		if diagEndpointStop != nil {
			diagEndpointStop()
		}
	}()

	if diagEndpointStop, err = s.diagnosticsEndpointStart(log); err != nil {
		return err
	}

	s.emitStartMsg(ctx, log)

	if err = s.execBefore(ctx, log); err != nil {
		return err
	}

	return s.userExecFn(ctx, s.conf, log)
}

func (s *impl[T]) execBefore(ctx context.Context, log logext.Logger) error {
	if s.userBeforeExecFn == nil {
		return nil
	}

	if err := s.userBeforeExecFn(ctx, s.conf, log); err != nil {
		return fmt.Errorf("failed to execute 'before actions' for command '%s'; %w", s, err)
	}

	return nil
}

func (s *impl[T]) emitStartMsg(ctx context.Context, log logext.Logger) {
	if log.IsNop() {
		return
	}

	argsString := placeholder.Omitted

	if log.LogLevel().IsTraceOrDebug() {
		argsString = fmt.Sprintf("%q", os.Args)
	}

	log.Info(ctx).Msgf("start '%s' (log level: %s, args: %s) ...", s, log.LogLevel().Name(), argsString)
}

func (s *impl[T]) stdinCloser(ctx context.Context, log logext.Logger) func() {
	if !s.cli {
		return func() {}
	}

	ctx, sig := context.WithCancel(ctx)

	context.AfterFunc(
		ctx,
		func() {
			if closeErr := os.Stdin.Close(); closeErr != nil {
				log.Err(ctx, fmt.Errorf("failed to close stdin; %w", closeErr)).Send()

				return
			}

			log.Trace(ctx).Msg("stdin closed")
		})

	return func() { sig() }
}

func (s *impl[T]) prepareStartConf(logAttrs attrname.Dict) (_ startconf.ConfVariant, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to prepare start configuration for command '%s'; %w", s, err)
		}
	}()

	const unspecifiedConfVariant = startconf.ConfVariantUnspecified

	flset, err := s.GenerateFlagSet()
	if err != nil {
		return unspecifiedConfVariant, err
	}

	//// ensure, that command start flags parsed.
	//// current version assumes that flags parsed using Cobra command before the call to this method.
	//if !flset.Parsed() {
	//	return unspecifiedConfVariant, errors.New("unexpected condition; flag set not parsed yet")
	//}

	// populate configuration with values from current environment.
	// for ex., ENV_A=1 ENV_B=2 <command> ...
	// ENV_A and ENV_B populate values of matching configuration fields.
	// consider, flag set includes bindings to start configuration fields.
	if err = flenv.PopulateFlagSetFromEnvironment(flset); err != nil {
		return unspecifiedConfVariant, err
	}

	if err = s.prepareConfForUse(&s.startConf); err != nil {
		return unspecifiedConfVariant, err
	}

	confVariant := s.startConf.ConfVariant()

	switch {
	case confVariant.IsAuto():
		confVariant, err = startconf.DetectConfVariant()
		if err != nil {
			return unspecifiedConfVariant, err
		}
	}

	if logAttrs == nil {
		return confVariant, nil
	}

	logAttrs[attrname.ConfVariant] = confVariant.String()

	if !s.startConf.CmdTimeout().IsInfinite() {
		logAttrs[attrname.AppCmdRunTimeoutMs] = s.startConf.CmdTimeout().Milliseconds()
	}

	return confVariant, nil
}

func (s *impl[T]) defaultLogAttrs(p ExecParams) attrname.Dict {
	dict := make(attrname.Dict, 8) // nolint:mnd,revive

	dict[attrname.AppCmdRunID] = p.RunID
	dict[attrname.AppCmdRunTimestamp] = p.RunTimestamp.Format(time.RFC3339)
	dict[attrname.AppCmdRunTimestampUnix] = p.RunTimestamp.Unix()

	return dict
}

func (s *impl[T]) overrideWithCmdLog(ctx context.Context, log logext.Logger, logAttrs attrname.Dict) (context.Context, logext.Logger, error) {
	if s.logFactoryFn == nil || !s.confSupplied {
		return ctx, log, nil
	}

	newLog, err := s.logFactoryFn(s.cmdLogConfGetter())
	if err != nil {
		return ctx, log, fmt.Errorf("failed to create a logger for command '%s'; %w", s, err)
	}

	if newLog == nil {
		return ctx, log, nil
	}

	newLog = newLog.Tune().AddAttrs(logAttrs).Tuned()

	return contextext.WithLog(ctx, newLog), newLog, nil
}

func (s *impl[T]) defaultLog(ctx context.Context, attrs attrname.Dict) (logext.Logger, error) {
	log, err := contextext.Log(ctx)
	if err != nil {
		return logext.NopLogger, err
	}

	if log == nil {
		log = logext.NopLogger
	}

	if !log.IsNop() && len(attrs) != 0 {
		log = log.Tune().AddAttrs(attrs).Tuned()
	}

	return log, nil
}

func (s *impl[T]) doExitDelay(ctx context.Context, log logext.Logger) {
	exitDelay := s.startConf.CmdExitDelay()
	if exitDelay < 1 {
		return
	}

	if log.IsNop() {
		fmt.Fprintf(os.Stderr, "[INFO] delay '%s' before exit ...", exitDelay)
	} else {
		log.Info(ctx).Msgf("delay '%s' before exit ...", exitDelay)
	}

	select {
	case <-ctx.Done():
	case <-time.After(exitDelay):
	}
}

func (s *impl[T]) prepareConf(confVariant startconf.ConfVariant, skipPrepareForUse bool) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to prepare configuration for command '%s'; %w", s, err)
		}
	}()

	if confVariant.IsAuto() {
		return fmt.Errorf("invalid configuration variant '%s'", confVariant)
	}

	flset, err := s.GenerateFlagSet()
	if err != nil {
		return err
	}

	var (
		envChange     bool
		useConfSubdir string
	)

	if s.startConf.UserProfileConfDir() == "" {
		useConfSubdir = s.defaultUserProfileConfSubdir()
	}

	if envChange, err = envext.SetEnvironment2(confVariant, useConfSubdir, s.startConf); err != nil {
		return err
	}

	switch {
	case envChange:
		// populate configuration with values from environment (as latter modified, see above).
		if err = flenv.PopulateFlagSetFromEnvironment(flset); err != nil {
			return err
		}
	}

	if skipPrepareForUse {
		return nil
	}

	if err = s.prepareConfForUse(s.conf); err != nil {
		return err
	}

	return nil
}

func (*impl[TConf]) prepareConfForUse(conf flenv.Conf) error {
	if err := conf.Validate(validation.DefaultValidator()); err != nil {
		return fmt.Errorf("configuration '%s' not valid; %w", stringsext.TypeStringOfVal(conf), err)
	}

	be, ok := conf.(flenv.UsePreparer)
	if !ok {
		return nil
	}

	if err := be.PrepareUse(); err != nil {
		return fmt.Errorf("failed to prepare configuration '%s' for use; %w", stringsext.TypeStringOfVal(conf), err)
	}

	return nil
}

func (s *impl[TConf]) defaultUserProfileConfSubdir() string {
	var (
		parts     = []string{}
		cmd   Cmd = s
	)

	for ; cmd != nil; cmd = cmd.Parent() {
		parts = append(parts, cmd.Use())
	}

	parts = append(parts, filepath.Base(os.Args[0]))

	slices.Reverse(parts)

	return filepath.Join(parts...)
}

func (s *impl[T]) printAndExit() error {
	confVariant, err := s.prepareStartConf(nil)
	if err != nil {
		return err
	}

	if err = s.prepareConf(confVariant, true); err != nil {
		return err
	}

	flset, err := s.GenerateFlagSet()
	if err != nil {
		return err
	}

	if s.printEnvironmentAndExit {
		flenv.PrintFlagEnvironment(flset)
	}

	if s.printUserProfileConfSubdirAndExit {
		useConfSubdir := ""
		if s.startConf.UserProfileConfDir() == "" {
			useConfSubdir = s.defaultUserProfileConfSubdir()
		}

		dir, preceding, succeeding, err := envext.UserProfileConfDir(useConfSubdir, s.startConf)
		if err != nil {
			return err
		}

		if len(preceding) == 0 && len(succeeding) == 0 {
			fmt.Println(dir) //nolint

			return nil
		}

		for i := range len(preceding) {
			fmt.Println(filepath.Join(dir, preceding[i])) //nolint
		}

		for i := range len(succeeding) {
			fmt.Println(filepath.Join(dir, succeeding[i])) //nolint
		}
	}

	return nil
}
