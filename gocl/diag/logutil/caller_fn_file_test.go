//go:build test_unit

package logutil_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	u "rn.m.vo7.io/vo7io/m/gocl/diag/logutil"
	hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

type (
	callerFnFileFixtireType2                 struct{}
	callerFnFileFixtureType2[T1 any, T2 any] struct {
		f1 T1
		f2 T2
	}
)

func callerFnFile1() (fnName string, fileLine string) {
	return u.CallerFnFile(1)
}

func callerFnFile2[TArg1 any]() (fnName string, fileLine string) {
	return u.CallerFnFile(1)
}

func (v callerFnFileFixtireType2) callerFnFile() (fnName string, fileLine string) {
	return u.CallerFnFile(1)
}

func (v callerFnFileFixtureType2[T1, T2]) callerFnFile() (fnName string, fileLine string) {
	return u.CallerFnFile(1)
}

func Test_CallerFnFile(t *testing.T) {
	hlp.Pre(t)

	type (
		testCase struct {
			fn               func() (fnName string, fileLine string)
			expectedFnName   string
			expectedFileLine string
			stdcase.CommonProps
		}
	)

	testCases := map[string]testCase{
		"case00: non-generic func": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			fn: callerFnFile1,

			expectedFnName:   "logutil_test/callerFnFile1",
			expectedFileLine: "caller_fn_file_test.go:24",
		},
		"case01: non-generic method": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			fn: (callerFnFileFixtireType2{}).callerFnFile,

			expectedFnName:   "logutil_test/callerFnFileFixtireType2.callerFnFile",
			expectedFileLine: "caller_fn_file_test.go:32",
		},
		"case02: generic func": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			fn: callerFnFile2[string],

			expectedFnName:   "logutil_test/callerFnFile2[...]",
			expectedFileLine: "caller_fn_file_test.go:28",
		},
		"case03: generic method": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			fn: (callerFnFileFixtureType2[string, int]{}).callerFnFile,

			expectedFnName:   "logutil_test/callerFnFileFixtureType2[...].callerFnFile",
			expectedFileLine: "caller_fn_file_test.go:36",
		},
	}

	for tcName := range testCases {
		tc := testCases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				tc.PreLog(t)

				fnName, fileLine := tc.fn()
				tc.RequireConformError(t, nil)

				require.Equal(t, tc.expectedFnName, fnName)
				require.Equal(t, tc.expectedFileLine, fileLine)
			},
		)
	}
}
