//go:build test_unit

package stringsext_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

func TestTruncateDefault(t *testing.T) {
	hlp.PreLog(t)

	type testCase struct {
		stdcase.CommonProps

		input    string
		expected string
		maxLen   int
	}

	testCases := map[string]testCase{
		"case00": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: true,
			},

			input:    "0123456789",
			expected: "0123456789",
			maxLen:   10,
		},
		"case01/odd max len": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:    "0123456789",
			expected: "01[…]89",
			maxLen:   9,
		},
		"case02/even max len": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:    "0123456789",
			expected: "0[…]89",
			maxLen:   8,
		},
		"case03/only elipsis": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:    "0123456789",
			expected: "[…]",
			maxLen:   5,
		},
	}

	for tcName := range testCases {
		tc := testCases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				actual := stringsext.Truncate(tc.input, tc.maxLen)

				require.Equal(t, tc.expected, actual)
				require.True(t, len(actual) == tc.maxLen)
			},
		)
	}
}
