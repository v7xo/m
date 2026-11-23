package enum

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

var (
	cache   = map[reflect.Type]any{}
	cacheMu = &sync.RWMutex{}

	_ encoding.TextMarshaler   = Of[int8]{}
	_ encoding.TextUnmarshaler = &Of[int8]{}
	_ yaml.Marshaler           = Of[int8]{}
	_ yaml.Unmarshaler         = &Of[int8]{}
	_ json.Marshaler           = Of[int8]{}
	_ json.Unmarshaler         = &Of[int8]{}
)

type Constraint interface {
	~int8 | ~int16 | ~int32 | ~int64 | ~int
}

type Of[T Constraint] struct{ Val T }

func (v Of[T]) IsZero() bool { return v.Val == T(0) }

func (v Of[T]) HasAllFlags(f T) bool { return v.Val&f == f }

func (v Of[T]) HasAnyFlags(f T) bool { return f == 0 || v.Val&f != 0 }

func (v Of[T]) Name() string {
	typ := reflect.TypeFor[T]()

	cacheMu.RLock()

	e, ok := cache[typ]

	cacheMu.RUnlock()

	if !ok {
		return placeholder.UnknownEnumInt(v.Val)
	}

	return e.(*descriptor[T]).name(v.Val) //nolint:forcetypeassert
}

func (v Of[T]) String() string { return placeholder.UnspecifiedIfEmpty(v.Name()) }

func (v Of[T]) MarshalYAML() (any, error) { return v.Name(), nil }

func (v *Of[T]) UnmarshalYAML(n *yaml.Node) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to unmarshal enum '%s' from YAML; %w", stringsext.TypeStringFor[T](), err)
		}
	}()

	var name string

	if err = n.Decode(&name); err != nil {
		return
	}

	v.Val, err = Parse[T](name)

	return nil
}

func (v Of[T]) MarshalText() ([]byte, error) { return []byte(v.Name()), nil }

func (v *Of[T]) UnmarshalText(text []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to unmarshal enum '%s' from '%s'; %w", stringsext.TypeStringFor[T](), stringsext.TruncateDefault(string(text)), err)
		}
	}()

	v.Val, err = Parse[T](string(text))

	return nil
}

func (v Of[T]) MarshalJSON() ([]byte, error) { return json.Marshal(v.Name()) }

func (v *Of[T]) UnmarshalJSON(b []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to unmarshal enum '%s' from JSON; %w", stringsext.TypeStringFor[T](), err)
		}
	}()

	var (
		name string
		val  T
	)

	if err = json.Unmarshal(b, &name); err != nil {
		unmarshalType := &json.UnmarshalTypeError{}

		if !errors.As(err, &unmarshalType) {
			return
		}

		if errNum := json.Unmarshal(b, &val); errNum != nil {
			return errNum
		}

		return
	} else if val, err = Parse[T](name); err != nil {
		return
	}

	v.Val = val

	return nil
}

func String[T Constraint](val T) string { return Of[T]{Val: val}.String() }

func StringSlice[E Constraint, S ~[]E](val S) string {
	if val == nil {
		return placeholder.Nil
	} else if len(val) == 0 {
		return placeholder.Empty
	}

	buf := make([]byte, 0, 128) // nolint:mnd,revive

	buf = append(buf, '[')

	buf = append(buf, "count: "...)
	buf = append(buf, stringsext.FormatInt(len(val))...)

	for i := range len(val) {
		buf = append(buf, ", "...)

		if i == 32 {
			buf = append(buf, []byte(stringsext.TruncateElipsis)...)

			break
		}

		buf = append(buf, String(val[i])...)
	}

	buf = append(buf, ']')

	return string(buf)
}

func Val[T Constraint](val T) Of[T] { return Of[T]{Val: val} }

func Names[T Constraint]() stringsext.ListString {
	typ := reflect.TypeFor[T]()

	cacheMu.RLock()

	cacheVal, ok := cache[typ]

	cacheMu.RUnlock()

	if !ok {
		panic(errs.ErrEnumNotRegistered)
	}

	desc := cacheVal.(*descriptor[T]) // nolint:forcetypeassert
	res := make([]string, 0, len(desc.valsSorted))

	for i := range desc.valsSorted {
		n := desc.valToName[desc.valsSorted[i]]
		if n == "" {
			continue
		}

		res = append(res, n)
	}

	return res
}

func Parse[T Constraint](s string) (_ T, err error) {
	defer func() {
		if err != nil {
			err = errs.JoinIfNot(fmt.Errorf("failed to parse '%s' as a value of enum '%s'; %w", s, stringsext.TypeStringFor[T](), err), errs.ErrBadArgs)
		}
	}()

	typ := reflect.TypeFor[T]()

	cacheMu.RLock()

	desc, ok := cache[typ]

	cacheMu.RUnlock()

	if !ok {
		return 0, errs.ErrEnumNotRegistered
	}

	return desc.(*descriptor[T]).val(s) // nolint:forcetypeassert
}

func ParseFallback[T Constraint](val string, fallback T) (T, error) {
	res, err := Parse[T](val)
	if errors.Is(err, errs.ErrUnknownEnumInvalidFormat) {
		return fallback, nil
	}

	return res, err
}

func RegisterFlag[T Constraint](names map[T]string) error { return register[T](names, true) }

func MustRegisterFlag[T Constraint](names map[T]string) struct{} {
	if err := RegisterFlag(names); err != nil {
		panic(err)
	}

	return struct{}{}
}

func Register[T Constraint](names map[T]string) (err error) { return register[T](names, false) }

func MustRegister[T Constraint](names map[T]string) struct{} {
	if err := Register(names); err != nil {
		panic(err)
	}

	return struct{}{}
}

func register[T Constraint](names map[T]string, flag bool) (err error) {
	typ := reflect.TypeFor[T]()

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to register enum type '%s'; %w", stringsext.TypeStringFor[T](), err)
		}
	}()

	valToName := maps.Clone(names)

	if len(valToName) == 0 {
		return fmt.Errorf("%w; enum names not specified or empty", errs.ErrBadArgs)
	}

	cacheMu.RLock()

	_, ok := cache[typ]

	cacheMu.RUnlock()

	if ok {
		return errors.New("already registered")
	}

	nameToVal := make(map[string]T, len(valToName))
	vals := make([]T, 0, len(valToName))

	for val, name := range valToName {
		if flag && strings.IndexByte(name, NameDelim) > -1 {
			return fmt.Errorf("%w; name '%s' contain '%s', it is not allowed for flag enum", errs.ErrBadArgs, name, string(NameDelim))
		}

		if otherVal, otherValExists := nameToVal[name]; otherValExists {
			return fmt.Errorf("%w; name '%s' is not unique, same name used for both values '%d' and '%d' at least", errs.ErrBadArgs, name, otherVal, val)
		}

		nameToVal[name] = val
		vals = append(vals, val)
	}

	slices.Sort(vals)

	entry := &descriptor[T]{
		flag:       flag,
		nameToVal:  nameToVal,
		valsSorted: vals,
		valToName:  valToName,
	}

	cacheMu.Lock()

	if _, ok = cache[typ]; ok {
		cacheMu.Unlock()

		return errors.New("already registered")
	}

	cache[typ] = entry

	cacheMu.Unlock()

	return
}

type descriptor[T Constraint] struct {
	valToName  map[T]string
	nameToVal  map[string]T
	valsSorted []T
	flag       bool
}

func (d *descriptor[T]) name(val T) string {
	switch {
	case !d.flag:
		n, ok := d.valToName[val]
		if ok {
			return n
		}

		return placeholder.UnknownEnumInt(val)
	}

	buf := make([]byte, 0, 64)
	c := 0

	for _, nameVal := range d.valsSorted {
		if nameVal == 0 {
			continue
		}

		if val&nameVal == nameVal {
			if c != 0 {
				buf = append(buf, NameDelim)
			}

			buf = append(buf, d.valToName[nameVal]...)

			c++
		}

		val = val & ^nameVal

		if val == 0 {
			break
		}
	}

	if val != 0 {
		if c != 0 {
			buf = append(buf, NameDelim)
		}

		buf = append(buf, placeholder.UnknownEnumInt(val)...)
	}

	return string(buf)
}

func (d *descriptor[T]) val(s string) (T, error) {
	parseName := func(n string) (T, error) {
		val, ok := d.nameToVal[n]
		if ok {
			return val, nil
		}

		return placeholder.ParseUknownEnumInt[T](n)
	}

	if !d.flag {
		return parseName(s)
	}

	var res T

	for name := range strings.SplitSeq(s, NameDelimString) {
		parsed, parseErr := parseName(name)
		if parseErr != nil {
			return 0, parseErr
		}

		res |= parsed
	}

	return res, nil
}

type ListOf[T Constraint] []Of[T]

func (v *ListOf[T]) FromString(val string) (err error) {
	defer func() {
		if err != nil {
			err = errs.JoinIfNot(
				fmt.Errorf("failed to parse list of enum '%s' from string '%s'", typestring.TypeStringFor[T](), stringsext.TruncateDefault(val)),
				errs.ErrBadArgs)
		}
	}()

	typ := reflect.TypeFor[T]()

	cacheMu.RLock()

	descAny, ok := cache[typ]

	cacheMu.RUnlock()

	if !ok {
		return errs.ErrEnumNotRegistered
	}

	desc := descAny.(*descriptor[T]) // nolint:forcetypeassert

	var cp ListOf[T]

	switch {
	case desc.flag:
		val, valErr := desc.val(val)
		if valErr != nil {
			return valErr
		}

		cp = ListOf[T]{Of[T]{Val: val}}
	default:
		for n := range strings.SplitSeq(val, NameDelimString) {
			val, valErr := desc.val(n)
			if valErr != nil {
				return valErr
			}

			cp = append(cp, Of[T]{Val: val})
		}
	}

	*v = cp

	return nil
}

func (v ListOf[T]) Vals() []T {
	if v == nil {
		return nil
	}

	res := make([]T, len(v))

	for i := range len(v) {
		res[i] = v[i].Val
	}

	return res
}

func ParseList[T Constraint](val string) (res ListOf[T], err error) {
	err = res.FromString(val)

	return
}

func MustParseList[T Constraint](val string) ListOf[T] {
	res, err := ParseList[T](val)
	if err != nil {
		panic(err)
	}

	return res
}
