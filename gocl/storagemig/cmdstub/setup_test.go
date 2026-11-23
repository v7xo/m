//go:build test_unit

package cmdstub_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/app/appstarter"
	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/cmdstub"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_Setup(t *testing.T) {
	ctx, log := thelper.PreLog(t)

	appOpts := appstarter.Option().CmdFn(cmdstub.SetupWithSourcesFS(os.DirFS(os.TempDir())))

	app, err := appstarter.New(appOpts)
	require.NoError(t, err)
	require.NotNil(t, app)

	os.Args = []string{os.Args[0]}
	t.Logf("\n\n\n****** test command '%s' ******", os.Args)
	err = app.Exec(ctx, log)
	require.NoError(t, err)

	os.Args = []string{os.Args[0], "migrate", "status"}
	t.Logf("\n\n\n****** test command '%s' ******", os.Args)
	err = app.Exec(ctx, log)
	require.ErrorIs(t, err, errs.ErrAppCommandFailure)

	os.Args = []string{os.Args[0], "migrate", "status", "--" + flenv.FlagPrintEnvironment}
	t.Logf("\n\n\n****** test command '%s' ******", os.Args)
	err = app.Exec(ctx, log)
	require.NoError(t, err)

	os.Args = []string{os.Args[0], "migrate", "status", "--help"}
	t.Logf("\n\n\n****** test command '%s' ******", os.Args)
	err = app.Exec(ctx, log)
	require.NoError(t, err)
}
