//go:build test_unit

package netipext_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/netipext"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

func Test_ToPrefixes(t *testing.T) {
	thelper.Pre(t)

	type testCase struct {
		stdcase.CommonProps

		start uint32
		end   uint32

		expectedCIDR []string
	}

	cases := map[string]testCase{
		"case00": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			start: 0,
			end:   0,

			expectedCIDR: []string{"0.0.0.0/32"},
		},
		"case01": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			start: 0,
			end:   255,

			expectedCIDR: []string{"0.0.0.0/24"},
		},
		"case02": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			start: 0,
			end:   math.MaxUint32,

			expectedCIDR: []string{"0.0.0.0/0"},
		},
		"case03": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			start: math.MaxUint32,
			end:   math.MaxUint32,

			expectedCIDR: []string{"255.255.255.255/32"},
		},
	}

	for tcName := range cases {
		tc := cases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				tc.Pre(t)

				block := netipext.IP4Block{}
				err := block.FromUint32(tc.start, tc.end)
				require.NoError(t, err)

				actualPrefixes := block.ToPrefixes()

				actualStrings := make([]string, len(actualPrefixes))
				for i, v := range actualPrefixes {
					actualStrings[i] = v.String()
				}

				require.ElementsMatch(t, tc.expectedCIDR, actualStrings)
			},
		)
	}
}
