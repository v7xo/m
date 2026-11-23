//go:build test_unit

package ioext_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	clio "rn.m.vo7.io/vo7io/m/gocl/misc/ioext"
	clhlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func TestCopyMaxN_Happy(t *testing.T) {
	clhlp.Pre(t)

	bufSrcReader := bytes.NewBuffer([]byte("hello"))
	bufDstWriter := bytes.NewBuffer(nil)

	_, err := clio.CopyMaxN(bufDstWriter, bufSrcReader, 5)
	require.NoError(t, err)

	require.Equal(t, "hello", string(bufDstWriter.Bytes()))
}

func Test_CopyMaxN_ShortWrite(t *testing.T) {
	clhlp.Pre(t)

	bufSrcReader := bytes.NewBuffer([]byte("hello"))
	bufDstWriter := bytes.NewBuffer(nil)

	_, err := clio.CopyMaxN(bufDstWriter, bufSrcReader, 4)
	require.ErrorIs(t, err, io.ErrShortWrite)
}
