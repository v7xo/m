package contract

import "rn.m.vo7.io/vo7io/m/gocl/value/enum"

//nolint:unused
type (
	// LocalName represents a special type used to designate a local name of [Contract].
	LocalName[L LocalNameLocator] struct{ localName L }
	localNameAttr                 interface{ isPolyLocalNameAttribute() }

	// Namespace represents a special type used to designate a namespace of [Contract].
	Namespace[N NamespaceLocator] struct{ namespace N }
	namespaceAttr                 interface{ isPolyNamespaceAttribute() }

	// Name represents a special type used to designate a full name of [Contract].
	Name[N NamespaceLocator, L LocalNameLocator] struct {
		namespace N
		localName L
	}
	nameAttr interface{ isPolyNameAttribute() }

	AttrKind int8
)

const (
	AttrKindUnspecified AttrKind = iota
	AttrKindLocalName
	AttrKindNamespace
	AttrKindName
)

var (
	_ = enum.MustRegister(
		map[AttrKind]string{
			AttrKindLocalName:   "local_name",
			AttrKindName:        "name",
			AttrKindNamespace:   "namespace",
			AttrKindUnspecified: "",
		})
	_ localNameAttr = LocalName[LocalNameLocator]{}
	_ nameAttr      = Name[NamespaceLocator, LocalNameLocator]{}
	_ namespaceAttr = Namespace[NamespaceLocator]{}
)

func (LocalName[L]) isPolyLocalNameAttribute() {}
func (Name[N, L]) isPolyNameAttribute()        {}
func (Namespace[N]) isPolyNamespaceAttribute() {}
func (v AttrKind) IsLocalName() bool           { return v == AttrKindLocalName }
func (v AttrKind) IsName() bool                { return v == AttrKindName }
func (v AttrKind) IsNamespace() bool           { return v == AttrKindNamespace }
func (v AttrKind) IsUnspecified() bool         { return v == AttrKindUnspecified }
func (v AttrKind) String() string              { return enum.String(v) }
func ParseAttrKind(s string) (AttrKind, error) { return enum.Parse[AttrKind](s) }
