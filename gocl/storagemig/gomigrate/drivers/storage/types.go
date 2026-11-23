package storage

import "time"

type (
	versionState struct {
		CreatedAt  time.Time `db:"created_at"`
		ModifiedAt time.Time `db:"modified_at"`

		DeployAppRelease string `db:"deploy_app_release"`

		DeployFingerprint int64 `db:"deploy_fingerprint"`

		ApplyOrder int32 `db:"apply_order"`
		Num        int32 `db:"version_num"`

		IsDeployDirty bool `db:"is_deploy_dirty"`
	}

	deployState struct {
		CreatedAt  time.Time `db:"created_at"`
		ModifiedAt time.Time `db:"modified_at"`

		AppRelease string `db:"app_release"`

		Fingerprint int64 `db:"fingerprint"`

		ApplyOrder        int32 `db:"apply_order"`
		VersionApplyOrder int32 `db:"version_apply_order"`
		VersionNum        int32 `db:"version_num"`

		IsDirty        bool `db:"is_dirty"`
		IsUnregistered bool `db:"is_unregistered"`
		NoVersion      bool `db:"no_version"`
	}
)
