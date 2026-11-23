package uuid_thelper

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	cluuid "rn.m.vo7.io/vo7io/m/gocl/misc/uuidext"
)

func Rand8BHex(t *testing.T) string {
	t.Helper()

	res, err := cluuid.Rand8BHex()
	require.NoError(t, err, "unexpected error while generating 8-bytes random value (UUID based)")
	require.Len(t, res, 16) //nolint:mnd,revive

	return res
}

func Rand4BHex(t *testing.T) string {
	t.Helper()

	res, err := cluuid.Rand4BHex()
	require.NoError(t, err, "unexpected error while generating 4-bytes random value (UUID based)")
	require.Len(t, res, 8) //nolint:mnd,revive

	return res
}

func Rand(t *testing.T) uuid.UUID {
	t.Helper()

	res, err := uuid.NewRandom()
	require.NoError(t, err, "unexpected error while generating UUID")

	return res
}
