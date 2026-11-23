package startconf

import "strings"

// [ConfVariant] represents special name of configuration set.
type ConfVariant string

const (
	ConfVariantUnspecified ConfVariant = ""

	ConfVariantAuto         ConfVariant = "auto"
	ConfVariantDevelopLocal ConfVariant = "develop-local"
	ConfVariantDevelop      ConfVariant = "develop"
	ConfVariantIgnore       ConfVariant = "ignore"
	ConfVariantTestLocal    ConfVariant = "test-local"
	ConfVariantTest         ConfVariant = "test"
	ConfVariantUserProfile  ConfVariant = "user-profile"
)

func (v ConfVariant) String() string { return string(v) }

func (v ConfVariant) IsAuto() bool { return strings.EqualFold(string(v), string(ConfVariantAuto)) }

func (v ConfVariant) IsDevelop() bool {
	return strings.EqualFold(string(v), string(ConfVariantDevelop))
}

func (v ConfVariant) IsDevelopLocal() bool {
	return strings.EqualFold(string(v), string(ConfVariantDevelopLocal))
}

func (v ConfVariant) IsIgnore() bool { return strings.EqualFold(string(v), string(ConfVariantIgnore)) }

func (v ConfVariant) IsTest() bool { return strings.EqualFold(string(v), string(ConfVariantTest)) }

func (v ConfVariant) IsUserProfile() bool {
	return strings.EqualFold(string(v), string(ConfVariantUserProfile))
}

func (v ConfVariant) IsTestLocal() bool {
	return strings.EqualFold(string(v), string(ConfVariantTestLocal))
}

func (v ConfVariant) IsUnspecified() bool {
	return strings.EqualFold(string(v), string(ConfVariantUnspecified))
}
