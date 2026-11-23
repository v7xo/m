package contextext

import (
	"context"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
)

func AttrDict(ctx context.Context) (attrname.Dict, bool, error) {
	return attrname.DictFromContext(ctx)
}

func MustAttrDict(ctx context.Context) attrname.Dict {
	return attrname.MustDictFromContext(ctx)
}

// WithAttrDictPairs creates context with attrs dictionary appended by p pairs of field name (see [attrname.Name]) and value: <name1>, <value1>, ..., <nameN>, <valueN>.
func WithAttrDictPairs(ctx context.Context, p ...any) context.Context {
	return attrname.DictPairsContext(ctx, p...)
}

func WithAttrDict(ctx context.Context, dict attrname.Dict) context.Context {
	return attrname.DictContext(ctx, dict)
}
