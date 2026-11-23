package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type (
	DeployFingerprint int64
	SourceFingerprint int64

	LatestDeployStorageStatusVersion struct {
		Num        uint
		ApplyOrder uint
	}

	LatestDeployStorageStatusVersions []LatestDeployStorageStatusVersion

	LatestVersionStorageStatus struct {
		DeployTimestamp   time.Time
		DeployAppRelease  string
		Num               uint
		DeployFingerprint DeployFingerprint
		ApplyOrder        uint
		Empty             bool
	}

	LatestDeployStorageStatus struct {
		Timestamp   time.Time
		AppRelease  string
		Versions    LatestDeployStorageStatusVersions
		ApplyOrder  uint
		Fingerprint DeployFingerprint
		IsDirty     bool
		Empty       bool
	}

	StorageStatus struct {
		UpgradableTo   SourceVersionList
		LatestDeploy   LatestDeployStorageStatus
		LatestVersion  LatestVersionStorageStatus
		NotInitialized bool
	}

	SourceVersion struct {
		Name   string
		Script string
		Num    uint
	}

	SourceVersionList []SourceVersion

	SourceGoTextTemplatingDataContext struct {
		Up       bool
		Down     bool
		MockFlow bool
	}

	SourceStatus struct {
		Up          SourceVersionList
		Down        SourceVersionList
		Fingerprint SourceFingerprint
	}

	Status struct {
		Source  SourceStatus
		Storage StorageStatus

		CurrentDeployFingerprint DeployFingerprint
	}

	Lib               string
	SourceDriverName  string
	StorageDriverName string
)

func (s StorageStatus) String() string {
	if s.NotInitialized {
		return placeholder.NotInitialized
	}

	return fmt.Sprintf("version: %s", s.LatestVersion)
}

func (s SourceVersion) String() string {
	//nolint:revive,mnd
	return strconv.FormatUint(uint64(s.Num), 10) + " '" + s.Name + "'"
}

func (s SourceVersionList) String() string {
	if len(s) == 0 {
		return placeholder.Empty
	}

	versionsStrs := make([]string, len(s))
	for i := range versionsStrs {
		versionsStrs[i] = s[i].String()
	}

	return strings.Join(versionsStrs, ", ")
}

func (s SourceStatus) String() string {
	return fmt.Sprintf("fingerprint: %s; up: %d version(s); down: %d version(s)", s.Fingerprint, len(s.Up), len(s.Down))
}

func (s Status) String() string {
	return fmt.Sprintf("DB: %s\nSource: %s\nCurrent deploy fingerprint: %s", s.Storage, s.Source, s.CurrentDeployFingerprint)
}

func (s LatestVersionStorageStatus) String() string {
	if s.Empty {
		return placeholder.Empty
	}

	return fmt.Sprintf("%d (deploy fingerprint: %s, timestamp: %s, deploy app release: '%s')", s.Num, s.DeployFingerprint, s.DeployTimestamp.Format(time.RFC3339Nano), s.DeployAppRelease)
}

func (s LatestDeployStorageStatusVersion) String() string {
	//nolint:revive,mnd
	return strconv.FormatUint(uint64(s.Num), 10)
}

func (s LatestDeployStorageStatusVersions) String() string {
	if len(s) == 0 {
		return placeholder.Empty
	}

	strs := make([]string, len(s))
	for i := range strs {
		strs[i] = s[i].String()
	}

	return strings.Join(strs, ", ")
}

func (s LatestDeployStorageStatus) String() string {
	if s.Empty {
		return placeholder.Empty
	}

	return fmt.Sprintf("%s (dirty: %v, versions: %s; timestamp: %s, app release: '%s')", s.Fingerprint, s.IsDirty, s.Versions, s.Timestamp.Format(time.RFC3339Nano), s.AppRelease)
}

func (f Lib) IsGoMigrate() bool {
	return strings.EqualFold(string(GoMigrateLib), string(f))
}

func (f Lib) IsUnspecified() bool {
	return f == ""
}

func (f Lib) String() string {
	return string(f)
}

func (f Lib) S() string {
	return string(f)
}

func (f SourceDriverName) String() string {
	return string(f)
}

func (f StorageDriverName) String() string { return string(f) }

func (v DeployFingerprint) String() string {
	if v == UnassignedDeployFingerprint {
		return placeholder.Unassigned
	}

	return "(" + stringsext.FormatInt(v) + ")"
}

func (v SourceFingerprint) String() string {
	if v == UnassignedSourcesFingerprint {
		return placeholder.Unassigned
	}

	return "(" + stringsext.FormatInt(v) + ")"
}
