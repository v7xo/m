//go:build test_unit

package stringsext_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

func Test_KVDict_TextMarshal(t *testing.T) {
	thelper.Pre(t)

	type testCase struct {
		stdcase.CommonProps

		input    string
		expected string
	}

	testCases := map[string]testCase{
		"case0/happy0": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:    `rw,noatime,compress-force=zstd:1,ssd,discard=async,space_cache=v2,subvolid=293,subvol=/kili/app-data/radon_gitlab-config`,
			expected: `compress-force=zstd:1,discard=async,noatime,rw,space_cache=v2,ssd,subvol=/kili/app-data/radon_gitlab-config,subvolid=293`,
		},
		"case0/happy1a": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:    `"rw","n",cf="2 3"`,
			expected: `cf=2 3,n,rw`,
		},
		"case0/happy1b": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:    `"rw","n",cf=2 3`,
			expected: `cf=2 3,n,rw`,
		},
		"case0/happy2": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:    `,"r",`,
			expected: `r`,
		},
		"case0/happy3": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:    `,`,
			expected: ``,
		},
		"case0/happy4": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:    `a=,b`,
			expected: `a=,b`,
		},
		"case0/happy5": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:    `a=,b,"c"="d\"e"`,
			expected: `a=,b,c="d\"e"`,
		},
		"case0/happy6": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:    `a=,b,"c"="d\"e",f`,
			expected: `a=,b,c="d\"e",f`,
		},
		"case0/happy7": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:    `a=,b,"c"="d\"e",f=",g"`,
			expected: `a=,b,c="d\"e",f=",g"`,
		},
	}

	for tcName := range testCases {
		tc := testCases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				tc.PreLog(t)

				actualDict, _, err := stringsext.ParseKVDictString(tc.input)
				tc.RequireConformError(t, err)

				if tc.ShouldFail {
					return
				}

				actual, err := actualDict.MarshalText()
				tc.RequireConformError(t, err)

				if tc.ShouldFail {
					return
				}

				require.NoError(t, err)
				require.Equal(t, tc.expected, string(actual))
			},
		)
	}
}

func Test_ParseKVDict(t *testing.T) {
	thelper.Pre(t)

	type testCase struct {
		stdcase.CommonProps

		input    string
		expected stringsext.KVDictString
	}

	testCases := map[string]testCase{
		"case0/happy0": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input: `rw,noatime,compress-force=zstd:1,ssd,discard=async,space_cache=v2,subvolid=293,subvol=/kili/app-data/radon_gitlab-config`,
			expected: stringsext.KVDictString{
				"rw":             nil,
				"noatime":        nil,
				"compress-force": stringsext.ListString{"zstd:1"},
				"ssd":            nil,
				"discard":        stringsext.ListString{"async"},
				"space_cache":    stringsext.ListString{"v2"},
				"subvolid":       stringsext.ListString{"293"},
				"subvol":         stringsext.ListString{"/kili/app-data/radon_gitlab-config"},
			},
		},
		"case0/happy1": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input: `"rw","n",cf="2 3"`,
			expected: stringsext.KVDictString{
				"rw": nil,
				"n":  nil,
				"cf": stringsext.ListString{"2 3"},
			},
		},
		"case0/happy2": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input: `,"r",`,
			expected: stringsext.KVDictString{
				"r": nil,
			},
		},
		"case0/happy3": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input:    `,`,
			expected: nil,
		},
		"case1/invalid0": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,

				FailureProps: stdcase.FailureProps{
					ShouldFail: true,
				},
			},

			input:    `"r""`,
			expected: nil,
		},
		"case0/happy4": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input: `a=,b`,
			expected: stringsext.KVDictString{
				"a": stringsext.ListString{""},
				"b": nil,
			},
		},
		"case0/happy5": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input: `a=,b,"c"="d\"e"`,
			expected: stringsext.KVDictString{
				"a": stringsext.ListString{""},
				"b": nil,
				"c": stringsext.ListString{`d"e`},
			},
		},
		"case0/happy6": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input: `a=,b,"c"="d\"e",f`,
			expected: stringsext.KVDictString{
				"a": stringsext.ListString{""},
				"b": nil,
				"c": stringsext.ListString{`d"e`},
				"f": nil,
			},
		},
		"case0/happy7": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input: `a=,b,"c"="d\"e",f=",g"`,
			expected: stringsext.KVDictString{
				"a": stringsext.ListString{""},
				"b": nil,
				"c": stringsext.ListString{`d"e`},
				"f": stringsext.ListString{",g"},
			},
		},
		"case0/happy8": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input: `"a=b"=0`,
			expected: stringsext.KVDictString{
				"a=b": stringsext.ListString{"0"},
			},
		},
		"case1/invalid1": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,

				FailureProps: stdcase.FailureProps{
					ShouldFail: true,
				},
			},

			input:    `""`,
			expected: nil,
		},
		"case1/invalid2": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,

				FailureProps: stdcase.FailureProps{
					ShouldFail: true,
				},
			},

			input:    `"""`,
			expected: nil,
		},
		"case1/invalid3": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,

				FailureProps: stdcase.FailureProps{
					ShouldFail: true,
				},
			},

			input:    `"=`,
			expected: nil,
		},
		"case1/invalid4": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,

				FailureProps: stdcase.FailureProps{
					ShouldFail: true,
				},
			},

			input:    `a=,b,c\`,
			expected: nil,
		},
		"case1/invalid5": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,

				FailureProps: stdcase.FailureProps{
					ShouldFail: true,
				},
			},

			input:    `a=,b,c=\`,
			expected: nil,
		},
	}

	for tcName := range testCases {
		tc := testCases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				tc.PreLog(t)

				actual, _, err := stringsext.ParseKVDictString(tc.input)

				tc.RequireConformError(t, err)

				if tc.ShouldFail {
					return
				}

				require.NoError(t, err)
				require.Equal(t, tc.expected, actual)
			},
		)
	}
}
