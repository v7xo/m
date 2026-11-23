package stringsext

import (
	"fmt"
	"math"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
)

type ListString = List[string]

type List[V ~string] []V

func (v List[V]) Set() map[V]struct{} {
	if v == nil {
		return nil
	}

	m := make(map[V]struct{}, min(1024, len(v))) // nolint:mnd,revive

	for i := range len(v) {
		m[v[i]] = struct{}{}
	}

	return m
}

func (v List[V]) ParseInt[T ~int | ~int8 | ~int16 | ~int32 | ~int64]() ([]T, error) {
	if v == nil {
		return nil, nil
	}

	typ := reflect.TypeFor[T]()
	size := 8 // nolint:mnd,revive

	// nolint:mnd,revive,exhaustive
	switch typ.Kind() {
	case reflect.Int64:
		size = 64
	case reflect.Int32:
		size = 32
	case reflect.Int16:
		size = 16
	case reflect.Int:
		if math.MaxInt == math.MaxInt32 {
			size = 32
		} else if math.MaxInt == math.MaxInt64 {
			size = 64
		}
	}

	res := make([]T, len(v))

	for i := range v {
		val, err := strconv.ParseInt(string(v[i]), 10, size)
		if err != nil {
			return nil, fmt.Errorf("failed to parse item '%s' at '%d' position as '%s'; %w", TruncateDefault(string(v[i])), i, typestring.TypeStringFor[T](), err)
		}

		res[i] = T(val)
	}

	return res, nil
}

func (v List[V]) ParseUint[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64]() ([]T, error) {
	if v == nil {
		return nil, nil
	}

	typ := reflect.TypeFor[T]()
	size := 8 // nolint:mnd,revive

	// nolint:mnd,revive,exhaustive
	switch typ.Kind() {
	case reflect.Uint64:
		size = 64
	case reflect.Uint32:
		size = 32
	case reflect.Uint16:
		size = 16
	case reflect.Uint:
		if math.MaxUint == math.MaxUint32 {
			size = 32
		} else if math.MaxUint == math.MaxUint64 {
			size = 64
		}
	}

	res := make([]T, len(v))

	for i := range v {
		val, err := strconv.ParseUint(string(v[i]), 10, size)
		if err != nil {
			return nil, fmt.Errorf("failed to parse item '%s' at '%d' position as '%s'; %w", TruncateDefault(string(v[i])), i, typestring.TypeStringFor[T](), err)
		}

		res[i] = T(val)
	}

	return res, nil
}

func (v List[V]) Equal(o List[V]) bool { return slices.Equal(v, o) }

func (v List[V]) LastOrZero() (V, bool) {
	if len(v) == 0 {
		return "", false
	}

	return v[len(v)-1], true
}

func (v List[V]) FirstOrZero() (V, bool) {
	if len(v) == 0 {
		return "", false
	}

	return v[0], true
}

func (v List[V]) String() string {
	if v == nil {
		return placeholder.Nil
	} else if len(v) == 0 {
		return placeholder.Empty
	}

	sb := &strings.Builder{}

	sb.WriteByte('[')

	sb.WriteString("count: ")
	sb.WriteString(FormatInt(len(v)))

	for i := range len(v) {
		sb.WriteString(", ")

		if i == 32 {
			sb.WriteRune(TruncateElipsisSym)

			break
		}

		sb.WriteString(placeholder.EmptyIf(string(v[i])))
	}

	sb.WriteByte(']')

	return sb.String()
}
