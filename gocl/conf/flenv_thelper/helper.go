package flenv_thelper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
)

func Populate[
	T any,
	TConf interface {
		flenv.Conf
		flenv.FlagSetBinder
		*T
	}](t *testing.T, conf TConf) {
	t.Helper()

	err := flenv.Populate(conf)
	require.NoError(t, err, "unexpected error while populating configuration structure '%s'", stringsext.TypeStringFor[TConf]())
}
