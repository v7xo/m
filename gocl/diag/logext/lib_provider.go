package logext

import "strings"

type LibProvider string

const (
	LibProviderUnspecified = LibProvider("")
	LibProviderGoKit       = LibProvider("gokit")
)

func (v LibProvider) IsGoKit() bool { return strings.EqualFold(string(LibProviderGoKit), string(v)) }

func (v LibProvider) IsKnown() bool       { return strings.EqualFold(string(LibProviderGoKit), string(v)) }
func (v LibProvider) IsUnspecified() bool { return v == LibProviderUnspecified }

func (v LibProvider) String() string { return string(v) }
