package envext

import "context"

type VarProvider interface {
	VariablesDict(context.Context) (map[string]string, error)
}
