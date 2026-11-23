package clock_thelper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/clock"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

func New(t *testing.T, ctx context.Context, opts ...clock.OptionFn) clock.Clock {
	t.Helper()

	res, err := clock.New(fopt.PrependFromContext(ctx, opts...))
	require.NoError(t, err, "unexpected error while creating instance of clock service")
	require.NotNil(t, res, "unexpected nil as a result of clock service creation")

	return res
}
