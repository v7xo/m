package pb_thelper

import (
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"rn.m.vo7.io/vo7io/m/gocl/misc/ioext_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
)

func RequireEq(t *testing.T, expected, actual proto.Message, msgAndArgs ...any) {
	t.Helper()

	if expected == nil {
		require.Nil(t, actual, "actual message not equals to nil expected message")
	} else {
		require.NotNil(t, actual, "actual message is nil whereas non-nil expected")
	}

	actualJSON, err := protojson.Marshal(actual)
	require.NoError(t, err, "unexpected error while marshaling actual Protobuf message to JSON")

	expectedJSON, err := protojson.Marshal(expected)
	require.NoError(t, err)
	require.NoError(t, err, "unexpected error while marshaling expected Protobuf message to JSON")

	require.JSONEq(t, string(expectedJSON), string(actualJSON), msgAndArgs...)
}

func WriteJSONPbToFile(t *testing.T, m proto.Message, n string) {
	t.Helper()

	opts := protojson.MarshalOptions{
		Multiline:         true,
		UseProtoNames:     true,
		EmitUnpopulated:   true,
		EmitDefaultValues: true,
	}

	buf, err := opts.Marshal(m)
	require.NoError(t, err, "unexpected error while marshaling JSON-protobuf to file '%s'", n)

	err = os.WriteFile(n, buf, 0o0664) //nolint
	require.NoError(t, err, "unexpected error while writing JSON-protobuf to file '%s'", n)
}

func ReadJSONPbFromFile[
	T any,
	TMessage interface {
		proto.Message
		*T
	}](
	t *testing.T,
	source fs.FS,
	path string,
	m TMessage,
) {
	t.Helper()

	file := ioext_thelper.OpenFile(t, source, path, false)
	buf, err := io.ReadAll(file)
	require.NoError(t, err, "unexpected error while reading content of file '%s' (file system '%s')", path, typestring.TypeStringOfVal(source))

	err = protojson.Unmarshal(buf, m)
	require.NoError(t, err, "unexpected error while unmarshalling JSON-protobuf content of file '%s'", path)
}
