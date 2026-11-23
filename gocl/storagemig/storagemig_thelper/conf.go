package storagemig_thelper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/storagemig"
)

func FlenvRootConf(t *testing.T) storagemig.Conf {
	t.Helper()

	cfg, err := storagemig.FlenvRootConf()
	require.NoError(t, err, "unexpected error while getting default storage migration configuration")

	return cfg
}
