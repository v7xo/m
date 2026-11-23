package textempl_thelper

import (
	"bytes"
	iofs "io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/ioext_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/textempl"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func ExecuteFile(t *testing.T, fs iofs.FS, n string, data any) *bytes.Buffer {
	t.Helper()

	text, _ := ioext_thelper.ReadAllTextFile(t, fs, n, false)
	templ, err := textempl.Parse(text)
	require.NoError(t, err, "unexpected error while parsing Go text template from specified file '%s' (fs '%v (%[1]T)') ", n, fs)

	buf := &bytes.Buffer{}

	err = templ.Execute(buf, data)
	require.NoError(t, err, "unexpected error while executing Go text template from specified file '%s' (fs '%v (%[1]T)') ", n, fs)

	return buf
}

func ExecuteFileToTemp(t *testing.T, fs iofs.FS, src string, data any) string {
	t.Helper()

	tmpDir := osutil_thelper.MkdirTemp(t, "v7xo-*", osutil.RemoveAll|osutil.RemoveYes)
	dst := filepath.Join(tmpDir, "go_text_templ_result")

	ExecuteFileToFile(t, fs, src, dst, data)

	return dst
}

func ExecuteFileToFile(t *testing.T, fs iofs.FS, src, dst string, data any) {
	t.Helper()

	buf := ExecuteFile(t, fs, src, data)

	err := os.WriteFile(dst, buf.Bytes(), os.ModePerm) //nolint:gosec
	require.NoError(t, err, "unexpected error while writing result of Go text template execution")

	thelper.Cleanup(t, func() error { return os.Remove(dst) }, "error while removing temporary file '%s'", dst)
}
