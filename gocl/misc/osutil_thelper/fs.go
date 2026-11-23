package osutil_thelper

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil"
	"rn.m.vo7.io/vo7io/m/gocl/misc/uuidext"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func MakeFile(t *testing.T, n string, size int64, rewrite bool) *os.File {
	t.Helper()

	res, err := osutil.OpenFileSize(t.Context(), n, size, rewrite)
	require.NoError(t, err, "unexpected error while generating file '%s' (size %d)", n, size)
	require.NotNil(t, res)

	thelper.Cleanup(
		t,
		func() error {
			err := os.Remove(n)
			if err != nil {
				return err
			}

			logext_thelper.Log(t).Info(t.Context()).Msgf("temporary file '%s' removed", n)

			return nil
		},
		"error while removing temporary file '%s'", n,
	)

	return res
}

func MkdirAll(t *testing.T, n string, p os.FileMode) string {
	t.Helper()

	err := os.MkdirAll(n, p)
	require.NoError(t, err)

	return n
}

func MkdirTemp(t *testing.T, pattern string, cleanup osutil.Remove) string {
	t.Helper()

	if pattern == "" {
		pattern = "cl_tests_dir_*"
	}

	tmpDir, err := osutil.MkdirTemp(pattern)
	require.NoError(t, err, "unexpected error while making temporary directory")

	if cleanup.Yes() {
		thelper.Cleanup(
			t,
			func() error {
				if cleanup.All() {
					return os.RemoveAll(tmpDir)
				}

				return os.Remove(tmpDir)
			},
			"error while removing temporary directory '%s' (removal flags '%s')",
			tmpDir,
			cleanup,
		)
	}

	return tmpDir
}

func MakeFileTemp(t *testing.T, tmpDirPattern string, size int64) string {
	t.Helper()

	tmpDir := MkdirTemp(t, tmpDirPattern, osutil.RemoveYes)
	fileName := filepath.Join(tmpDir, uuidext.MustRand8BHex())

	MakeFile(t, fileName, size, false)

	return fileName
}
