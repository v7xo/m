package contract

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"

	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
)

type (
	JSONUnmarshalerFn[T any]       func([]byte) (T, error)
	YAMLUnmarshalerFn[T any]       func(*yaml.Node) (T, error)
	RegistrationOptionForFn[T any] func(*RegistrationOptionsFor[T]) error

	RegistrationOptionForIface[T any] interface {
		Apply(*RegistrationOptionsFor[T]) error
	}

	RegistrationOptionsFor[T any] struct {
		yamlUnmarshaler      YAMLUnmarshalerFn[T]
		jsonUnmarshaler      JSONUnmarshalerFn[T]
		registrar            Registrar
		hasCustomUnmarshaler bool
	}
)

var (
	_ RegistrationOptions             = RegistrationOptionsFor[int]{}
	_ RegistrationOptionForIface[int] = RegistrationOptionForFn[int](nil)
)

func (c RegistrationOptionsFor[T]) Registrar() Registrar { return c.registrar }

func (c RegistrationOptionsFor[T]) HasCustomUnmarshaler() bool { return c.hasCustomUnmarshaler }

func (c RegistrationOptionsFor[T]) JSONUnmarshaler() JSONUnmarshalerFn[any] {
	if c.jsonUnmarshaler == nil {
		return nil
	}

	return func(b []byte) (any, error) { return c.jsonUnmarshaler(b) }
}

func (c RegistrationOptionsFor[T]) YAMLUnmarshaler() YAMLUnmarshalerFn[any] {
	if c.yamlUnmarshaler == nil {
		return nil
	}

	return func(n *yaml.Node) (any, error) { return c.yamlUnmarshaler(n) }
}

func newRegistrationOptionsFor[T any]() *RegistrationOptionsFor[T] {
	return &RegistrationOptionsFor[T]{
		yamlUnmarshaler: yamlUnmarshaler[T],
		jsonUnmarshaler: jsonUnmarshaler[T],
	}
}

func yamlUnmarshaler[T any](n *yaml.Node) (val T, err error) {
	err = n.Decode(&val)

	return
}

func jsonUnmarshaler[T any](b []byte) (val T, err error) {
	err = json.Unmarshal(b, &val)

	return
}

func (o RegistrationOptionForFn[T]) WithRegistrar(val Registrar) RegistrationOptionForFn[T] {
	return fopt.Chain(o, func(t *RegistrationOptionsFor[T]) { t.registrar = val })
}

func (o RegistrationOptionForFn[T]) WithYAMLUnmarshaler(val YAMLUnmarshalerFn[T]) RegistrationOptionForFn[T] {
	return fopt.Chain(o, func(t *RegistrationOptionsFor[T]) { t.yamlUnmarshaler, t.hasCustomUnmarshaler = val, true })
}

func (o RegistrationOptionForFn[T]) WithJSONUnmarshaler(val JSONUnmarshalerFn[T]) RegistrationOptionForFn[T] {
	return fopt.Chain(o, func(t *RegistrationOptionsFor[T]) { t.jsonUnmarshaler, t.hasCustomUnmarshaler = val, true })
}

func (o RegistrationOptionForFn[T]) Apply(t *RegistrationOptionsFor[T]) error { return o(t) }

func RegistrationOptionFor[T any]() RegistrationOptionForFn[T] {
	return fopt.Nop[RegistrationOptionsFor[T]]
}

func Register[T any](opts ...RegistrationOptionForIface[T]) (err error) {
	defer func() {
		if err == nil || errors.Is(err, ErrRegistration) {
			return
		}

		err = fmt.Errorf("%w; failed to register contract for type '%s'; %w", ErrRegistration, stringsext.TypeStringFor[T](), err)
	}()

	optsColl := newRegistrationOptionsFor[T]()

	for _, o := range opts {
		if o == nil {
			continue
		}

		if err = o.Apply(optsColl); err != nil {
			return err
		}
	}

	registrar := optsColl.registrar
	switch registrar {
	case nil:
		if registrar, err = DefaultRegistrar(); err != nil {
			return
		}
	}

	ctx := contextext.AppGracefulEnd(nil)

	return registrar.Register(ctx, reflect.TypeFor[T](), optsColl)
}

func MustRegister[T any](opts ...RegistrationOptionForIface[T]) {
	err := Register(opts...)
	if err != nil {
		panic(err)
	}
}

func NameFor[T any]() (string, error) {
	reg, err := DefaultRegistrar()
	if err != nil {
		return "", err
	}

	return reg.ContractName(reflect.TypeFor[T]())
}
