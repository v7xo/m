package findmnt_thelper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/findmnt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

func New(t *testing.T, ctx context.Context, opts ...findmnt.OptionFn) findmnt.Findmnt {
	t.Helper()

	res, err := findmnt.New(fopt.PrependFromContext(ctx, opts...))
	require.NoError(t, err)
	require.NotNil(t, res)

	return res
}
