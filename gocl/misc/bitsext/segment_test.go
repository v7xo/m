//go:build test_unit

package bitsext_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/bitsext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	clhlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

func Test_Segment(t *testing.T) {
	clhlp.Pre(t)

	type testCase struct {
		stdcase.CommonProps

		inputMap  bitsext.SegmentsMap
		inputVals []uint64

		expectedCarrier uint64
	}

	cases := map[string]testCase{
		"case00/default": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},
		},
		"case01/happy": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			inputMap:        bitsext.SegmentsMap{10, 10, 10, 10, 10, 10, 4},
			inputVals:       []uint64{1023, 1023, 1023, 1023, 1023, 1023, 15},
			expectedCarrier: 0b1111111111_1111111111_1111111111_1111111111_1111111111_1111111111_1111,
		},
		"case02/happy": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			inputMap:        bitsext.SegmentsMap{4, 10, 10, 10, 10, 10, 10},
			inputVals:       []uint64{15, 1023, 1023, 1023, 1023, 1023, 1023},
			expectedCarrier: 0b1111_1111111111_1111111111_1111111111_1111111111_1111111111_1111111111,
		},
		"case03/happy": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			inputMap:        bitsext.SegmentsMap{4, 1, 19, 10, 10, 10, 10},
			inputVals:       []uint64{15, 1, 524287, 1023, 1023, 1023, 1023},
			expectedCarrier: 0b1111_1_1111111111111111111_1111111111_1111111111_1111111111_1111111111,
		},
		"case04/happy": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			inputMap:        bitsext.SegmentsMap{4, 1, 19, 10, 10, 10, 10},
			inputVals:       []uint64{15, 1, 524287, 7, 1023, 1023, 1023},
			expectedCarrier: 0b1111_1_1111111111111111111_0000000111_1111111111_1111111111_1111111111,
		},
		"case05/happy": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			inputMap:        bitsext.SegmentsMap{4, 1, 19, 10, 10, 10, 10},
			inputVals:       []uint64{15, 1, 524287, 7, 1023, 0, 1023},
			expectedCarrier: 0b1111_1_1111111111111111111_0000000111_1111111111_0000000000_1111111111,
		},
		"case06/fail/invalid-segment-size": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,

				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailIsErr: errs.ErrBadArgs,
				},
			},

			inputMap:        bitsext.SegmentsMap{4, 1, 19, 10, 10, 10, 20},
			inputVals:       []uint64{15, 1, 524287, 7, 1023, 1023, 1023},
			expectedCarrier: 0b1111_1_1111111111111111111_0000000111_1111111111_1111111111_1111111111,
		},
		"case07/fail/invalid-segment-size": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,

				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailIsErr: errs.ErrBadArgs,
				},
			},

			inputMap:        bitsext.SegmentsMap{4, 1, 19, 10, 10, 10, 0},
			inputVals:       []uint64{15, 1, 524287, 7, 1023, 1023, 1023},
			expectedCarrier: 0b1111_1_1111111111111111111_0000000111_1111111111_1111111111_1111111111,
		},
		"case08/fail/invalid-segment-val": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,

				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailIsErr: errs.ErrBadArgs,
				},
			},

			inputMap:        bitsext.SegmentsMap{4, 1, 19, 10, 10, 10, 10},
			inputVals:       []uint64{15, 1, 524287, 7, 1023, 1023, 1024},
			expectedCarrier: 0b1111_1_1111111111111111111_0000000111_1111111111_1111111111_1111111111,
		},
	}

	for tcName := range cases {
		tc := cases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				tc.Pre(t)

				require.True(t, len(tc.inputMap) == len(tc.inputVals))

				var (
					actualCarrier uint64
					err           error
				)

				for i := range len(tc.inputMap) {
					if err = bitsext.SetSegmentVal(&actualCarrier, tc.inputMap, i, tc.inputVals[i]); err != nil {
						break
					}
				}

				tc.RequireConformError(t, err)

				if tc.ShouldFail {
					return
				}

				require.Equal(t, tc.expectedCarrier, actualCarrier)

				for i := range len(tc.inputMap) {
					val, err := bitsext.SegmentVal(actualCarrier, tc.inputMap, i)
					require.NoError(t, err)
					require.Equal(t, tc.inputVals[i], val)
				}
			},
		)
	}
}
