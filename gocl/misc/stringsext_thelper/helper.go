package stringsext_thelper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
)

func KVOptsFromDictString(t *testing.T, d stringsext.KVDictString) (r stringsext.KVOpts) {
	t.Helper()

	err := r.FromKVDict(d)
	require.NoError(t, err)

	return r
}
