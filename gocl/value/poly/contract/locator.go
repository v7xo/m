package contract

import "rn.m.vo7.io/vo7io/m/gocl/value/enum"

type (
	NamespaceLocator interface{ IsPolyContractNamespaceLocator() }
	LocalNameLocator interface{ IsPolyContractLocalNameLocator() }

	AsPackage struct{}
	AsTag     struct{}
	AsType    struct{}

	LocatorKind int32
)

const (
	LocatorKindUnspecified LocatorKind = iota
	LocatorKindAsTag
	LocatorKindAsPackage
	LocatorKindAsType
	LocatorKindRedirect
)

var (
	_ LocalNameLocator = AsTag{}
	_ LocalNameLocator = AsType{}
	_ NamespaceLocator = AsPackage{}
	_ NamespaceLocator = AsTag{}

	_ = enum.MustRegister(
		map[LocatorKind]string{
			LocatorKindAsPackage:   "as_package",
			LocatorKindAsTag:       "as_tag",
			LocatorKindAsType:      "as_type",
			LocatorKindRedirect:    "redirect",
			LocatorKindUnspecified: "",
		})
)

func (AsPackage) IsPolyContractNamespaceLocator() {}
func (AsTag) IsPolyContractLocalNameLocator()     {}
func (AsTag) IsPolyContractNamespaceLocator()     {}
func (AsType) IsPolyContractLocalNameLocator()    {}
func (AsType) IsPolyContractNamespaceLocator()    {}

func (v LocatorKind) IsAsPackage() bool   { return v == LocatorKindAsPackage }
func (v LocatorKind) IsAsTag() bool       { return v == LocatorKindAsTag }
func (v LocatorKind) IsAsType() bool      { return v == LocatorKindAsType }
func (v LocatorKind) IsRedirect() bool    { return v == LocatorKindRedirect }
func (v LocatorKind) IsUnspecified() bool { return v == LocatorKindUnspecified }
func (v LocatorKind) String() string      { return enum.String(v) }
