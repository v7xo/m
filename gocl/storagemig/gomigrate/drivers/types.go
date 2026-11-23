package drivers

import (
	"fmt"
	"time"

	value "rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/types"
)

type (
	VersionDBState struct {
		CreateTs          time.Time
		ModifyTs          time.Time
		DeployAppRelease  string
		DeployFingerprint types.DeployFingerprint
		Num               uint
		ApplyOrder        uint
		Empty             bool
		IsDeployDirty     bool
	}

	DeployDBStateVersion struct {
		Num        uint
		ApplyOrder uint
	}

	DeployDBState struct {
		CreateTs    time.Time
		ModifyTs    time.Time
		AppRelease  string
		Versions    []DeployDBStateVersion
		Fingerprint types.DeployFingerprint
		ApplyOrder  uint
		IsDirty     bool
		Empty       bool
	}
)

func (v VersionDBState) String() string {
	if v.Empty {
		return value.Empty
	}

	return fmt.Sprintf("%d (deploy fingerprint: %s, apply order: %d, timestamp: %s, deploy app release: '%s')", v.Num, v.DeployFingerprint, v.ApplyOrder, v.ModifyTs.Format(time.RFC3339Nano), v.DeployAppRelease)
}

func (v DeployDBState) String() string {
	if v.Empty {
		return value.Empty
	}

	return fmt.Sprintf("%s (dirty: %v, timestamp: %s, app release: '%s')", v.Fingerprint, v.IsDirty, v.ModifyTs.Format(time.RFC3339Nano), v.AppRelease)
}
