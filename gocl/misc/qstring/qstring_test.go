//go:build test_unit

package qstring_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/qstring"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

type expectedVal[T any] struct {
	val  T
	fail stdcase.FailureProps
	eq   func(t *testing.T, e, a T)
}

func Test_QString(t *testing.T) {
	thelper.Pre(t)

	type testCase struct {
		stdcase.CommonProps

		input          string
		expectedSyntax qstring.Syntax
		expectedPlain  expectedVal[string]
		expectedLi     expectedVal[[]string]
		expectedKV     expectedVal[stringsext.KVOpts]
		expectedRE     expectedVal[*regexp.Regexp]
	}

	cases := map[string]testCase{
		"plain syntax/case00": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:          "",
			expectedSyntax: qstring.SyntaxPlain,

			expectedPlain: expectedVal[string]{val: ""},

			expectedKV: expectedVal[stringsext.KVOpts]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedLi: expectedVal[[]string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedRE: expectedVal[*regexp.Regexp]{fail: stdcase.FailureProps{ShouldFail: true}},
		},
		"plain syntax/case01": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:          ":",
			expectedSyntax: qstring.SyntaxPlain,

			expectedPlain: expectedVal[string]{val: ":"},

			expectedKV: expectedVal[stringsext.KVOpts]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedLi: expectedVal[[]string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedRE: expectedVal[*regexp.Regexp]{fail: stdcase.FailureProps{ShouldFail: true}},
		},
		"li syntax/case00": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:          "li:,a,,a,bb,,,c,",
			expectedSyntax: qstring.SyntaxLi,

			expectedLi: expectedVal[[]string]{val: []string{"a", "a", "bb", "c"}},

			expectedKV:    expectedVal[stringsext.KVOpts]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedPlain: expectedVal[string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedRE:    expectedVal[*regexp.Regexp]{fail: stdcase.FailureProps{ShouldFail: true}},
		},
		"li syntax/case01": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:          "li:",
			expectedSyntax: qstring.SyntaxLi,

			expectedLi: expectedVal[[]string]{val: nil},

			expectedKV:    expectedVal[stringsext.KVOpts]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedPlain: expectedVal[string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedRE:    expectedVal[*regexp.Regexp]{fail: stdcase.FailureProps{ShouldFail: true}},
		},
		"li syntax/case02": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug:     false,
				ShouldFail:      true,
				ShouldFailIsErr: errs.ErrInvalidSyntax,
			},

			input:          "li:a=b,c,c=",
			expectedSyntax: qstring.SyntaxLi,

			expectedLi: expectedVal[[]string]{fail: stdcase.FailureProps{ShouldFail: true, ShouldFailIsErr: errs.ErrNotInitializedValue}},

			expectedKV:    expectedVal[stringsext.KVOpts]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedPlain: expectedVal[string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedRE:    expectedVal[*regexp.Regexp]{fail: stdcase.FailureProps{ShouldFail: true}},
		},
		"li syntax/case03a": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug:     false,
				ShouldFail:      true,
				ShouldFailIsErr: errs.ErrInvalidSyntax,
			},

			input:          "li:c=",
			expectedSyntax: qstring.SyntaxLi,

			expectedLi: expectedVal[[]string]{fail: stdcase.FailureProps{ShouldFail: true, ShouldFailIsErr: errs.ErrNotInitializedValue}},

			expectedKV:    expectedVal[stringsext.KVOpts]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedPlain: expectedVal[string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedRE:    expectedVal[*regexp.Regexp]{fail: stdcase.FailureProps{ShouldFail: true}},
		},
		"li syntax/case03b": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug:     false,
				ShouldFail:      true,
				ShouldFailIsErr: errs.ErrInvalidSyntax,
			},

			input:          "li:,c=",
			expectedSyntax: qstring.SyntaxLi,

			expectedLi: expectedVal[[]string]{fail: stdcase.FailureProps{ShouldFail: true, ShouldFailIsErr: errs.ErrNotInitializedValue}},

			expectedKV:    expectedVal[stringsext.KVOpts]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedPlain: expectedVal[string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedRE:    expectedVal[*regexp.Regexp]{fail: stdcase.FailureProps{ShouldFail: true}},
		},
		"li syntax/case03c": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug:     false,
				ShouldFail:      true,
				ShouldFailIsErr: errs.ErrInvalidSyntax,
			},

			input:          "li:c=,",
			expectedSyntax: qstring.SyntaxLi,

			expectedLi: expectedVal[[]string]{fail: stdcase.FailureProps{ShouldFail: true, ShouldFailIsErr: errs.ErrNotInitializedValue}},

			expectedKV:    expectedVal[stringsext.KVOpts]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedPlain: expectedVal[string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedRE:    expectedVal[*regexp.Regexp]{fail: stdcase.FailureProps{ShouldFail: true}},
		},
		"li syntax/case03d": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug:     false,
				ShouldFail:      true,
				ShouldFailIsErr: errs.ErrInvalidSyntax,
			},

			input:          "li:=",
			expectedSyntax: qstring.SyntaxLi,

			expectedLi: expectedVal[[]string]{fail: stdcase.FailureProps{ShouldFail: true, ShouldFailIsErr: errs.ErrNotInitializedValue}},

			expectedKV:    expectedVal[stringsext.KVOpts]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedPlain: expectedVal[string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedRE:    expectedVal[*regexp.Regexp]{fail: stdcase.FailureProps{ShouldFail: true}},
		},
		"li syntax/case04": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:          "li:a,c,c",
			expectedSyntax: qstring.SyntaxLi,

			expectedLi: expectedVal[[]string]{val: []string{"a", "c", "c"}},

			expectedKV:    expectedVal[stringsext.KVOpts]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedPlain: expectedVal[string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedRE:    expectedVal[*regexp.Regexp]{fail: stdcase.FailureProps{ShouldFail: true}},
		},
		"li syntax/case05": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:          `li:"a=b",c,c`,
			expectedSyntax: qstring.SyntaxLi,

			expectedLi: expectedVal[[]string]{val: []string{`a=b`, "c", "c"}},

			expectedKV:    expectedVal[stringsext.KVOpts]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedPlain: expectedVal[string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedRE:    expectedVal[*regexp.Regexp]{fail: stdcase.FailureProps{ShouldFail: true}},
		},
		"kv syntax/case00": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:          `kv:"a=b"=,a=00,c,bb="hello"`,
			expectedSyntax: qstring.SyntaxKV,

			expectedKV: expectedVal[stringsext.KVOpts]{
				val: stringsext_thelper.KVOptsFromDictString(
					t,
					stringsext.KVDictString{
						"a=b": stringsext.ListString{""},
						"a":   stringsext.ListString{"00"},
						"c":   nil,
						"bb":  stringsext.ListString{"hello"},
					}),
				eq: hlpKVOptsEq,
			},

			expectedLi:    expectedVal[[]string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedPlain: expectedVal[string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedRE:    expectedVal[*regexp.Regexp]{fail: stdcase.FailureProps{ShouldFail: true}},
		},
		"kv syntax/case01": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug:     false,
				ShouldFail:      true,
				ShouldFailIsErr: errs.ErrInvalidSyntax,
			},

			input:          `kv:"a=b"=",a=00,c,bb="hello"`,
			expectedSyntax: qstring.SyntaxKV,

			expectedKV: expectedVal[stringsext.KVOpts]{fail: stdcase.FailureProps{ShouldFail: true, ShouldFailIsErr: errs.ErrNotInitializedValue}},

			expectedLi:    expectedVal[[]string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedPlain: expectedVal[string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedRE:    expectedVal[*regexp.Regexp]{fail: stdcase.FailureProps{ShouldFail: true}},
		},
		"re syntax/case00": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:          `re:.+`,
			expectedSyntax: qstring.SyntaxRE,

			expectedRE: expectedVal[*regexp.Regexp]{
				val: regexp.MustCompile(".+"),
			},

			expectedKV:    expectedVal[stringsext.KVOpts]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedLi:    expectedVal[[]string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedPlain: expectedVal[string]{fail: stdcase.FailureProps{ShouldFail: true}},
		},
		"re syntax/case01": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug:     false,
				ShouldFail:      true,
				ShouldFailIsErr: errs.ErrInvalidSyntax,
			},

			input:          `re:.+(`,
			expectedSyntax: qstring.SyntaxRE,

			expectedRE: expectedVal[*regexp.Regexp]{fail: stdcase.FailureProps{ShouldFail: true, ShouldFailIsErr: errs.ErrNotInitializedValue}},

			expectedKV:    expectedVal[stringsext.KVOpts]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedLi:    expectedVal[[]string]{fail: stdcase.FailureProps{ShouldFail: true}},
			expectedPlain: expectedVal[string]{fail: stdcase.FailureProps{ShouldFail: true}},
		},
	}

	for name := range cases {
		tc := cases[name]

		t.Run(
			name,
			func(t *testing.T) {
				tc.Pre(t)

				actual, err := qstring.Parse(tc.input)
				tc.RequireConformError(t, err)

				if !tc.ShouldFail {
					require.Equal(t, tc.expectedSyntax, actual.Syntax())
				}

				hlpValidateExpectedVal(t, tc.expectedPlain, actual.Plain)
				hlpValidateExpectedVal(t, tc.expectedKV, actual.KVOpts)
				hlpValidateExpectedVal(t, tc.expectedLi, actual.List)
				hlpValidateExpectedVal(t, tc.expectedRE, actual.Regexp)
			},
		)
	}
}

func hlpValidateExpectedVal[T any](t *testing.T, expected expectedVal[T], actualFn func() (T, error)) {
	t.Helper()

	var err error

	actual := expectedVal[T]{}
	actual.val, err = actualFn()

	expected.fail.RequireConformError(t, err)

	if expected.eq == nil {
		require.Equal(t, expected.val, actual.val)
	} else {
		expected.eq(t, expected.val, actual.val)
	}
}

func hlpKVOptsEq(t *testing.T, e, a stringsext.KVOpts) {
	t.Helper()

	require.Equal(t, e.CloneKVDict(), a.CloneKVDict())
}
