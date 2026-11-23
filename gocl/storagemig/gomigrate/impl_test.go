//go:build test_postgres

package gomigrate_test

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/storage/query"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi/storageapi_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/errs"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/gomigrate/testdata/migrations_dummy"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/gomigrate/testdata/migrations_edit1"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/gomigrate/testdata/migrations_edit2"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/gomigrate/testdata/migrations_edit3_same_as_edit2"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/gomigrate/testdata/migrations_edit4"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/storagemig_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/types"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
)

const (
	DeployJob2ID = "job-id-02"
	DeployJob3ID = "job-id-03"
)

var (
	publicSchemeTablesNamesGetQuery = query.MustNewNamed(
		"migration_test_public_scheme_tables_names_get",
		`select "table_name" from "information_schema"."tables" where "table_schema" = 'public'`,
	)

	appReleaseFn = fmt.Sprintf("Go tests deploy app release, %s", time.Now().Format(time.RFC3339))

	migrations_edit1_DeployFingerprint     = types.DeployFingerprint(-4712434606456645726)
	migrations_edit2_DeployFingerprint     = types.DeployFingerprint(-7242059287795885910)
	migrations_edit2Job2_DeployFingerprint = types.DeployFingerprint(2753471794662736487)
	migrations_edit2Job3_DeployFingerprint = types.DeployFingerprint(-6139101687439953432)
	migrations_edit4_DeployFingerprint     = types.DeployFingerprint(5861052401568564156)
)

func Test_MigrationInitialize(t *testing.T) {
	thelper.PreLog(t)

	type testCase struct {
		stdcase.MigratedStorageProps
		stdcase.CommonProps

		expectedInDBStatus  types.StorageStatus
		expectedOutDBStatus types.StorageStatus

		DryRun bool
	}

	var (
		mig1Sources = migrations_edit1.Sources
		mig1UpSteps = hlpTestMigrations1UpStepsOpen(t)
	)

	testCases := map[string]testCase{
		"case00: dry-run initializtion": {
			DryRun: true,
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},
			MigratedStorageProps: stdcase.MigratedStorageProps{
				Sources1: mig1Sources,
			},

			expectedInDBStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,

				NotInitialized: true,
			},
			expectedOutDBStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,

				NotInitialized: true,
			},
		},
		"case01: initializtion": {
			DryRun: false,
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},
			MigratedStorageProps: stdcase.MigratedStorageProps{
				Sources1: mig1Sources,
			},

			expectedInDBStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,

				NotInitialized: true,
			},
			expectedOutDBStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,
			},
		},
	}

	for tcName := range testCases {
		tc := testCases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				ctx := tc.Pre(t)

				api, _, _ := storageapi_thelper.NewPgDB[storageapi.SQLAPI](t, ctx)
				migSvc := storagemig_thelper.NewService(t, ctx, tc.Sources1, nil, api)

				hlpRequireDBStatusWithService(t, ctx, migSvc, tc.expectedInDBStatus)

				err := migSvc.Initialize(ctx, tc.DryRun)
				tc.RequireConformError(t, err)

				hlpRequireDBStatusWithService(t, ctx, migSvc, tc.expectedOutDBStatus)
			},
		)
	}
}

func Test_MigrationUpFromScratch(t *testing.T) {
	thelper.PreLog(t)

	type testCase struct {
		stdcase.MigratedStorageProps
		stdcase.CommonProps

		expectedOutDBTablesNames []string
		expectedInDBStatus       types.StorageStatus
		expectedOutDBStatus      types.StorageStatus

		DryRun bool
	}

	var (
		mig1Sources = migrations_edit1.Sources
		mig1UpSteps = hlpTestMigrations1UpStepsOpen(t)
	)

	testCases := map[string]testCase{
		"case00: up": {
			DryRun: false,
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},
			MigratedStorageProps: stdcase.MigratedStorageProps{
				Sources1: mig1Sources,
			},

			expectedInDBStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,

				NotInitialized: true,
			},
			expectedOutDBStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					ApplyOrder:        2,
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit1_DeployFingerprint,
					Num:               2,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  1,
					Fingerprint: migrations_edit1_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        1,
							ApplyOrder: 1,
						},
						{
							Num:        2,
							ApplyOrder: 2,
						},
					},
				},
			},
			expectedOutDBTablesNames: []string{
				`test1`,
				`test2`,
			},
		},
		"case01: dry-run up": {
			DryRun: true,
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},
			MigratedStorageProps: stdcase.MigratedStorageProps{
				Sources1: mig1Sources,
			},

			expectedInDBStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,

				NotInitialized: true,
			},
			expectedOutDBStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,

				NotInitialized: true,
			},
			expectedOutDBTablesNames: []string{},
		},
	}

	for tcName := range testCases {
		tc := testCases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				ctx := tc.Pre(t)

				api, _, _ := storageapi_thelper.NewPgDB[storageapi.SQLAPI](t, ctx)
				migSvc := storagemig_thelper.NewService(t, ctx, tc.Sources1, nil, api)

				hlpRequireDBStatusWithService(t, ctx, migSvc, tc.expectedInDBStatus)

				err := migSvc.Up(ctx, tc.DryRun)
				tc.RequireConformError(t, err)

				hlpRequireDBStatusWithService(t, ctx, migSvc, tc.expectedOutDBStatus)
				hlpRequirePublicSchemeTablesNamesWithDBAPI(t, ctx, api, tc.expectedOutDBTablesNames)
			},
		)
	}
}

func Test_MigrationDownFromScratch(t *testing.T) {
	thelper.PreLog(t)

	type testCase struct {
		stdcase.MigratedStorageProps
		stdcase.CommonProps

		expectedInDBStatus  types.StorageStatus
		expectedOutDBStatus types.StorageStatus

		DryRun bool
	}

	var (
		mig1Sources = migrations_edit1.Sources
		mig1UpSteps = hlpTestMigrations1UpStepsOpen(t)
	)

	testCases := map[string]testCase{
		"case00: down": {
			DryRun: false,
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailIsErr: errs.ErrNotInitialized,
				},
			},
			MigratedStorageProps: stdcase.MigratedStorageProps{
				Sources1: mig1Sources,
			},

			expectedInDBStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,

				NotInitialized: true,
			},
			expectedOutDBStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,

				NotInitialized: true,
			},
		},
	}

	for tcName := range testCases {
		tc := testCases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				ctx := tc.Pre(t)

				api, _, _ := storageapi_thelper.NewPgDB[storageapi.SQLAPI](t, ctx)
				migSvc := storagemig_thelper.NewService(t, ctx, tc.Sources1, nil, api)

				hlpRequireDBStatusWithService(t, ctx, migSvc, tc.expectedInDBStatus)

				err := migSvc.Down(ctx, tc.DryRun)
				tc.RequireConformError(t, err)

				hlpRequireDBStatusWithService(t, ctx, migSvc, tc.expectedOutDBStatus)
			},
		)
	}
}

func Test_MigrationZeroUp(t *testing.T) {
	thelper.PreLog(t)

	type testCase struct {
		stdcase.MigratedStorageProps
		stdcase.CommonProps

		expectedOutDBTablesNames []string

		expectedInDBStatus  types.StorageStatus
		expectedOutDBStatus types.StorageStatus

		DryRun bool
	}

	var (
		mig1Sources = migrations_edit1.Sources
		mig1UpSteps = hlpTestMigrations1UpStepsOpen(t)

		_ = mig1Sources
		_ = mig1UpSteps
	)

	testCases := map[string]testCase{
		"case00: up": {
			DryRun: false,
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},
			MigratedStorageProps: stdcase.MigratedStorageProps{
				Sources1: mig1Sources,
			},

			expectedInDBStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit1_DeployFingerprint,
					Num:               2,
					ApplyOrder:        2,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  1,
					Fingerprint: migrations_edit1_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        1,
							ApplyOrder: 1,
						},
						{
							Num:        2,
							ApplyOrder: 2,
						},
					},
				},
			},
			expectedOutDBStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit1_DeployFingerprint,
					Num:               2,
					ApplyOrder:        2,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  1,
					Fingerprint: migrations_edit1_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        1,
							ApplyOrder: 1,
						},
						{
							Num:        2,
							ApplyOrder: 2,
						},
					},
				},
			},
			expectedOutDBTablesNames: []string{
				`test1`,
				`test2`,
			},
		},
	}

	for tcName := range testCases {
		tc := testCases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				ctx := tc.Pre(t)

				api, _, _ := storageapi_thelper.NewPgDB[storageapi.SQLAPI](t, ctx)
				storagemig_thelper.Up(t, ctx, tc.Sources1, nil, api)

				// apply empty migrations set.
				//
				{
					migSvc := storagemig_thelper.NewService(t, ctx, tc.Sources1, nil, api)

					hlpRequireDBStatusWithService(t, ctx, migSvc, tc.expectedInDBStatus)

					err := migSvc.Up(ctx, tc.DryRun)
					tc.RequireConformError(t, err)

					hlpRequireDBStatusWithService(t, ctx, migSvc, tc.expectedOutDBStatus)
					hlpRequirePublicSchemeTablesNamesWithDBAPI(t, ctx, api, tc.expectedOutDBTablesNames)
				}
			},
		)
	}
}

func Test_MigrationUpDown(t *testing.T) {
	thelper.PreLog(t)

	type (
		migrationOpCode int

		migrationOp struct {
			Sources     fs.FS
			DeployJobID string
			stdcase.FailureProps
			Code   migrationOpCode
			DryRun bool
		}

		testCase struct {
			stdcase.CommonProps

			ExpectedInStatus_UseSource  fs.FS
			ExpectedOutStatus_UseSource fs.FS
			MigrationOp                 []migrationOp

			ExpectedOutDBTablesNames []string
			ExpectedInStatus         types.StorageStatus
			ExpectedOutStatus        types.StorageStatus
		}
	)

	const (
		up migrationOpCode = iota
		down
	)

	var (
		migs1Sources     = migrations_edit1.Sources
		migs2Sources     = migrations_edit2.Sources
		migs3Sources     = migrations_edit3_same_as_edit2.Sources
		migs4Sources     = migrations_edit4.Sources
		migsDummySources = migrations_dummy.Sources

		mig1UpSteps = hlpTestMigrations1UpStepsOpen(t)

		_ = migs1Sources
		_ = migs2Sources
		_ = migs3Sources
		_ = migs4Sources

		_ = mig1UpSteps
	)

	testCases := map[string]testCase{
		"case00/up(migration1)->down(migration1)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migs1Sources,
			ExpectedOutStatus_UseSource: migs1Sources,
			MigrationOp: []migrationOp{
				{
					Code:    up,
					Sources: migs1Sources,
				},
				{
					Code:    down,
					Sources: migs1Sources,
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,

				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,
			},
			ExpectedOutDBTablesNames: []string{},
		},
		"case01/down(migration1)->up(migration1)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migs1Sources,
			ExpectedOutStatus_UseSource: migs1Sources,
			MigrationOp: []migrationOp{
				{
					Code: down,
					FailureProps: stdcase.FailureProps{
						ShouldFail:      true,
						ShouldFailIsErr: errs.ErrNotInitialized,
					},
					Sources: migs1Sources,
				},
				{
					Code:    up,
					Sources: migs1Sources,
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,

				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit1_DeployFingerprint,
					Num:               2,
					ApplyOrder:        2,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  1,
					Fingerprint: migrations_edit1_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        1,
							ApplyOrder: 1,
						},
						{
							Num:        2,
							ApplyOrder: 2,
						},
					},
				},
			},
			ExpectedOutDBTablesNames: []string{
				`test1`,
				`test2`,
			},
		},
		"case02/up(migration1)->up(migration1)->up(migration1)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migs1Sources,
			ExpectedOutStatus_UseSource: migs1Sources,
			MigrationOp: []migrationOp{
				{
					Code:    up,
					Sources: migs1Sources,
				},
				{
					Code:    up,
					Sources: migs1Sources,
				},
				{
					Code:    up,
					Sources: migs1Sources,
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,

				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit1_DeployFingerprint,
					Num:               2,
					ApplyOrder:        2,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  1,
					Fingerprint: migrations_edit1_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        1,
							ApplyOrder: 1,
						},
						{
							Num:        2,
							ApplyOrder: 2,
						},
					},
				},
			},
			ExpectedOutDBTablesNames: []string{
				`test1`,
				`test2`,
			},
		},
		"case03/up(migration1)->down(migration1)->down(migration1)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migs1Sources,
			ExpectedOutStatus_UseSource: migs1Sources,
			MigrationOp: []migrationOp{
				{
					Code:    up,
					Sources: migs1Sources,
				},
				{
					Code:    down,
					Sources: migs1Sources,
				},
				{
					Code:    down,
					Sources: migs1Sources,
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,

				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,
			},
			ExpectedOutDBTablesNames: []string{},
		},
		"case04/up(migration1)->down(migration1)->down(migration1)->up(migration1)->up(migration1)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migs1Sources,
			ExpectedOutStatus_UseSource: migs1Sources,
			MigrationOp: []migrationOp{
				{
					Code:    up,
					Sources: migs1Sources,
				},
				{
					Code:    down,
					Sources: migs1Sources,
				},
				{
					Code:    down,
					Sources: migs1Sources,
				},
				{
					Code:    up,
					Sources: migs1Sources,
				},
				{
					Code:    up,
					Sources: migs1Sources,
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,

				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit1_DeployFingerprint,
					Num:               2,
					ApplyOrder:        4,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  1,
					Fingerprint: migrations_edit1_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        1,
							ApplyOrder: 3,
						},
						{
							Num:        2,
							ApplyOrder: 4,
						},
					},
				},
			},
			ExpectedOutDBTablesNames: []string{
				`test1`,
				`test2`,
			},
		},
		"case05/up(migration1)->up(migration2)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migs1Sources,
			ExpectedOutStatus_UseSource: migs1Sources,
			MigrationOp: []migrationOp{
				{
					Code:    up,
					Sources: migs1Sources,
				},
				{
					Code:    up,
					Sources: migs2Sources,
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,

				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit2_DeployFingerprint,
					Num:               3,
					ApplyOrder:        3,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  2,
					Fingerprint: migrations_edit2_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        3,
							ApplyOrder: 3,
						},
					},
				},
			},
			ExpectedOutDBTablesNames: []string{
				`test1`,
				`test2`,
				`test3`,
			},
		},
		"case06/up(migration1)->up(migration2)->up(migration3_same_as_migration2)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migs1Sources,
			ExpectedOutStatus_UseSource: migs1Sources,
			MigrationOp: []migrationOp{
				{
					Code:    up,
					Sources: migs1Sources,
				},
				{
					Code:    up,
					Sources: migs2Sources,
				},
				{
					Code:    up,
					Sources: migs3Sources,
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
				UpgradableTo: mig1UpSteps,

				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit2_DeployFingerprint,
					Num:               3,
					ApplyOrder:        3,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  2,
					Fingerprint: migrations_edit2_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        3,
							ApplyOrder: 3,
						},
					},
				},
			},
			ExpectedOutDBTablesNames: []string{
				`test1`,
				`test2`,
				`test3`,
			},
		},
		"case07/up(migration4)->up(migration2)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migsDummySources,
			ExpectedOutStatus_UseSource: migsDummySources,
			MigrationOp: []migrationOp{
				{
					Code:    up,
					Sources: migs4Sources,
				},
				{
					Code:    up,
					Sources: migs2Sources,
					FailureProps: stdcase.FailureProps{
						ShouldFail:      true,
						ShouldFailIsErr: errs.ErrCurrentDeployBehindLatestVersion,
					},
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},

				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit4_DeployFingerprint,
					Num:               4,
					ApplyOrder:        4,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  1,
					Fingerprint: migrations_edit4_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        1,
							ApplyOrder: 1,
						},
						{
							Num:        2,
							ApplyOrder: 2,
						},
						{
							Num:        3,
							ApplyOrder: 3,
						},
						{
							Num:        4,
							ApplyOrder: 4,
						},
					},
				},
			},
			ExpectedOutDBTablesNames: []string{
				`test1`,
				`test2`,
				`test3`,
				`test4`,
			},
		},
		"case08/up(migration4)->down(migration2)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migsDummySources,
			ExpectedOutStatus_UseSource: migsDummySources,
			MigrationOp: []migrationOp{
				{
					Code:    up,
					Sources: migs4Sources,
				},
				{
					Code:    down,
					Sources: migs2Sources,
					FailureProps: stdcase.FailureProps{
						ShouldFail:      true,
						ShouldFailIsErr: errs.ErrCurrentDeployNotRegistered,
					},
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},

				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit4_DeployFingerprint,
					Num:               4,
					ApplyOrder:        4,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  1,
					Fingerprint: migrations_edit4_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        1,
							ApplyOrder: 1,
						},
						{
							Num:        2,
							ApplyOrder: 2,
						},
						{
							Num:        3,
							ApplyOrder: 3,
						},
						{
							Num:        4,
							ApplyOrder: 4,
						},
					},
				},
			},
			ExpectedOutDBTablesNames: []string{
				`test1`,
				`test2`,
				`test3`,
				`test4`,
			},
		},
		"case09/up(migration2)->up(migration4)->down(migration2)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migsDummySources,
			ExpectedOutStatus_UseSource: migsDummySources,
			MigrationOp: []migrationOp{
				{
					Code:    up,
					Sources: migs2Sources,
				},
				{
					Code:    up,
					Sources: migs4Sources,
				},
				{
					Code:    down,
					Sources: migs2Sources,
					FailureProps: stdcase.FailureProps{
						ShouldFail:      true,
						ShouldFailIsErr: errs.ErrCurrentDeployNotLatest,
					},
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},

				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit4_DeployFingerprint,
					Num:               4,
					ApplyOrder:        4,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  2,
					Fingerprint: migrations_edit4_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        4,
							ApplyOrder: 4,
						},
					},
				},
			},
			ExpectedOutDBTablesNames: []string{
				`test1`,
				`test2`,
				`test3`,
				`test4`,
			},
		},
		"case10/up(migration2)->up(migration4)->down(migration4)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migsDummySources,
			ExpectedOutStatus_UseSource: migsDummySources,
			MigrationOp: []migrationOp{
				{
					Code:    up,
					Sources: migs2Sources,
				},
				{
					Code:    up,
					Sources: migs4Sources,
				},
				{
					Code:    down,
					Sources: migs4Sources,
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},

				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit2_DeployFingerprint,
					Num:               3,
					ApplyOrder:        3,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  1,
					Fingerprint: migrations_edit2_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        1,
							ApplyOrder: 1,
						},
						{
							Num:        2,
							ApplyOrder: 2,
						},
						{
							Num:        3,
							ApplyOrder: 3,
						},
					},
				},
			},
			ExpectedOutDBTablesNames: []string{
				`test1`,
				`test2`,
				`test3`,
			},
		},
		"case11/up(migration2)->up(migration4)->down(migration4)->down(migration4)->down(migration2)->down(migration2)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migsDummySources,
			ExpectedOutStatus_UseSource: migsDummySources,
			MigrationOp: []migrationOp{
				{
					Code:    up,
					Sources: migs2Sources,
				},
				{
					Code:    up,
					Sources: migs4Sources,
				},
				{
					Code:    down,
					Sources: migs4Sources,
				},
				{
					Code:    down,
					Sources: migs4Sources,
				},
				{
					Code:    down,
					Sources: migs2Sources,
				},
				{
					Code:    down,
					Sources: migs2Sources,
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},

				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},
			},
			ExpectedOutDBTablesNames: []string{},
		},
		"case12/up(migration2)->up(migration4)->down(migration4)->down(migration4)->down(migration2)->down(migration2)->up(migration4)->up(migration2)->down(migration2)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migsDummySources,
			ExpectedOutStatus_UseSource: migsDummySources,
			MigrationOp: []migrationOp{
				{
					Code:    up,
					Sources: migs2Sources,
				},
				{
					Code:    up,
					Sources: migs4Sources,
				},
				{
					Code:    down,
					Sources: migs4Sources,
				},
				{
					Code:    down,
					Sources: migs4Sources,
				},
				{
					Code:    down,
					Sources: migs2Sources,
				},
				{
					Code:    down,
					Sources: migs2Sources,
				},
				{
					Code:    up,
					Sources: migs4Sources,
				},
				{
					Code:    up,
					Sources: migs2Sources,
					FailureProps: stdcase.FailureProps{
						ShouldFail:      true,
						ShouldFailIsErr: errs.ErrCurrentDeployBehindLatestVersion,
					},
				},
				{
					Code:    down,
					Sources: migs2Sources,
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},

				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit4_DeployFingerprint,
					Num:               4,
					ApplyOrder:        8,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  2,
					Fingerprint: migrations_edit4_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        1,
							ApplyOrder: 5,
						},
						{
							Num:        2,
							ApplyOrder: 6,
						},
						{
							Num:        3,
							ApplyOrder: 7,
						},
						{
							Num:        4,
							ApplyOrder: 8,
						},
					},
				},
			},
			ExpectedOutDBTablesNames: []string{
				`test1`,
				`test2`,
				`test3`,
				`test4`,
			},
		},
		"case13/up(migration2,job2)->up(migration2,job3)->down(migration2,job3)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migsDummySources,
			ExpectedOutStatus_UseSource: migsDummySources,
			MigrationOp: []migrationOp{
				{
					Code:        up,
					Sources:     migs2Sources,
					DeployJobID: DeployJob2ID,
				},
				{
					Code:        up,
					Sources:     migs2Sources,
					DeployJobID: DeployJob3ID,
				},
				{
					Code:        down,
					Sources:     migs2Sources,
					DeployJobID: DeployJob3ID,
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},

				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit2Job2_DeployFingerprint,
					Num:               3,
					ApplyOrder:        3,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  1,
					Fingerprint: migrations_edit2Job2_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        1,
							ApplyOrder: 1,
						},
						{
							Num:        2,
							ApplyOrder: 2,
						},
						{
							Num:        3,
							ApplyOrder: 3,
						},
					},
				},
			},
			ExpectedOutDBTablesNames: []string{
				`test1`,
				`test2`,
				`test3`,
			},
		},
		"case14/up(migration2,job2)->down(migration2,job3)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migsDummySources,
			ExpectedOutStatus_UseSource: migsDummySources,
			MigrationOp: []migrationOp{
				{
					Code:        up,
					Sources:     migs2Sources,
					DeployJobID: DeployJob2ID,
				},
				{
					Code:        down,
					Sources:     migs2Sources,
					DeployJobID: DeployJob3ID,
					FailureProps: stdcase.FailureProps{
						ShouldFail:      true,
						ShouldFailIsErr: errs.ErrCurrentDeployNotRegistered,
					},
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					Empty: true,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					Empty: true,
				},

				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit2Job2_DeployFingerprint,
					Num:               3,
					ApplyOrder:        3,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  1,
					Fingerprint: migrations_edit2Job2_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        1,
							ApplyOrder: 1,
						},
						{
							Num:        2,
							ApplyOrder: 2,
						},
						{
							Num:        3,
							ApplyOrder: 3,
						},
					},
				},
			},
			ExpectedOutDBTablesNames: []string{
				`test1`,
				`test2`,
				`test3`,
			},
		},
		"case15/up(migration1,job2)->up(migration2,job3)->up(migration1,job2)->down(migration2,job3)->up(migration2,job3)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migsDummySources,
			ExpectedOutStatus_UseSource: migsDummySources,
			MigrationOp: []migrationOp{
				{
					Code:        up,
					Sources:     migrations_edit1.Sources,
					DeployJobID: DeployJob2ID,
				},
				{
					Code:        up,
					Sources:     migrations_edit2.Sources,
					DeployJobID: DeployJob3ID,
				},
				{
					Code:        up,
					Sources:     migrations_edit1.Sources,
					DeployJobID: DeployJob2ID,
				},
				{
					Code:        down,
					Sources:     migrations_edit2.Sources,
					DeployJobID: DeployJob3ID,
				},
				{
					Code:        up,
					Sources:     migrations_edit2.Sources,
					DeployJobID: DeployJob3ID,
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion:  types.LatestVersionStorageStatus{Empty: true},
				LatestDeploy:   types.LatestDeployStorageStatus{Empty: true},
				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit2Job3_DeployFingerprint,
					Num:               3,
					ApplyOrder:        4,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  2,
					Fingerprint: migrations_edit2Job3_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        3,
							ApplyOrder: 4,
						},
					},
				},
			},
			ExpectedOutDBTablesNames: []string{
				`test1`,
				`test2`,
				`test3`,
			},
		},
		"case16/up(migration1,job2)->up(migration2,job3)->up(migration1,job3)->down(migration2,job3)->up(migration2,job3)": {
			CommonProps: stdcase.CommonProps{
				OmitOnDebug: false,
			},

			ExpectedInStatus_UseSource:  migsDummySources,
			ExpectedOutStatus_UseSource: migsDummySources,
			MigrationOp: []migrationOp{
				{
					Code:        up,
					Sources:     migrations_edit1.Sources,
					DeployJobID: DeployJob2ID,
				},
				{
					Code:        up,
					Sources:     migrations_edit2.Sources,
					DeployJobID: DeployJob3ID,
				},
				{
					Code:        up,
					Sources:     migrations_edit1.Sources,
					DeployJobID: DeployJob3ID,
					FailureProps: stdcase.FailureProps{
						ShouldFail:      true,
						ShouldFailIsErr: errs.ErrCurrentDeployBehindLatestVersion,
					},
				},
				{
					Code:        down,
					Sources:     migrations_edit2.Sources,
					DeployJobID: DeployJob3ID,
				},
				{
					Code:        up,
					Sources:     migrations_edit2.Sources,
					DeployJobID: DeployJob3ID,
				},
			},

			ExpectedInStatus: types.StorageStatus{
				LatestVersion:  types.LatestVersionStorageStatus{Empty: true},
				LatestDeploy:   types.LatestDeployStorageStatus{Empty: true},
				NotInitialized: true,
			},
			ExpectedOutStatus: types.StorageStatus{
				LatestVersion: types.LatestVersionStorageStatus{
					DeployAppRelease:  appReleaseFn,
					DeployFingerprint: migrations_edit2Job3_DeployFingerprint,
					Num:               3,
					ApplyOrder:        4,
				},
				LatestDeploy: types.LatestDeployStorageStatus{
					AppRelease:  appReleaseFn,
					ApplyOrder:  2,
					Fingerprint: migrations_edit2Job3_DeployFingerprint,
					IsDirty:     false,
					Versions: []types.LatestDeployStorageStatusVersion{
						{
							Num:        3,
							ApplyOrder: 4,
						},
					},
				},
			},
			ExpectedOutDBTablesNames: []string{
				`test1`,
				`test2`,
				`test3`,
			},
		},
	}

	for tcName := range testCases {
		tc := testCases[tcName]

		t.Run(
			tcName,
			func(t *testing.T) {
				ctx := tc.Pre(t)

				api, _, _ := storageapi_thelper.NewPgDB[storageapi.SQLAPI](t, ctx)

				hlpRequireDBStatusWithDBAPI(t, ctx, nil, tc.ExpectedInStatus_UseSource, api, tc.ExpectedInStatus, "\"in\" DB migration status not matches expected one")

				for y := 0; y < len(tc.MigrationOp); y++ {
					op := tc.MigrationOp[y]

					cfgCustomizer := func(cfg *storagemig.Conf) {
						if op.DeployJobID != "" {
							cfg.DeployJobID = op.DeployJobID

							return
						}
					}

					switch op.Code {
					case up:
						hlpRequireUpWithDBAPI(t, ctx, cfgCustomizer, fmt.Sprintf("op%d", y+1), op.Sources, api, op.DryRun, op.FailureProps)
					case down:
						hlpRequireDownWithDBAPI(t, ctx, cfgCustomizer, fmt.Sprintf("op%d", y+1), op.Sources, api, op.DryRun, op.FailureProps)
					default:
						t.Logf("[ERROR] unknown test case migration service operation code %d", op.Code)
						t.FailNow()
					}
				}

				hlpRequireDBStatusWithDBAPI(t, ctx, nil, tc.ExpectedOutStatus_UseSource, api, tc.ExpectedOutStatus, "\"out\" DB migration status not matches expected one")
				hlpRequirePublicSchemeTablesNamesWithDBAPI(t, ctx, api, tc.ExpectedOutDBTablesNames)
			},
		)
	}
}

func hlpRequireUpWithDBAPI(t *testing.T, ctx context.Context, confCustomizer func(*storagemig.Conf), name string, sources fs.FS, api storageapi.SQLAPI, dryRun bool, failure stdcase.FailureProps) {
	t.Helper()

	svc := storagemig_thelper.NewService(
		t,
		ctx,
		sources,
		func(t *testing.T) storagemig.Conf {
			conf := storagemig_thelper.FlenvRootConf(t)
			if confCustomizer != nil {
				confCustomizer(&conf)
			}

			return conf
		},
		api,
	)

	err := svc.Up(ctx, dryRun)
	failure.RequireConformErrorMsgf(t, err, "actual migration UP operation '%s' result error not matches to expected one", name)
}

func hlpRequireDownWithDBAPI(t *testing.T, ctx context.Context, confCustomizer func(*storagemig.Conf), name string, sources fs.FS, api storageapi.SQLAPI, dryRun bool, failure stdcase.FailureProps) {
	t.Helper()

	svc := storagemig_thelper.NewService(
		t,
		ctx,
		sources,
		func(t *testing.T) storagemig.Conf {
			conf := storagemig_thelper.FlenvRootConf(t)
			if confCustomizer != nil {
				confCustomizer(&conf)
			}

			return conf
		},
		api,
	)

	err := svc.Down(ctx, dryRun)
	failure.RequireConformErrorMsgf(t, err, "actual migration DOWN operation '%s' result error not matches to expected one", name)
}

func hlpRequireDBStatusWithService(t *testing.T, ctx context.Context, svc storagemig.Service, expected types.StorageStatus, msgAndArgs ...any) {
	t.Helper()

	actual, err := svc.Status(ctx)
	require.NoError(t, err, "unexpected error while getting DB migrations status")

	actual.Storage.LatestVersion.DeployTimestamp = expected.LatestVersion.DeployTimestamp
	actual.Storage.LatestDeploy.Timestamp = expected.LatestDeploy.Timestamp

	require.Equal(t, expected, actual.Storage, msgAndArgs...)
}

func hlpRequireDBStatusWithDBAPI(t *testing.T, ctx context.Context, confCustomizer func(*storagemig.Conf), sources fs.FS, api storageapi.SQLAPI, expected types.StorageStatus, msgAndArgs ...any) {
	t.Helper()

	svc := storagemig_thelper.NewService(
		t,
		ctx,
		sources,
		func(t *testing.T) storagemig.Conf {
			conf := storagemig_thelper.FlenvRootConf(t)
			if confCustomizer != nil {
				confCustomizer(&conf)
			}

			return conf
		},
		api,
	)

	hlpRequireDBStatusWithService(t, ctx, svc, expected, msgAndArgs...)
}

func hlpRequirePublicSchemeTablesNamesWithDBAPI(t *testing.T, ctx context.Context, api storageapi.SQLAPI, expected []string) {
	t.Helper()

	actual := []string{}

	err := api.ExecSQL(ctx, publicSchemeTablesNamesGetQuery, &actual)
	require.NoError(t, err, "unexpected error while executing query of DB tables' names")

	require.ElementsMatch(t, expected, actual, "list of DB public scheme tables' names doesn't match to expected one")
}

func hlpTestMigrationsNUpStepsOpen(t *testing.T, st types.SourceVersionList, sources fs.FS) {
	t.Helper()

	for i := 0; i < len(st); i++ {
		ver := &st[i]

		fn := fmt.Sprintf("%d_%s.up.sql", ver.Num, ver.Name)

		f, err := sources.Open(fn)
		require.NoError(t, err, "error while opening migration source file; file: %s", fn)

		fc, err := io.ReadAll(f)
		require.NoError(t, err, "error while reading migration source file; file: %s", fn)

		ver.Script = string(fc)
	}
}

func hlpTestMigrations1UpStepsOpen(t *testing.T) types.SourceVersionList {
	t.Helper()

	st := types.SourceVersionList{
		{
			Name: "initial",
			Num:  1,
		},
		{
			Name: "next",
			Num:  2,
		},
	}

	hlpTestMigrationsNUpStepsOpen(t, st, migrations_edit1.Sources)

	return st
}
