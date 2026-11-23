//go:build test_unit

package lsblk_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/lsblk"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_List_Happy(t *testing.T) {
	ctx, _ := hlp.PreLog(t)
	s := HelperNew(t, ctx)

	_, err := s.List(ctx, lsblk.Filter{})
	require.NoError(t, err)
}

func HelperNew(t *testing.T, ctx context.Context, opts ...lsblk.OptionFn) lsblk.Lsblk {
	t.Helper()

	res, err := lsblk.New(fopt.Pack(lsblk.Option().Log(logext_thelper.FromContext(t, ctx)), opts...))
	require.NoError(t, err)
	require.NotNil(t, res)

	return res
}
