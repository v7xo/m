package stdcase

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/conf/envext"
	"rn.m.vo7.io/vo7io/m/gocl/conf/startconf"
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
)

type setEnvironmentResultType struct {
	err         error
	confVariant startconf.ConfVariant
}

var (
	setEnvironmentMu     sync.RWMutex
	setEnvironmentResult *setEnvironmentResultType
)

type CommonProps struct {
	Desc string
	FailureProps
	OmitOnDebug bool
}

func (tc CommonProps) controlFlow(t *testing.T, parallel bool) {
	t.Helper()

	t.Helper()

	if envext.DebugTest() {
		switch {
		case tc.OmitOnDebug:
			t.Skipf("test case skipped because of test debug mode (see env. variable %s)", envext.DebugTestEnvVar)
		}
	} else if parallel {
		t.Parallel()
	}
}

// [PreLog] executes test preparation steps: environment setup, creation of test context, log.
func (tc CommonProps) PreLog(t *testing.T) (context.Context, logext.Logger) {
	t.Helper()

	ctx := tc.Pre(t)

	log, err := contextext.Log(ctx)
	require.NoError(t, err)

	return ctx, log
}

// [Pre] executes test preparation steps: sets environment, creates context, enriches created context with logger.
func (tc CommonProps) Pre(t *testing.T) context.Context {
	t.Helper()

	confVariant := tc.setEnvironment(t)
	tc.controlFlow(t, true)

	ctx := tc.context(t)
	logger := logext_thelper.
		Log(t).
		Tune().
		AddAttrs(
			attrname.Dict{
				attrname.GoTestName:  t.Name(),
				attrname.ConfVariant: confVariant.String(),
			}).
		Tuned()

	return context.WithValue(ctx, contextext.KeyLog, logger)
}

func (CommonProps) context(t *testing.T) context.Context {
	t.Helper()

	shutdownCtx, _ := contextext.TrySetAppShutdownAsTerm()
	require.NotNil(t, shutdownCtx, "unexpected nil context")

	res := context.WithValue(shutdownCtx, contextext.KeyGoTestName, t.Name())
	res, canc := context.WithCancel(res)

	t.Cleanup(canc)

	return res
}

func (CommonProps) setEnvironment(t *testing.T) (confVariant startconf.ConfVariant) {
	t.Helper()

	var err error

	defer func() {
		t.Helper()

		require.NoError(t, err, "unexpected error while setting test runtime environment")
	}()

	setEnvironmentMu.RLock()

	if setEnvironmentResult != nil {
		err, confVariant = setEnvironmentResult.err, setEnvironmentResult.confVariant

		setEnvironmentMu.RUnlock()

		return
	}

	setEnvironmentMu.RUnlock()

	setEnvironmentMu.Lock()

	defer func() { setEnvironmentMu.Unlock() }()

	if setEnvironmentResult != nil {
		err, confVariant = setEnvironmentResult.err, setEnvironmentResult.confVariant

		return
	}

	defer func() { setEnvironmentResult = &setEnvironmentResultType{err: err, confVariant: confVariant} }()

	confVariant = startconf.ConfVariantTestLocal

	if envext.RunningCIContainer() {
		confVariant = startconf.ConfVariantTest
	}

	startConf, err := startconf.FlenvRootConf()
	if err != nil {
		return
	}

	_, err = envext.SetEnvironment(confVariant, "", startConf.PrecedingFileNames(), startConf.SucceedingFileNames())

	return
}
