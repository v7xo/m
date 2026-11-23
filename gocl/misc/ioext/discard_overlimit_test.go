//go:build test_unit

package ioext_test

import (
	"bytes"
	"io"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	clio "rn.m.vo7.io/vo7io/m/gocl/misc/ioext"
	hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

type limitedBuffer struct {
	*bytes.Buffer

	limit int
}

func (b *limitedBuffer) Write(p []byte) (n int, err error) {
	if b.limit <= 0 {
		err = io.ErrShortWrite
		return
	}
	if len(p) > b.limit {
		p = p[:b.limit]
		err = io.ErrShortWrite
	}

	n, writeErr := b.Buffer.Write(p)
	if writeErr != nil {
		err = writeErr
	}

	b.limit -= n

	return
}

func Test_DiscardOverlimit(t *testing.T) {
	hlp.PreLog(t)

	type (
		testCaseWriteOp struct {
			expectedWriteResultErr stdcase.FailureProps
			inputData              []byte
			expectedTotalDiscarded int
			expectedWriteResultN   int
		}

		testCase struct {
			buf *limitedBuffer

			inputWrites                    []testCaseWriteOp
			expectedOverallWriteResultData []byte

			stdcase.CommonProps

			inputDiscardOverlimit int
		}
	)

	testCases := map[string]testCase{
		"case00": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			buf: &limitedBuffer{
				Buffer: &bytes.Buffer{},
				limit:  math.MaxInt,
			},

			inputDiscardOverlimit: 10,
			inputWrites: []testCaseWriteOp{
				{
					inputData: []byte("012345678"),
					expectedWriteResultErr: stdcase.FailureProps{
						ShouldFail: false,
					},
					expectedTotalDiscarded: 0,
					expectedWriteResultN:   9,
				},
				{
					inputData: []byte("9abc"),
					expectedWriteResultErr: stdcase.FailureProps{
						ShouldFail: false,
					},
					expectedTotalDiscarded: 3,
					expectedWriteResultN:   4,
				},
				{
					inputData: []byte("def"),
					expectedWriteResultErr: stdcase.FailureProps{
						ShouldFail: false,
					},
					expectedTotalDiscarded: 6,
					expectedWriteResultN:   3,
				},
			},

			expectedOverallWriteResultData: []byte("0123456789"),
		},
		"case01": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			buf: &limitedBuffer{
				Buffer: &bytes.Buffer{},
				limit:  math.MaxInt,
			},

			inputDiscardOverlimit: 10,
			inputWrites: []testCaseWriteOp{
				{
					inputData: []byte("0123456789abcdef"),
					expectedWriteResultErr: stdcase.FailureProps{
						ShouldFail: false,
					},
					expectedTotalDiscarded: 6,
					expectedWriteResultN:   16,
				},
			},

			expectedOverallWriteResultData: []byte("0123456789"),
		},
		"case02/underlying writer failed": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			buf: &limitedBuffer{
				Buffer: &bytes.Buffer{},
				limit:  5,
			},

			inputDiscardOverlimit: 10,
			inputWrites: []testCaseWriteOp{
				{
					inputData: []byte("0123456789"),
					expectedWriteResultErr: stdcase.FailureProps{
						ShouldFail:      true,
						ShouldFailIsErr: io.ErrShortWrite,
					},
					expectedTotalDiscarded: 0,
					expectedWriteResultN:   5,
				},
				{
					inputData: []byte("abc"),
					expectedWriteResultErr: stdcase.FailureProps{
						ShouldFail:      true,
						ShouldFailIsErr: io.ErrShortWrite,
					},
					expectedTotalDiscarded: 0,
					expectedWriteResultN:   0,
				},
			},

			expectedOverallWriteResultData: []byte("01234"),
		},
	}

	for tcName := range testCases {
		tc := testCases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				tc.PreLog(t)

				wBuf := tc.buf
				w := clio.NewDiscardOverlimit(tc.inputDiscardOverlimit, wBuf, false)

				for i := 0; i < len(tc.inputWrites); i++ {
					op := tc.inputWrites[i]

					n, err := w.Write(op.inputData)

					op.expectedWriteResultErr.RequireConformErrorMsgf(t, err, "result error not matches expected one; write operation: %d", i)
					require.Equal(t, op.expectedWriteResultN, n, "count of written not matches expected one; write operation: %d", i)
					require.Equal(t, op.expectedTotalDiscarded, w.TotalDiscarded(), "total discarded not matches expected one; write operation: %d", i)
				}

				require.Equal(t, tc.expectedOverallWriteResultData, wBuf.Bytes(), "overall write result of test case not matches expected one")
			},
		)
	}
}
