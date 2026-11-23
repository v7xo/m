package info

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/appmeta/buildinfo"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

// Name returns application name.
func Name() string {
	if buildinfo.Name == "" {
		return "app"
	}

	return buildinfo.Name
}

// CmdName returns entry point command name (if any).
func CmdName() string {
	return placeholder.UnknownIfEmpty(buildinfo.CmdName)
}

// CmdAppQualifiedName returns app name and command name in form <app_name>/<cmd_name>.
func CmdAppQualifiedName() string {
	return Name() + "/" + CmdName()
}

func AboutShort() string {
	return stringsext.TruncateEnd(
		fmt.Sprintf(
			"%s, v. %s",
			placeholder.UnknownIfEmpty(buildinfo.Name),
			placeholder.UnknownIfEmpty(buildinfo.Version())),
		UserAgentAppNameMaxLen)
}

func AboutLong() string {
	return fmt.Sprintf(
		"%s (version: %s, git: %s, release ts: %s, metrics namespace: %s)",
		placeholder.UnknownIfEmpty(buildinfo.Name),
		placeholder.UnknownIfEmpty(buildinfo.Version()),
		placeholder.UnknownIfEmpty(buildinfo.GitSha),
		placeholder.UnknownIfEmpty(buildinfo.ReleaseTimestamp),
		placeholder.UnknownIfEmpty(buildinfo.MetricsNamespace))
}

func Version() VersionString {
	return VersionString(placeholder.UnknownIfEmpty(buildinfo.Version()))
}

func VersionLong() string {
	return fmt.Sprintf(
		"%s, git: %s, release ts: %s",
		placeholder.UnknownIfEmpty(buildinfo.Version()),
		placeholder.UnknownIfEmpty(buildinfo.GitSha),
		placeholder.UnknownIfEmpty(buildinfo.ReleaseTimestamp),
	)
}

// UserAgentLong returns special client app identification string.
// It can be used as "UserAgent" in some scenario (Kafka client, HTTP client, PostreSQL conn, etc).
func UserAgentLong() string {
	return fmt.Sprintf(
		"%s, %s (version: %s, git: %s, release ts: %s)",
		placeholder.UnknownIfEmpty(buildinfo.Name),
		placeholder.UnknownIfEmpty(buildinfo.CmdName),
		placeholder.UnknownIfEmpty(buildinfo.Version()),
		placeholder.UnknownIfEmpty(buildinfo.GitSha),
		placeholder.UnknownIfEmpty(buildinfo.ReleaseTimestamp),
	)
}

// UserAgentShort returns special client app identification string.
// It can be used as "UserAgent" in some scenario (Kafka client, HTTP client, storage connection, etc).
// Max. len of returned string is [UserAgentAppNameMaxLen].
func UserAgentShort() string {
	return stringsext.TruncateEnd(
		fmt.Sprintf(
			"%s, %s, v. %s",
			placeholder.UnknownIfEmpty(buildinfo.Name),
			placeholder.UnknownIfEmpty(buildinfo.CmdName),
			placeholder.UnknownIfEmpty(buildinfo.Version())),
		UserAgentAppNameMaxLen)
}

func BuildInfoString() string {
	if bi, ok := debug.ReadBuildInfo(); ok {
		return formatBuildInfo(bi)
	}

	return placeholder.Missing
}

func formatBuildInfo(bi *debug.BuildInfo) string {
	// see original source code here go/src/runtime/debug/mod.go:104
	buf := new(strings.Builder)

	if bi.GoVersion != "" {
		fmt.Fprintf(buf, "go\t%s\n", bi.GoVersion)
	}

	if bi.Path != "" {
		fmt.Fprintf(buf, "path\t%s\n", bi.Path)
	}

	quoteKey := func(key string) bool {
		return len(key) == 0 || strings.ContainsAny(key, "= \t\r\n\"`")
	}

	quoteValue := func(value string) bool {
		return strings.ContainsAny(value, " \t\r\n\"`")
	}

	var formatMod func(string, debug.Module)

	formatMod = func(word string, m debug.Module) {
		buf.WriteString(word)
		buf.WriteByte('\t')
		buf.WriteString(m.Path)
		buf.WriteByte('\t')
		buf.WriteString(m.Version)

		if m.Replace == nil {
			buf.WriteByte('\t')
			buf.WriteString(m.Sum)
		} else {
			buf.WriteByte('\n')
			formatMod("=>", *m.Replace)
		}

		buf.WriteByte('\n')
	}

	if bi.Main != (debug.Module{}) {
		formatMod("mod", bi.Main)
	}

	for _, s := range bi.Settings {
		key := s.Key
		if quoteKey(key) {
			key = strconv.Quote(key)
		}

		value := s.Value
		if quoteValue(value) {
			value = strconv.Quote(value)
		}

		fmt.Fprintf(buf, "build\t%s=%s\n", key, value)
	}

	return buf.String()
}
