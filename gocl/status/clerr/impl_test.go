//go:build test_unit

package clerr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	opst "rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
	clhlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

func Test_Wrap(t *testing.T) {
	clhlp.Pre(t)

	type testCase struct {
		Chain      error
		IsOpTarget error

		ExpectedErrorString string
		ExpectedLevel       level.Name
		ExpectedOpStatus    opst.Name
		stdcase.CommonProps
		ExpectedIsOpResult bool
	}

	const (
		existingErr1Msg   = "error1"
		existingPlfErrMsg = "error2"
	)

	var (
		existingErr1   = errors.New(existingErr1Msg)
		existingPlfErr = clerr.ErrStatusLevelf(opst.NotFound, level.Debug, existingPlfErrMsg)
	)

	_ = existingPlfErr

	cases := map[string]testCase{
		"case00: declared error wrapped with CL error": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			Chain:      clerr.Wrap(existingErr1),
			IsOpTarget: existingErr1,

			ExpectedIsOpResult:  true,
			ExpectedOpStatus:    opst.Fail,
			ExpectedLevel:       level.Error,
			ExpectedErrorString: existingErr1Msg,
		},
		"case01: change level and status": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			Chain:      clerr.Wrap(existingErr1).WithStatusLevel(opst.IdempotencyViolation, level.Warn),
			IsOpTarget: existingErr1,

			ExpectedIsOpResult:  true,
			ExpectedOpStatus:    opst.IdempotencyViolation,
			ExpectedLevel:       level.Warn,
			ExpectedErrorString: fmt.Sprintf("[ %s %s ] %s", level.Warn, opst.IdempotencyViolation, existingErr1Msg),
		},
		"case02: wrapping with standard fmt.Errorf": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			Chain:      fmt.Errorf("some wrapping text; err: %w", clerr.Wrap(existingErr1).WithLevel(level.Warn)),
			IsOpTarget: existingErr1,

			ExpectedIsOpResult:  true,
			ExpectedOpStatus:    opst.Fail,
			ExpectedLevel:       level.Warn,
			ExpectedErrorString: fmt.Sprintf("some wrapping text; err: [ %s %s ] %s", level.Warn, opst.Fail, existingErr1Msg),
		},
		"case03: wrapping with standard fmt.Errorf + change status": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			Chain:      fmt.Errorf("some wrapping text; err: %w", clerr.Errf("unexpected error; foo: %d; err: %w", 10, existingErr1).WithStatus(opst.Cancel)),
			IsOpTarget: existingErr1,

			ExpectedIsOpResult:  true,
			ExpectedOpStatus:    opst.Cancel,
			ExpectedLevel:       level.Error,
			ExpectedErrorString: fmt.Sprintf("some wrapping text; err: [ %s %s ] unexpected error; foo: 10; err: %s", level.Error, opst.Cancel, existingErr1Msg),
		},
		"case04: existing CL error": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			Chain:      existingPlfErr,
			IsOpTarget: existingPlfErr,

			ExpectedIsOpResult:  true,
			ExpectedOpStatus:    opst.NotFound,
			ExpectedLevel:       level.Debug,
			ExpectedErrorString: fmt.Sprintf("[ %s %s ] %s", level.Debug, opst.NotFound, existingPlfErrMsg),
		},
		"case05: declared CL error wrapped with new CL error": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			Chain:      clerr.WrapStatus(opst.IdempotencyViolation, existingPlfErr).WithLevel(level.Error),
			IsOpTarget: existingPlfErr,

			ExpectedIsOpResult:  true,
			ExpectedOpStatus:    opst.IdempotencyViolation,
			ExpectedLevel:       level.Error,
			ExpectedErrorString: fmt.Sprintf("[ %s %s ] %s", level.Error, opst.IdempotencyViolation, existingPlfErr.Error()),
		},
	}

	for tcName := range cases {
		tc := cases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				tc.PreLog(t)

				isOpRes := errors.Is(tc.Chain, tc.IsOpTarget)
				require.Equal(t, tc.ExpectedIsOpResult, isOpRes)
				require.Equal(t, tc.ExpectedOpStatus, errs.TranslateStatus(tc.Chain))
				require.Equal(t, tc.ExpectedLevel, errs.TranslateLevel(tc.Chain))
				require.Equal(t, tc.ExpectedErrorString, tc.Chain.Error())
			},
		)
	}
}
