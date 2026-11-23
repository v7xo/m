package poly

import "rn.m.vo7.io/vo7io/m/gocl/value/poly/contract"

type (
	AsPackage                                                      = contract.AsPackage
	AsTag                                                          = contract.AsTag
	AsType                                                         = contract.AsType
	LocalName[L contract.LocalNameLocator]                         = contract.LocalName[L]
	Name[N contract.NamespaceLocator, L contract.LocalNameLocator] = contract.Name[N, L] // Declares Poly-contract name for type.
	Namespace[N contract.NamespaceLocator]                         = contract.Namespace[N]
)
