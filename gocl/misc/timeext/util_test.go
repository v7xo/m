//go:build test_unit

package timeext_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cltime "rn.m.vo7.io/vo7io/m/gocl/misc/timeext"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

func Test_SplitPeriodBy(t *testing.T) {
	thelper.PreLog(t)

	type testCase struct {
		InputFrom time.Time
		InputTo   time.Time
		Expected  []cltime.Interval

		stdcase.CommonProps

		InputByMonth bool
	}

	cases := map[string]testCase{
		"case00/by month": {
			InputFrom:    time.Date(2024, 1, 7, 17, 25, 0, 0, time.UTC),
			InputTo:      time.Date(2024, 3, 7, 14, 25, 0, 0, time.UTC),
			InputByMonth: true,

			Expected: []cltime.Interval{
				{
					From: time.Date(2024, 1, 7, 17, 25, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 31, 23, 59, 59, int(time.Microsecond*999999), time.UTC),
				},
				{
					From: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 2, 29, 23, 59, 59, int(time.Microsecond*999999), time.UTC),
				},
				{
					From: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 3, 7, 23, 59, 59, int(time.Microsecond*999999), time.UTC),
				},
			},
		},
		"case01/by month": {
			InputFrom:    time.Date(2024, 1, 7, 17, 25, 0, 0, time.UTC),
			InputTo:      time.Date(2024, 1, 20, 14, 25, 0, 0, time.UTC),
			InputByMonth: true,

			Expected: []cltime.Interval{
				{
					From: time.Date(2024, 1, 7, 17, 25, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 20, 23, 59, 59, int(time.Microsecond*999999), time.UTC),
				},
			},
		},
		"case02/by day": {
			InputFrom:    time.Date(2024, 1, 30, 2, 25, 0, 0, time.UTC),
			InputTo:      time.Date(2024, 2, 2, 17, 0, 0, 0, time.UTC),
			InputByMonth: false,

			Expected: []cltime.Interval{
				{
					From: time.Date(2024, 1, 30, 2, 25, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 30, 23, 59, 59, int(time.Microsecond*999999), time.UTC),
				},
				{
					From: time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 31, 23, 59, 59, int(time.Microsecond*999999), time.UTC),
				},
				{
					From: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 2, 1, 23, 59, 59, int(time.Microsecond*999999), time.UTC),
				},
				{
					From: time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC),
					To:   time.Date(2024, 2, 2, 23, 59, 59, int(time.Microsecond*999999), time.UTC),
				},
			},
		},
		"case03/by day": {
			InputFrom:    time.Date(2024, 1, 30, 2, 25, 0, 0, time.UTC),
			InputTo:      time.Date(2024, 1, 30, 23, 17, 0, 0, time.UTC),
			InputByMonth: false,

			Expected: []cltime.Interval{
				{
					From: time.Date(2024, 1, 30, 2, 25, 0, 0, time.UTC),
					To:   time.Date(2024, 1, 30, 23, 59, 59, int(time.Microsecond*999999), time.UTC),
				},
			},
		},
	}

	for name := range cases {
		tc := cases[name]

		t.Run(
			name,
			func(t *testing.T) {
				var actual []cltime.Interval

				if tc.InputByMonth {
					actual = cltime.SplitByMonth(tc.InputFrom, tc.InputTo, time.UTC)
				} else {
					actual = cltime.SplitByDay(tc.InputFrom, tc.InputTo, time.UTC)
				}

				require.Equal(t, tc.Expected, actual, "unexpected mismatch of result")
			},
		)
	}
}
