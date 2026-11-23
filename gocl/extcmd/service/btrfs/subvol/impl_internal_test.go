//go:build test_unit

package subvol

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

func Test_FromBtrfsSubvolListOutputReader(t *testing.T) {
	hlp.Pre(t)

	type testCase struct {
		stdcase.CommonProps

		input          string
		expectedOutput SummaryList
	}

	cases := map[string]testCase{
		"case00/happy": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input: `gen 1222728 ID 901 cgen 1222728 parent 5 top level 5 parent_uuid c02f3c23-6fc2-2849-8566-5d35bf8b63db received_uuid -                                    uuid 58c38160-ccd9-bb4d-9f9a-334a1a945407 path snapshot/kili__app-data__radon_gitlab-log_2026-03-03T00_20_27+00_00`,
			expectedOutput: SummaryList{
				Summary{
					Generation:         1222728,
					ID:                 901,
					CreationGeneration: 1222728,
					ParentID:           5,
					Path:               "snapshot/kili__app-data__radon_gitlab-log_2026-03-03T00_20_27+00_00",
					ReceivedUUID:       uuid.Nil,
					ParentUUID:         uuid.MustParse("c02f3c23-6fc2-2849-8566-5d35bf8b63db"),
					TopLevelID:         5,
					UUID:               uuid.MustParse("58c38160-ccd9-bb4d-9f9a-334a1a945407"),
				},
			},
		},
		"case01/happy_multiline": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input: `gen 1222728 ID 901 cgen 1222728 parent 5 top level 5 parent_uuid c02f3c23-6fc2-2849-8566-5d35bf8b63db received_uuid -                                    uuid 58c38160-ccd9-bb4d-9f9a-334a1a945407 path snapshot/kili__app-data__radon_gitlab-log_2026-03-03T00_20_27+00_00
ID 259 gen 1159224 cgen 111 parent 256 top level 256 parent_uuid -                                    received_uuid -                                    uuid 14a93b49-cd09-304f-94f7-e33cccfa2d56 path root/var/lib/portables`,
			expectedOutput: SummaryList{
				Summary{
					Generation:         1222728,
					ID:                 901,
					CreationGeneration: 1222728,
					ParentID:           5,
					Path:               "snapshot/kili__app-data__radon_gitlab-log_2026-03-03T00_20_27+00_00",
					ParentUUID:         uuid.MustParse("c02f3c23-6fc2-2849-8566-5d35bf8b63db"),
					ReceivedUUID:       uuid.Nil,
					TopLevelID:         5,
					UUID:               uuid.MustParse("58c38160-ccd9-bb4d-9f9a-334a1a945407"),
				},
				Summary{
					Generation:         1159224,
					ID:                 259,
					CreationGeneration: 111,
					ParentID:           256,
					Path:               "root/var/lib/portables",
					ParentUUID:         uuid.Nil,
					ReceivedUUID:       uuid.Nil,
					TopLevelID:         256,
					UUID:               uuid.MustParse("14a93b49-cd09-304f-94f7-e33cccfa2d56"),
				},
			},
		},
		"case02/fail": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
				FailureProps: stdcase.FailureProps{
					ShouldFail: true,
				},
			},

			input: `ID 901 cgen 1222728 parent 5 top level 5 parent_uuid c02f3c23-6fc2-2849-8566-5d35bf8b63db received_uuid -                                    uuid 58c38160-ccd9-bb4d-9f9a-334a1a945407 path snapshot/kili__app-data__radon_gitlab-log_2026-03-03T00_20_27+00_00`,
		},
		"case03/happy": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			input: `ID 256 gen 7 cgen 7 parent 5 top level 5 parent_uuid -                                    received_uuid -                                    uuid 179092c9-5f08-6b47-92fd-56690811f39d path subvol1
ID 257 gen 7 cgen 7 parent 5 top level 5 parent_uuid -                                    received_uuid -                                    uuid 020d0a30-9b41-0246-a4cc-6e6c7d574477 path subvol2`,
			expectedOutput: SummaryList{
				Summary{
					Generation:         7,
					ID:                 256,
					CreationGeneration: 7,
					ParentID:           5,
					Path:               "subvol1",
					ParentUUID:         uuid.Nil,
					ReceivedUUID:       uuid.Nil,
					TopLevelID:         5,
					UUID:               uuid.MustParse("179092c9-5f08-6b47-92fd-56690811f39d"),
				},
				Summary{
					Generation:         7,
					ID:                 257,
					CreationGeneration: 7,
					ParentID:           5,
					Path:               "subvol2",
					ParentUUID:         uuid.Nil,
					ReceivedUUID:       uuid.Nil,
					TopLevelID:         5,
					UUID:               uuid.MustParse("020d0a30-9b41-0246-a4cc-6e6c7d574477"),
				},
			},
		},
	}

	for tcName := range cases {
		tc := cases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				tc.Pre(t)

				var actual SummaryList

				err := actual.FromBtrfsSubvolumeListOutputReader(bytes.NewBufferString(tc.input))
				tc.RequireConformError(t, err)

				if tc.ShouldFail {
					return
				}

				require.ElementsMatch(t, actual, tc.expectedOutput)
			},
		)
	}
}
