//go:build test_unit

package query_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	db_errs "rn.m.vo7.io/vo7io/m/gocl/status/errs"
	db_q "rn.m.vo7.io/vo7io/m/gocl/storage/query"
	hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

func TestToPositionalSyntax(t *testing.T) {
	hlp.PreLog(t)

	type testCase struct {
		inputQueryParams    any
		inputQuery          string
		desc                string
		expectedQuery       string
		expectedQueryParams []interface{}
		stdcase.CommonProps
	}

	testCases := []testCase{
		{
			desc: "test001_query_param_is_missing",
			CommonProps: stdcase.CommonProps{
				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailIsErr: db_errs.ErrNamedQueryParamMissing,
				},
			},

			inputQuery: "select * from t where t.f1 = :param1",
			inputQueryParams: struct {
				p2 string `db:"p2"`
			}{
				p2: "",
			},
		},
		{
			desc: "test002_same_query_param_multiple_times",

			inputQuery: `select t.*, :param1 as param1 from t where t.f1 = :param1 and t.f2 = :param1`,
			inputQueryParams: struct {
				Param1 string `db:"param1"`
			}{
				Param1: "param1_value",
			},

			expectedQuery: `select t.*, $1 as param1 from t where t.f1 = $1 and t.f2 = $1`,
			expectedQueryParams: []interface{}{
				"param1_value",
			},
		},
		{
			desc: "test003_query_param_name_invalid",
			CommonProps: stdcase.CommonProps{
				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailIsErr: db_errs.ErrInvalidQueryParamName,
				},
			},

			inputQuery: `select t.*, : as param1 from t where t.f1 = :param1`,
			inputQueryParams: struct {
				Param1 string `db:"param1"`
			}{
				Param1: "param1_value",
			},
		},
		{
			desc: "test004_query_param_name_invalid",
			CommonProps: stdcase.CommonProps{
				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailIsErr: db_errs.ErrInvalidQueryParamName,
				},
			},

			inputQuery: `select t.*, :0param as param1 from t`,
			inputQueryParams: struct {
				Param1 string `db:"param1"`
			}{
				Param1: "param1_value",
			},
		},
		{
			desc: "test005_query_param_name_invalid",
			CommonProps: stdcase.CommonProps{
				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailIsErr: db_errs.ErrInvalidQueryParamName,
				},
			},

			inputQuery: `select t.*, :-param as param1 from t`,
			inputQueryParams: struct {
				Param1 string `db:"param1"`
			}{
				Param1: "param1_value",
			},
		},
		{
			desc: "test006_query_param_name_invalid",
			CommonProps: stdcase.CommonProps{
				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailIsErr: db_errs.ErrInvalidQueryParamName,
				},
			},

			inputQuery: `select t.*, :param1: as param1 from t`,
			inputQueryParams: struct {
				Param1 string `db:"param1"`
			}{
				Param1: "param1_value",
			},
		},
		{
			desc: "test007_query_param_args_ambigious_field",
			CommonProps: stdcase.CommonProps{
				FailureProps: stdcase.FailureProps{
					ShouldFail: true,
				},
			},

			inputQuery: `select t.*, :_param1 as param1 from t where t.f1 = :param2`,
			inputQueryParams: struct {
				Param1 string `db:"_param1"`
				param2 string `db:"_param1"`
			}{
				Param1: "param1_value",
				param2: "param2_value",
			},
		},
		{
			desc: "test008_multiple_query_params",

			inputQuery: `select t.*, :_param1 as param1 from t where t.f1 = :param2 and t.f2 = ':param2:param3' and t.f3=:_param1`,
			inputQueryParams: struct {
				Param1 string `db:"_param1"`
				Param2 string `db:"param2"`
				Param3 []int  `db:"param3"`
			}{
				Param1: "param1_value",
				Param2: "param2_value",
				Param3: nil,
			},

			expectedQuery: `select t.*, $1 as param1 from t where t.f1 = $2 and t.f2 = '$2$3' and t.f3=$1`,
			expectedQueryParams: []any{
				"param1_value",
				"param2_value",
				[]int(nil),
			},
		},
		{
			desc: "test009_query_params_args_as_pointer",

			inputQuery: `select t.* from t where t.f1 = :param1`,
			inputQueryParams: &struct {
				Param1 string `db:"param1"`
			}{
				Param1: "param1_value",
			},

			expectedQuery: `select t.* from t where t.f1 = $1`,
			expectedQueryParams: []any{
				"param1_value",
			},
		},
		{
			desc: "test0010_query_params_args_as_nil_pointer",

			inputQuery: `select t.* from t where t.f1 = :param1`,
			inputQueryParams: func() any {
				//nolint:ineffassign,staticcheck
				queryParamsArgs := struct {
					Param1 string `db:"param1"`
				}{}
				queryParamsArgsPtr := &queryParamsArgs //nolint:ineffassign,staticcheck,wastedassign
				queryParamsArgsPtr = nil

				return queryParamsArgsPtr
			}(),

			expectedQuery: `select t.* from t where t.f1 = $1`,
			expectedQueryParams: []interface{}{
				nil,
			},
		},
		{
			desc: "test011_query_params_args_as_nil",
			CommonProps: stdcase.CommonProps{
				FailureProps: stdcase.FailureProps{
					ShouldFail: true,
				},
			},

			inputQuery:       `select t.* from t where t.f1 = :param1`,
			inputQueryParams: nil,
		},
		{
			desc: "test011_query_params_args_as_nil_and_no_params_in_query",

			inputQuery:       `select t.* from t where t.f1 = 1`,
			inputQueryParams: nil,

			expectedQuery:       `select t.* from t where t.f1 = 1`,
			expectedQueryParams: nil,
		},
		{
			desc: "test012_query_params_args_and_no_params_in_query",

			inputQuery: `select t.* from t where t.f1 = 1`,
			inputQueryParams: &struct {
				Param1 string `db:"param1"`
			}{
				Param1: "param1_value",
			},

			expectedQuery:       `select t.* from t where t.f1 = 1`,
			expectedQueryParams: nil,
		},
		{
			desc: "test013_query_param_name_qualifier_escaping",

			inputQuery: `select t.* from t where t.f1 = '::1' and ::`,
			inputQueryParams: &struct {
				Param1 string `db:"param1"`
			}{
				Param1: "param1_value",
			},

			expectedQuery:       `select t.* from t where t.f1 = ':1' and :`,
			expectedQueryParams: nil,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(
			tc.desc,
			func(t *testing.T) {
				tc.PreLog(t)

				actualQuery, actualQueryParams, err := db_q.ToPositionalSyntax(tc.inputQuery, db_q.DefaultParamQualifier, db_q.DefaultPositionalSyntaxVariant, tc.inputQueryParams)

				if tc.ShouldFail {
					require.Error(t, err, "error expected while executing test case")

					if tc.ShouldFailAsErr != nil {
						require.ErrorAs(t, err, tc.ShouldFailAsErr)

						return
					}

					if tc.ShouldFailIsErr != nil {
						require.ErrorIs(t, err, tc.ShouldFailIsErr)
					}

					return
				}

				require.NoError(t, err, "unexpected error while executing test case")

				require.Equal(t, tc.expectedQuery, actualQuery, "actual resulting query text doesn't match expected one")
				require.Equal(t, tc.expectedQueryParams, actualQueryParams, "actual resulting query parameter set dosn't match expected one")
			},
		)
	}
}
