//go:build test_unit

package flenv_test

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

type testConfA struct {
	fieldX string
}

const (
	testConfASelfPrefix = "conf-a"
	testConfAFieldXFlag = "field-x"
)

var _ flenv.FlagSetBinderPrefix = &testConfA{}

func (c *testConfA) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	prefix = flenv.NormalizeSelfPrefix(prefix, testConfASelfPrefix)

	flset.StringVar(&c.fieldX, prefix+testConfAFieldXFlag, "", "")

	return nil
}

func Test_NormalizeSelfPrefix(t *testing.T) {
	thelper.Pre(t)

	type testCase struct {
		stdcase.CommonProps

		inputPrefix         string
		expectedEnvironment map[string]struct{}
	}

	cases := map[string]testCase{
		"case a/empty prefix": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			inputPrefix: "",
			expectedEnvironment: map[string]struct{}{
				"CONF_A_FIELD_X": {},
			},
		},
		"case b1/prefix same as self prefix": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			inputPrefix: testConfASelfPrefix,
			expectedEnvironment: map[string]struct{}{
				"CONF_A_FIELD_X": {},
			},
		},
		"case b2/prefix same as self prefix plus delim": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			inputPrefix: testConfASelfPrefix + flenv.FlagNameDelimiter,
			expectedEnvironment: map[string]struct{}{
				"CONF_A_FIELD_X": {},
			},
		},
		"case b3/prefix same as self prefix 2": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			inputPrefix: testConfASelfPrefix + flenv.FlagNameDelimiter + testConfASelfPrefix,
			expectedEnvironment: map[string]struct{}{
				"CONF_A_CONF_A_FIELD_X": {},
			},
		},
		"case b4/prefix same as self prefix plus delim 2": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			inputPrefix: testConfASelfPrefix + flenv.FlagNameDelimiter + testConfASelfPrefix + flenv.FlagNameDelimiter,
			expectedEnvironment: map[string]struct{}{
				"CONF_A_CONF_A_FIELD_X": {},
			},
		},
		"case c1/any prefix": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			inputPrefix: "foo",
			expectedEnvironment: map[string]struct{}{
				"FOO_CONF_A_FIELD_X": {},
			},
		},
		"case c2/any prefix": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			inputPrefix: "foo-",
			expectedEnvironment: map[string]struct{}{
				"FOO_CONF_A_FIELD_X": {},
			},
		},
	}

	for tcName := range cases {
		tc := cases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				tc.Pre(t)

				actual, err := flenv.FlagEnvironmentNamesForBinderPrefix(tc.inputPrefix, &testConfA{})
				tc.RequireConformError(t, err)

				if tc.ShouldFail {
					return
				}

				require.Equal(t, tc.expectedEnvironment, actual)
			})
	}
}
