//go:build test_unit

package envext_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/conf/envext"
	"rn.m.vo7.io/vo7io/m/gocl/conf/startconf"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_Addenv_FromFileVarsProvider(t *testing.T) {
	ctx := thelper.Pre(t)

	type environStateType struct {
		val    string
		lookup bool
	}

	//nolint
	type testCaseType struct {
		shouldFailAsErr any
		shouldFailIsErr error

		expectedEnvs map[string]environStateType

		entryDir string

		desc string

		envFileNames []string
		shouldFail   bool
	}

	testCases := []testCaseType{
		{
			desc: "case00",

			entryDir: "./" + startconf.DefaultTestDataDirName + "/case00/a/b/c",
			envFileNames: []string{
				"conf-test-local.yml",
				"conf-test.yml",
			},
			expectedEnvs: map[string]environStateType{
				"COMMONLIB_GO_TEST_VAR1": {
					lookup: true,
					val:    `ABC`,
				},
				"COMMONLIB_GO_TEST_VAR2": {
					lookup: true,
					val: `Line1_local
Line2_local

Line4_local
`,
				},
				"COMMONLIB_GO_TEST_VAR3": {
					lookup: true,
					val:    ``,
				},
				"COMMONLIB_GO_TEST_VAR4": {
					lookup: true,
					val:    `ABC`,
				},
				"COMMONLIB_GO_TEST_VAR5": {
					lookup: false,
					val:    ``,
				},
				"COMMONLIB_GO_TEST_VAR6": {
					lookup: false,
					val:    ``,
				},
				"COMMONLIB_GO_TEST_VAR7": {
					lookup: true,
					val:    `0123456789`,
				},
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(
			tc.desc,
			func(t *testing.T) {
				prov := envext.VarProviderFile(tc.entryDir, tc.envFileNames...)
				varsDict, err := prov.VariablesDict(ctx)

				if err == nil {
					_, err = envext.Addenv(varsDict)
				}

				if tc.shouldFail {
					require.Error(t, err, "error expected while executing test case")

					if tc.shouldFailAsErr != nil {
						require.ErrorAs(t, err, tc.shouldFailAsErr)

						return
					}

					if tc.shouldFailIsErr != nil {
						require.ErrorIs(t, err, tc.shouldFailIsErr)

						return
					}

					return
				}

				require.NoError(t, err, "unexpected error while executing test case")

				for envName, envState := range tc.expectedEnvs {
					actualVal, actualLookup := os.LookupEnv(envName)

					require.Equal(t, envState.lookup, actualLookup, "unexpected env lookup; var name: %s", envName)

					if envState.lookup {
						require.Equal(t, envState.val, actualVal, "unexpected env val; var name: %s", envName)
					}
				}
			},
		)
	}
}
