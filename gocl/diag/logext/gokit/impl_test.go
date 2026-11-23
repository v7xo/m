//go:build test_unit

package gokit_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/go-kit/log/level"
	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext/gokit"
	cllevel "rn.m.vo7.io/vo7io/m/gocl/status/level"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

type logOutputMsg struct {
	Err   string `json:"err"`
	Level string `json:"level"`
	Msg   string `json:"message"`
}

func unmarshalLogOutputMsgsJSON(t *testing.T, data []byte) []logOutputMsg {
	t.Helper()

	res := []logOutputMsg{}

	err := json.Unmarshal(data, &res)
	require.NoError(t, err, "error while unmarshalling log output JSON")

	return res
}

func Test_ErrLevelAndNonErrLevelOutput(t *testing.T) {
	ctx := thelper.Pre(t)

	type logInputMsg struct {
		err   error
		level cllevel.Name
		msg   string
	}

	type testCaseT struct {
		logLevel cllevel.Name

		input []logInputMsg

		expectedErrLevelOutput    []logOutputMsg
		expectedNonErrLevelOutput []logOutputMsg
	}

	const (
		debugLogMsg  = "debug message"
		errLogMsg    = "error message"
		infoLogMsg   = "information message"
		tracegLogMsg = "trace message"
		warnLogMsg   = "warning message"
	)
	errLogErr := errors.New("this is test error")

	testCases := map[string]testCaseT{
		"case00: debug level logger": {
			logLevel: cllevel.Debug,

			input: []logInputMsg{
				{
					level: cllevel.Debug,
					msg:   debugLogMsg,
				},
				{
					level: cllevel.Trace,
					msg:   tracegLogMsg,
				},
				{
					level: cllevel.Info,
					msg:   infoLogMsg,
				},
				{
					level: cllevel.Warn,
					msg:   warnLogMsg,
				},
				{
					level: cllevel.Error,
					msg:   errLogMsg,
					err:   errLogErr,
				},
			},

			expectedErrLevelOutput: []logOutputMsg{
				{
					Err:   errLogErr.Error(),
					Level: level.ErrorValue().String(),
					Msg:   errLogMsg,
				},
			},
			expectedNonErrLevelOutput: []logOutputMsg{
				{
					Level: level.DebugValue().String(),
					Msg:   debugLogMsg,
				},
				{
					Level: level.DebugValue().String(), // trace level equals to debug.
					Msg:   tracegLogMsg,
				},
				{
					Level: level.InfoValue().String(),
					Msg:   infoLogMsg,
				},
				{
					Level: level.WarnValue().String(),
					Msg:   warnLogMsg,
				},
			},
		},
		"case01: error level logger": {
			logLevel: cllevel.Error,

			input: []logInputMsg{
				{
					level: cllevel.Debug,
					msg:   debugLogMsg,
				},
				{
					level: cllevel.Trace,
					msg:   tracegLogMsg,
				},
				{
					level: cllevel.Info,
					msg:   infoLogMsg,
				},
				{
					level: cllevel.Warn,
					msg:   warnLogMsg,
				},
				{
					level: cllevel.Error,
					msg:   errLogMsg,
					err:   errLogErr,
				},
			},

			expectedErrLevelOutput: []logOutputMsg{
				{
					Err:   errLogErr.Error(),
					Level: level.ErrorValue().String(),
					Msg:   errLogMsg,
				},
			},
			expectedNonErrLevelOutput: []logOutputMsg{},
		},
		"case02: info level logger": {
			logLevel: cllevel.Info,

			input: []logInputMsg{
				{
					level: cllevel.Debug,
					msg:   debugLogMsg,
				},
				{
					level: cllevel.Trace,
					msg:   tracegLogMsg,
				},
				{
					level: cllevel.Info,
					msg:   infoLogMsg,
				},
				{
					level: cllevel.Warn,
					msg:   warnLogMsg,
				},
			},

			expectedErrLevelOutput: []logOutputMsg{},
			expectedNonErrLevelOutput: []logOutputMsg{
				{
					Level: level.InfoValue().String(),
					Msg:   infoLogMsg,
				},
				{
					Level: level.WarnValue().String(),
					Msg:   warnLogMsg,
				},
			},
		},
	}

	for tcName := range testCases {
		tc := testCases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				t.Parallel()

				nonErrLevelOutputWriter := bytes.NewBuffer(nil)
				errLevelOutputWriter := bytes.NewBuffer(nil)

				logOpts := gokit.
					Option().
					ErrLevelWriter(errLevelOutputWriter).
					NonErrorLevelWriter(nonErrLevelOutputWriter).
					Format(logext.FormatJSON).
					Level(tc.logLevel)

				log, err := gokit.New(logOpts)
				require.NoError(t, err)

				for _, inpMsg := range tc.input {
					log.
						Level(ctx, inpMsg.level.Num(), inpMsg.err).
						WithSendMsg(inpMsg.msg).
						Send()
				}

				outputJSON := fmt.Sprintf(
					"[%s]",
					strings.ReplaceAll(
						strings.TrimSuffix(nonErrLevelOutputWriter.String(), "\n"),
						"\n",
						",",
					),
				)
				nonErrLevelOutput := unmarshalLogOutputMsgsJSON(t, []byte(outputJSON))
				require.Equal(t, tc.expectedNonErrLevelOutput, nonErrLevelOutput, "non-error level log output not equal to expected one")

				outputJSON = fmt.Sprintf(
					"[%s]",
					strings.ReplaceAll(
						strings.TrimSuffix(errLevelOutputWriter.String(), "\n"),
						"\n",
						",",
					),
				)
				errLevelOutput := unmarshalLogOutputMsgsJSON(t, []byte(outputJSON))
				require.Equal(t, tc.expectedErrLevelOutput, errLevelOutput, "error level log output not equal to expected one")
			},
		)
	}
}
