//go:build test_unit

package validation_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

func Test_ValidationTag(t *testing.T) {
	hlp.PreLog(t)

	type (
		testCase struct {
			val  any
			desc string

			stdcase.CommonProps
		}

		requirePropWithLen2 struct {
			Prop1 string `validate:"len=2"`
		}

		validatableWithDiveTag struct {
			Field []requirePropWithLen2 `validate:"dive"`
		}
		validatableWithoutDiveTag struct {
			Field []requirePropWithLen2
		}
	)

	testCases := []testCase{
		{
			desc: "001/tag-required-without",

			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			val: struct {
				FieldA string `validate:"required_without=FieldB"`
				FieldB string `validate:"required_without=FieldA"`
			}{
				FieldA: "",
				FieldB: "value",
			},
		},
		{
			desc: "002/tag-required-without",

			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailAsErr: &validation.Error{},
				},
			},

			val: struct {
				FieldA string `validate:"required_without=FieldB"`
				FieldB string `validate:"required_without=FieldA"`
			}{
				FieldA: "",
				FieldB: "",
			},
		},
		{
			desc: "003/tag-required-without",

			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			val: struct {
				FieldA string `validate:"required_without=FieldB"`
				FieldB string `validate:"required_without=FieldA"`
			}{
				FieldA: "A",
				FieldB: "B",
			},
		},
		{
			desc: "004/only-one-of-field-present",

			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailAsErr: &validation.Error{},
				},
			},

			val: struct {
				FieldA string `validate:"excluded_with=FieldB,required_without=FieldB"`
				FieldB string `validate:"excluded_with=FieldA,required_without=FieldA"`
			}{
				FieldA: "",
				FieldB: "",
			},
		},
		{
			desc: "005/only-one-of-field-present",

			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			val: struct {
				FieldA string `validate:"excluded_with=FieldB,required_without=FieldB"`
				FieldB string `validate:"excluded_with=FieldA,required_without=FieldA"`
			}{
				FieldA: "A",
				FieldB: "",
			},
		},
		{
			desc: "006/only-one-of-field-present",

			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailAsErr: &validation.Error{},
				},
			},

			val: struct {
				FieldA string `validate:"excluded_with=FieldB,required_without=FieldB"`
				FieldB string `validate:"excluded_with=FieldA,required_without=FieldA"`
			}{
				FieldA: "A",
				FieldB: "B",
			},
		},
		{
			desc: "007/only-one-of-field-present-conditionally",

			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			val: struct {
				FieldA string `validate:"excluded_with=FieldB,required_without=FieldB|required_with=FieldC"`
				FieldB string `validate:"excluded_with=FieldA,required_without=FieldA|required_with=FieldC"`
				FieldC bool
			}{
				FieldA: "",
				FieldB: "",
				FieldC: false,
			},
		},
		{
			desc: "008/only-one-of-field-present-conditionally",

			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailAsErr: &validation.Error{},
				},
			},

			val: struct {
				FieldA string `validate:"excluded_with=FieldB,required_without=FieldB|required_with=FieldC"`
				FieldB string `validate:"excluded_with=FieldA,required_without=FieldA|required_with=FieldC"`
				FieldC bool
			}{
				FieldA: "",
				FieldB: "",
				FieldC: true,
			},
		},
		{
			desc: "009/only-one-of-field-present-conditionally",

			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			val: struct {
				FieldA string `validate:"excluded_with=FieldB,required_without=FieldB|required_with=FieldC"`
				FieldB string `validate:"excluded_with=FieldA,required_without=FieldA|required_with=FieldC"`
				FieldC bool
			}{
				FieldA: "",
				FieldB: "FieldB",
				FieldC: true,
			},
		},
		{
			desc: "010/only-one-of-field-present-conditionally",

			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			val: struct {
				FieldA string `validate:"excluded_with=FieldB,required_without=FieldB|required_with=FieldC"`
				FieldB string `validate:"excluded_with=FieldA,required_without=FieldA|required_with=FieldC"`
				FieldC bool
			}{
				FieldA: "FieldA",
				FieldB: "",
				FieldC: true,
			},
		},
		{
			desc: "011/only-one-of-field-present-conditionally",

			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailAsErr: &validation.Error{},
				},
			},

			val: struct {
				FieldA string `validate:"excluded_with=FieldB,required_without=FieldB|required_with=FieldC"`
				FieldB string `validate:"excluded_with=FieldA,required_without=FieldA|required_with=FieldC"`
				FieldC bool
			}{
				FieldA: "FieldA",
				FieldB: "FieldB",
				FieldC: true,
			},
		},
		{
			desc: "012/only-one-of-field-present-conditionally",

			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailAsErr: &validation.Error{},
				},
			},

			val: struct {
				FieldA string `validate:"excluded_with=FieldB,required_without=FieldB|required_with=FieldC"`
				FieldB string `validate:"excluded_with=FieldA,required_without=FieldA|required_with=FieldC"`
				FieldC bool
			}{
				FieldA: "FieldA",
				FieldB: "FieldB",
				FieldC: false,
			},
		},
		{
			desc: "013/dive-tag/slice/tag-declared",

			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,

				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailAsErr: &validation.Error{},
				},
			},

			val: validatableWithDiveTag{
				Field: []requirePropWithLen2{
					{Prop1: "AA"},
					{Prop1: "A"},
					{Prop1: "AAA"},
					{Prop1: ""},
				},
			},
		},
		{
			desc: "014/dive-tag/slice/tag-not-declared",

			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,

				FailureProps: stdcase.FailureProps{
					ShouldFail: false,
				},
			},

			val: validatableWithoutDiveTag{
				Field: []requirePropWithLen2{
					{Prop1: "AA"},
					{Prop1: "A"},
					{Prop1: "AAA"},
					{Prop1: ""},
				},
			},
		},
		{
			desc: "015/unexported-field-ignored",

			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			val: struct {
				prop1 string `validate:"len=2"`
			}{
				prop1: "AAA",
			},
		},
	}

	validator := validation.DefaultValidator()
	require.NotNil(t, validator)

	for i := range testCases {
		tc := testCases[i]

		t.Run(
			tc.desc,
			func(t *testing.T) {
				tc.PreLog(t)

				err := validator.Struct(tc.val)

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
			},
		)
	}
}
