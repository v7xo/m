package ioext_thelper

import (
	"errors"
	"io"
	iofs "io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/ioext"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func OpenFile(t *testing.T, fs iofs.FS, n string, ignoreNotExist bool) iofs.File {
	t.Helper()

	var (
		err  error
		file iofs.File
	)

	switch fs {
	case nil:
		file, err = os.Open(n)
	default:
		file, err = fs.Open(n)
	}

	if errors.Is(err, os.ErrNotExist) {
		switch {
		case ignoreNotExist:
			return nil
		}
	}

	require.NoError(t, err, "unexpected error while opening file; source (fs): %v (type: %[1]T); file: %s", fs, n)
	require.NotNil(t, file, "unexpected nil as a result of file opening; source (fs): %v (type: %[1]T); file: %s", fs, n)

	thelper.Cleanup(t, file.Close, "unexpected error while closing file; source (fs): %v (type: %[1]T); file: %s", fs, n)

	return file
}

func ReadAllTextFile(t *testing.T, fs iofs.FS, n string, ignoreNotExist bool) (string, bool) {
	t.Helper()

	file := OpenFile(t, fs, n, ignoreNotExist)
	if file == nil {
		return "", false
	}

	buf, err := io.ReadAll(file)
	require.NoError(t, err, "unexpected error while reading all content of file as text; source (fs): %v (type: %[1]T); file: %s", fs, n)

	return string(buf), true
}

func ReadAllFile(t *testing.T, fs iofs.FS, n string, ignoreNotExist bool) ([]byte, bool) {
	t.Helper()

	file := OpenFile(t, fs, n, ignoreNotExist)
	if file == nil {
		return nil, false
	}

	buf, err := io.ReadAll(file)
	require.NoError(t, err, "unexpected error while reading all content of file; source (fs): %v (type: %[1]T); file: %s", fs, n)

	return buf, true
}

func SHA256SumHexFile(t *testing.T, fs iofs.FS, n string) string {
	t.Helper()

	file := OpenFile(t, fs, n, false)

	res, err := ioext.SHA256SumHex(file)
	require.NoError(t, err)
	require.Len(t, res, 64)

	return res
}
