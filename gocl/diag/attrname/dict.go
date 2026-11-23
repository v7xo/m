package attrname

import (
	"context"
	"fmt"
	"iter"
	"maps"
	"strconv"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type (
	Dict map[Name]any
	Attr struct {
		V any
		N Name
	}
)

func (d Dict) Seq() iter.Seq2[Name, any] { return maps.All(d) }

// FromContext returns value of attribute n (see [Name]) from specified context ctx.
// If context ctx contains attributes dictionary [Dict] and that dictionary contains value for specified key, and that value is of specified type T, then value returned.
func FromContext[T any](ctx context.Context, n Name) (val T, ok bool) {
	dict, ok, err := DictFromContext(ctx)
	if !ok || err != nil {
		return
	}

	valAny, ok := dict[n]
	if !ok {
		return
	}

	val, ok = valAny.(T)

	return
}

func DictFromContext(ctx context.Context) (Dict, bool, error) {
	if ctx == nil {
		return nil, false, nil
	}

	val := ctx.Value(ContextKeyLogAttrs)
	if val == nil {
		return nil, false, nil
	}

	if valTyped, ok := val.(Dict); ok {
		return valTyped, true, nil
	}

	return nil,
		false,
		fmt.Errorf(
			"%w; such type '%s' of value of context key '%s' is unexpected, expected '%s'",
			errs.ErrUnexpectedTypeOfValueInContext,
			typestring.TypeStringOfVal(val),
			ContextKeyLogAttrs,
			typestring.TypeStringFor[Dict](),
		)
}

func MustDictFromContext(ctx context.Context) Dict {
	d, _, err := DictFromContext(ctx)
	if err != nil {
		panic(err)
	}

	return d
}

// DictPairsContext creates context with attrs dictionary [Dict] appended by p pairs of field name (see [Name]) and value: <name1>, <value1>, ..., <nameN>, <valueN>.
func DictPairsContext(ctx context.Context, p ...any) context.Context {
	if len(p) == 0 {
		return ctx
	}

	pSet, _, _ := DictFromContext(ctx)
	res := make(Dict, len(pSet)+len(p)/2+1)

	maps.Copy(res, pSet)

	for i := 0; i < len(p); i += 2 {
		var (
			n Name
			v any
		)

		if nTyped, ok := p[i].(Name); ok {
			n = nTyped
		} else if nTyped, ok := p[i].(string); ok {
			n = Name(nTyped)
		} else if nTyped, ok := p[i].(interface{ String() string }); ok {
			n = Name(nTyped.String())
		} else {
			n = Name(string(Unnamed) + strconv.Itoa(i/2+1))
		}

		if i != len(p)-1 {
			v = p[i+1]
		}

		res[n] = v
	}

	return context.WithValue(ctx, ContextKeyLogAttrs, res)
}

func DictContext(ctx context.Context, dict Dict) context.Context {
	if outer, ok, _ := DictFromContext(ctx); ok {
		dict = MergeDict(dict, outer, true)
	}

	return context.WithValue(ctx, ContextKeyLogAttrs, dict)
}

func MergeDict(left, right Dict, leftWins bool) Dict {
	if len(left) == 0 {
		return right
	}

	if len(right) == 0 {
		return left
	}

	var res, winner, loser Dict

	res = make(Dict, len(right)+len(left))
	winner, loser = right, left

	if leftWins {
		winner, loser = left, right
	}

	maps.Copy(res, winner)

	for k, v := range loser {
		if _, ok := res[k]; ok {
			continue
		}

		res[k] = v
	}

	return res
}

func (v Dict) Merge(dict Dict) Dict { return MergeDict(v, dict, false) }
