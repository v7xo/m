//go:build test_unit

package osutil_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ptrof"
	hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

func Test_ExecTextOutput(t *testing.T) {
	hlp.Pre(t)

	type expectation struct {
		NonZeroLength *bool
		FixedLength   *int
	}

	type testCase struct {
		InputName          string
		InputArg           []string
		InputMaxTokens     int
		InputNilTerminator bool

		Expected *expectation

		stdcase.CommonProps
	}

	cases := map[string]testCase{
		"case00/git/status": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			InputName: "git",
			InputArg: []string{
				"status",
				"-uall",
				"-z",
				"--porcelain=v1",
				"--ignored=traditional",
			},
			InputMaxTokens:     65536,
			InputNilTerminator: true,

			Expected: &expectation{},
		},
		//
		"case01/git/show-ref": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			InputName: "git",
			InputArg: []string{
				"show-ref",
				"--hash",
				"--verify",
				"HEAD",
			},
			InputMaxTokens:     1,
			InputNilTerminator: false,

			Expected: &expectation{
				FixedLength: ptrof.Of(1),
			},
		},
	}

	for name := range cases {
		tc := cases[name]

		t.Run(
			name,
			func(t *testing.T) {
				ctx, log := tc.PreLog(t)
				actual, err := osutil.ExecTextOutput(
					ctx,
					osutil.ExecParamsTextOutput{
						Log:            log,
						Name:           tc.InputName,
						Args:           tc.InputArg,
						StdoutMaxLines: tc.InputMaxTokens,
						NilTerminator:  tc.InputNilTerminator,
					},
				)

				tc.RequireConformError(t, err)

				if tc.ShouldFail || tc.Expected == nil {
					return
				}

				if tc.Expected.NonZeroLength != nil {
					switch {
					case *tc.Expected.NonZeroLength:
						require.True(t, len(actual) != 0, "unexpected mismatch of result")
					default:
						require.True(t, len(actual) == 0, "unexpected mismatch of result")
					}
				}

				if tc.Expected.FixedLength != nil {
					require.Len(t, actual, *tc.Expected.FixedLength, "unexpected mismatch of result")
				}
			},
		)
	}
}
