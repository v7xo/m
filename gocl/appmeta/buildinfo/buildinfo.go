package buildinfo

import (
	"context"
	"regexp"
	"runtime/debug"

	"rn.m.vo7.io/vo7io/m/gocl/value/valonce"
)

const DevelVersion = "devel"

// WARNING! Constants below are set via CI (on build app binary). Be careful to change its names.
//
// Not intended for direct use in code.
var (
	Name             = "" //nolint
	VersionExplicit  = "" //nolint
	GitSha           = "" //nolint
	ReleaseAuthor    = "" //nolint
	ReleaseTimestamp = "" //nolint
	MetricsNamespace = "" //nolint
)

// nolint
var (
	versionWithSuffixRegexp = regexp.MustCompile(`^v[0-9]+.*$`)

	// cmdName represents currently executing app command.
	CmdName = ""

	buildInfoVersion = valonce.New[string](
		func(_ context.Context) (string, error) {
			if bi, ok := debug.ReadBuildInfo(); ok {
				v := bi.Main.Version

				switch {
				case v == "(devel)":
					return DevelVersion, nil
				case versionWithSuffixRegexp.MatchString(v):
					return v[1:], nil
				}

				return v, nil
			}

			return "", nil
		})
)

func Version() string {
	if VersionExplicit != "" {
		return VersionExplicit
	}

	v, err := buildInfoVersion.Eval(context.Background())
	if err != nil {
		return ""
	}

	return v
}
