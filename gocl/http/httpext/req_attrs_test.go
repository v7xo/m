//go:build test_unit

package httpext

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_ReplaceNumberWithPlaceholder(t *testing.T) {
	thelper.Pre(t)

	input := "/1/12/123/aaa/5896655/abcd110"
	output := numberRegexp.ReplaceAllString(input, NumberPlaceholder)

	require.Equal(t, fmt.Sprintf("/1/%[1]s/%[1]s/aaa/%[1]s/abcd%[1]s", NumberPlaceholder), output)
}

func Test_ReplaceUUIDWithPlaceholder(t *testing.T) {
	thelper.Pre(t)

	input := "/a/14585/da43f5d5-b297-4e55-af72-f832a2d4664f/12/123/da43f5d5-b297-4e55-af72-f832a2d4664f/f832a2d4664f"
	output := uuidRegexp.ReplaceAllString(input, UUIDPlaceholder)

	require.Equal(t, fmt.Sprintf("/a/14585/%[1]s/12/123/%[1]s/f832a2d4664f", UUIDPlaceholder), output)
}
