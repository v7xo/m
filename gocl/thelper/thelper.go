package thelper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

// [PreLog] executes test preparation steps: environment setup, creation of test context, log.
func PreLog(t *testing.T) (context.Context, logext.Logger) {
	t.Helper()

	return (stdcase.CommonProps{}).PreLog(t)
}

// [Pre] executes test preparation steps: sets environment, creates context, enriches created context with logger.
func Pre(t *testing.T) context.Context {
	t.Helper()

	return (stdcase.CommonProps{}).Pre(t)
}

func CleanupFn(t *testing.T, fn func() error, msgAndArgs ...any) func() {
	t.Helper()

	return func() {
		t.Helper()

		userMsg := ""
		if len(msgAndArgs) != 0 {
			userMsg = "; " + fmt.Sprintf(fmt.Sprintf("%s", msgAndArgs[0]), msgAndArgs[1:]...)
		}

		if err := fn(); err != nil {
			t.Logf("[ERROR] error while executing cleanup func%s; %s", userMsg, err)

			if !t.Failed() {
				defer func() {
					if panicErr := errs.PanicErr(recover()); panicErr != nil {
						t.Logf("[ERROR] error while marking test '%s' as failed; %s", t.Name(), panicErr)
					}
				}()

				t.Fail()
			}
		}
	}
}

func Cleanup(t *testing.T, fn func() error, msgAndArgs ...any) {
	t.Helper()

	t.Cleanup(CleanupFn(t, fn, msgAndArgs...))
}

func Factorial(t *testing.T, c uint8) int {
	t.Helper()

	if c == 0 {
		return 1
	}

	r := int(1)

	for i := range c {
		r = r * (int(i) + 1)
	}

	return r
}

func Permutations(t *testing.T, c uint8) [][]uint8 {
	t.Helper()

	if c == 0 {
		return nil
	}

	s := make([]uint8, c)
	qty := Factorial(t, c)

	for i := range c {
		s[i] = i
	}

	if c == 1 {
		return [][]uint8{s}
	}

	res := make([][]uint8, qty)
	resIdx := -1

	var helper func(a []uint8, i int)

	helper = func(a []uint8, left int) {
		if left == len(a)-1 {
			c := make([]uint8, len(a))

			copy(c, a)

			resIdx++
			res[resIdx] = c

			return
		}

		helper(a, left+1)

		for j := left + 1; j < len(a); j++ {
			a[left], a[j] = a[j], a[left]

			helper(a, left+1)

			a[left], a[j] = a[j], a[left]
		}
	}

	helper(s, 0)

	return res
}

func RequireNoErrFn[T any](t *testing.T, fn func() (T, error), msgAndArgs ...any) T {
	t.Helper()

	msgPref := ""
	msgArgs := []any{}

	if len(msgAndArgs) != 0 {
		msgPref = msgAndArgs[0].(string) //nolint:forcetypeassert
		msgPref = msgPref + "; "
		msgArgs = msgAndArgs[1:]
	}

	require.NotNil(t, fn, msgPref+"nil-pointer fn specified", msgArgs)

	res, err := fn()
	require.NoError(t, err, msgPref+"unexpected error", msgArgs)

	return res
}

func RequireNotNilFn[T any](t *testing.T, fn func() (T, error), msgAndArgs ...any) T {
	t.Helper()

	msgPref := ""
	msgArgs := []any{}

	if len(msgAndArgs) != 0 {
		msgPref = msgAndArgs[0].(string) //nolint:forcetypeassert
		msgPref = msgPref + "; "
		msgArgs = msgAndArgs[1:]
	}

	require.NotNil(t, fn, msgPref+"nil-pointer fn specified", msgArgs)

	res, err := fn()
	require.NoError(t, err, msgPref+"unexpected error", msgArgs)
	require.NotNil(t, res, msgPref+"unexpected nil-result", msgArgs)

	return res
}

func RequireFileContentByPathEqYAML(t *testing.T, expectedPath, actualPath string) {
	t.Helper()

	RequireFileContentByPathEqYAMLf(t, expectedPath, actualPath, "")
}

func RequireFileContentByPathEqYAMLf(t *testing.T, expectedPath, actualPath string, msg string, args ...any) {
	t.Helper()

	userMsg := ""
	if len(msg) != 0 {
		userMsg = "; " + fmt.Sprintf(msg, args...)
	}

	expectedDocuments := decodeYAMLDocuments(t, expectedPath, userMsg)
	actualDocuments := decodeYAMLDocuments(t, actualPath, userMsg)

	require.Equal(
		t,
		len(expectedDocuments),
		len(actualDocuments),
		"expected and actual number of decoded YAML documents differ",
	)

	for i := range expectedDocuments {
		require.Equal(
			t,
			expectedDocuments[i],
			actualDocuments[i],
			"actual content of YAML document doesn't match expected one; i: %d; %s", i, userMsg,
		)
	}
}

func IsDebuggerAttached(t *testing.T) bool {
	t.Helper()

	status, err := os.ReadFile("/proc/self/status")
	require.NoError(t, err)

	tracerField := []byte("TracerPid:\t")

	idx := bytes.Index(status, tracerField)
	if idx == -1 {
		return false
	}

	start := idx + len(tracerField)

	end := bytes.IndexByte(status[start:], '\n')
	if end == -1 {
		return false
	}

	pid, err := strconv.Atoi(string(bytes.TrimSpace(status[start : start+end])))
	require.NoError(t, err)

	return pid > 0
}

func decodeYAMLDocuments(t *testing.T, filepath string, userMsg string) []any {
	t.Helper()

	f, err := os.Open(filepath)
	require.NoError(t, err, "unexpected error while opening file with YAML; file: %s%s", filepath, userMsg)

	defer f.Close()

	var (
		document  any
		documents []any
		decoder   = yaml.NewDecoder(f)
	)

	for i := 0; ; i++ {
		if err := decoder.Decode(&document); errors.Is(err, io.EOF) {
			break
		} else {
			require.NoError(t, err, "unexpected error while decoding YAML document; file: %s; i: %d; %s", filepath, i, userMsg)
		}

		documents = append(documents, document)
	}

	return documents
}
