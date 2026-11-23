package logext

import (
	"strings"
)

type Format string

const (
	FormatJSON   Format = "json"
	FormatLogfmt Format = "logfmt"
)

func (f Format) IsJSON() bool   { return strings.EqualFold(string(FormatJSON), string(f)) }
func (f Format) IsLogfmt() bool { return strings.EqualFold(string(FormatLogfmt), string(f)) }
func (f Format) String() string { return string(f) }

func (f Format) IsKnown() bool {
	return strings.EqualFold(string(FormatJSON), string(f)) ||
		strings.EqualFold(string(FormatLogfmt), string(f))
}
