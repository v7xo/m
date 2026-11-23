//go:build test_unit

package fsutil_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/fsutil"
	"rn.m.vo7.io/vo7io/m/gocl/misc/syncext"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func TestSHA256SumDir(t *testing.T) {
	ctx := thelper.Pre(t)

	const (
		dir             = "./test-data/sha256sum"
		dirExpectedHash = "d01fbf2bdcf6ddb8872b464b92107e1ec36c53da2144c3346c6810fb8ab6a352"
	)

	h, err := fsutil.DirSHA256SumHex(ctx, os.DirFS(dir), ".", syncext.MaxDOPAuto, nil)
	require.NoError(t, err, "unexpected error while computing SHA256 hash of directory '%s'", dir)
	require.Equal(t, dirExpectedHash, h, "actual hash of directory mismatches to expected one")
}
