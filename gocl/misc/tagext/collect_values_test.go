//go:build test_unit

package tagext_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/tagext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	clhlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

func Test_OfTagged(t *testing.T) {
	clhlp.Pre(t)

	const dbTagName = "db"

	type (
		testCase struct {
			InVal       any
			ExpectedOut map[string]any
			InTag       string

			stdcase.CommonProps
		}

		Ordinary1CaseStub struct {
			Prop1 int `db:"prop1"`
			Prop2 int `db:"prop2"`
		}

		Ordinary2CaseStub struct {
			Prop1Dup int `db:"prop1"` // such tag set intentionally
			Prop2Dup int `db:"prop2"` // such tag set intentionally
		}

		DuplicatedTagCaseStub struct {
			A int `db:"a"`
			B int `db:"a"`
		}

		EmptyTagCaseStub struct {
			C int `db:"c"`
			D int `db:""`
		}

		EmbeddedAndDuplicatedTagCaseStub struct {
			Ordinary1CaseStub

			E int `db:"e"`
			F int `db:"e"`
		}

		EmbeddedCaseStub struct {
			Ordinary1CaseStub

			G int `db:"g"`
			H int `db:"h"`
		}

		SkippedCaseStub struct {
			Ordinary1CaseStub
			EmbeddedCaseStub `db:"-"`

			K int `db:"k"`
			L int `db:"-"`
		}

		EmbeddedAsFieldValCaseStub struct {
			Ordinary1CaseStub `db:"obj"`
			Ordinary2CaseStub

			M int `db:"m"`
			N int `db:"n"`
		}

		UnexportedFieldCaseStub struct {
			Ordinary1CaseStub

			o Ordinary1CaseStub `db:"o"`
			p int               `db:""`
			Q int               `db:"q"`
			R int               `db:"r"`
		}

		ComplexCaseStub struct {
			*EmbeddedCaseStub
			t *EmbeddedCaseStub
			S *EmbeddedCaseStub `db:"obj2"`
			T Ordinary1CaseStub `db:"obj1"`
			U Ordinary1CaseStub `db:"-"`
			V int
		}
	)

	var (
		embeddedCaseStub1Ptr = &EmbeddedCaseStub{
			Ordinary1CaseStub: Ordinary1CaseStub{
				Prop1: 10,
				Prop2: 20,
			},
			G: 100,
			H: 200,
		}
		embeddedCaseStub2Ptr = &EmbeddedCaseStub{
			Ordinary1CaseStub: Ordinary1CaseStub{
				Prop1: 500,
				Prop2: 600,
			},
			G: 10000,
			H: 20000,
		}
		embeddedCaseStub3Ptr = &EmbeddedCaseStub{
			Ordinary1CaseStub: Ordinary1CaseStub{
				Prop1: 1,
				Prop2: 2,
			},
			G: 3,
			H: 4,
		}
	)

	testCases := map[string]testCase{
		"case00: simple": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			InVal: Ordinary1CaseStub{
				Prop1: 500,
				Prop2: 600,
			},
			InTag: dbTagName,
			ExpectedOut: map[string]any{
				"prop1": 500,
				"prop2": 600,
			},
		},
		"case01: ambiguity": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailIsErr: errs.ErrFieldTagValueAmbiguous,
				},
			},

			InVal: DuplicatedTagCaseStub{
				A: 5000,
				B: 6000,
			},
			InTag: dbTagName,
		},
		"case02: ambiguity+embedded": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailIsErr: errs.ErrFieldTagValueAmbiguous,
				},
			},

			InVal: EmbeddedAndDuplicatedTagCaseStub{
				Ordinary1CaseStub: Ordinary1CaseStub{
					Prop1: 100,
					Prop2: 200,
				},
				E: 1,
				F: 2,
			},
			InTag: dbTagName,
		},
		"case03: empty tag": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailIsErr: errs.ErrFieldTagValueEmpty,
				},
			},

			InVal: EmptyTagCaseStub{
				C: -20,
				D: -30,
			},
			InTag: dbTagName,
		},
		"case04: unexported": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			InVal: UnexportedFieldCaseStub{
				o: Ordinary1CaseStub{
					Prop1: 10,
					Prop2: 20,
				},
				Ordinary1CaseStub: Ordinary1CaseStub{
					Prop1: 30,
					Prop2: 40,
				},
				p: 1,
				Q: 2,
				R: 3,
			},
			InTag: dbTagName,
			ExpectedOut: map[string]any{
				"prop1": 30,
				"prop2": 40,
				"q":     2,
				"r":     3,
			},
		},
		"case05: hyphen tag": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			InVal: SkippedCaseStub{
				Ordinary1CaseStub: Ordinary1CaseStub{
					Prop1: 43,
					Prop2: 44,
				},
				EmbeddedCaseStub: EmbeddedCaseStub{
					Ordinary1CaseStub: Ordinary1CaseStub{
						Prop1: 10000,
						Prop2: 20000,
					},
					G: 15,
					H: 16,
				},
				K: 2500,
				L: 2501,
			},
			InTag: dbTagName,
			ExpectedOut: map[string]any{
				"prop1": 43,
				"prop2": 44,
				"k":     2500,
			},
		},
		"case06: embedded": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			InVal: EmbeddedCaseStub{
				Ordinary1CaseStub: Ordinary1CaseStub{
					Prop1: 1000,
					Prop2: 2000,
				},
				G: 100,
				H: 200,
			},
			InTag: dbTagName,
			ExpectedOut: map[string]any{
				"prop1": 1000,
				"prop2": 2000,
				"g":     100,
				"h":     200,
			},
		},
		"case07: embedded as obj": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			InVal: EmbeddedAsFieldValCaseStub{
				Ordinary1CaseStub: Ordinary1CaseStub{
					Prop1: 500,
					Prop2: 600,
				},
				Ordinary2CaseStub: Ordinary2CaseStub{
					Prop1Dup: 100,
					Prop2Dup: 200,
				},
				M: 99,
				N: 100,
			},
			InTag: dbTagName,
			ExpectedOut: map[string]any{
				"prop1": 100,
				"prop2": 200,
				"m":     99,
				"n":     100,
				"obj": Ordinary1CaseStub{
					Prop1: 500,
					Prop2: 600,
				},
			},
		},
		"case08: complex": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			InVal: ComplexCaseStub{
				t:                embeddedCaseStub1Ptr,
				EmbeddedCaseStub: embeddedCaseStub2Ptr,
				T: Ordinary1CaseStub{
					Prop1: 0,
					Prop2: 2,
				},
				S: embeddedCaseStub3Ptr,
				U: Ordinary1CaseStub{
					Prop1: 1000000,
					Prop2: 2000000,
				},
				V: -1,
			},
			InTag: dbTagName,
			ExpectedOut: map[string]any{
				"prop1": 500,
				"prop2": 600,
				"g":     10000,
				"h":     20000,
				"obj1": Ordinary1CaseStub{
					Prop1: 0,
					Prop2: 2,
				},
				"obj2": embeddedCaseStub3Ptr,
			},
		},
	}

	for tcName := range testCases {
		tc := testCases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				tc.PreLog(t)

				actualOut, actualErr := tagext.CollectValues(tc.InVal, tc.InTag)
				tc.RequireConformError(t, actualErr)

				if !tc.ShouldFail {
					require.Equal(t, tc.ExpectedOut, actualOut)
				}
			},
		)
	}
}
