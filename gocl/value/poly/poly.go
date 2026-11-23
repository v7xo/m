package poly

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/value/poly/contract"
)

var ErrPoly = contract.ErrContract

type Of[T any] struct{ Val T }

func (v Of[T]) IsZero() bool { return reflect.TypeOf(v.Val) == nil }

func (v Of[T]) MarshalJSON() (_ []byte, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to marshal poly-contract value of '%s' to JSON; %w", stringsext.TypeStringFor[T](), err)
			err = errs.JoinIfNot(err, ErrPoly)
		}
	}()

	valTyp := reflect.TypeOf(v.Val)
	if valTyp == nil {
		return json.Marshal(nil)
	}

	// nolint:tagliatelle
	type carrier struct {
		Data         T      `json:".data"`
		ContractName string `json:".type"`
	}

	contractName, err := contractNameOf(valTyp)
	if err != nil {
		return nil, err
	}

	return json.Marshal(carrier{ContractName: contractName, Data: v.Val})
}

func (v *Of[T]) UnmarshalJSON(b []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to unmarshal poly-contract value of '%s' from JSON; %w", stringsext.TypeStringFor[T](), err)
			err = errs.JoinIfNot(err, ErrPoly)
		}
	}()

	// nolint:tagliatelle
	carrier := struct {
		ContractName string          `json:".type"`
		Data         json.RawMessage `json:".data"`
	}{}

	if err = json.Unmarshal(b, &carrier); err != nil {
		return
	}

	unmarshaler, err := jsonUnmarshaler[T](carrier.ContractName)
	if err != nil {
		return
	}

	val, err := unmarshaler(carrier.Data)
	if err != nil {
		return err
	}

	*v = Of[T]{Val: val}

	return
}

func (v Of[T]) MarshalYAML() (_ any, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to marshal poly-contract value of '%s' to YAML; %w", stringsext.TypeStringFor[T](), err)
			err = errs.JoinIfNot(err, ErrPoly)
		}
	}()

	valTyp := reflect.TypeOf(v.Val)
	if valTyp == nil {
		return
	}

	// nolint:tagliatelle
	type carrier struct {
		Data         T      `yaml:".data"`
		ContractName string `yaml:".type"`
	}

	contractName, err := contractNameOf(valTyp)
	if err != nil {
		return nil, err
	}

	return carrier{ContractName: contractName, Data: v.Val}, nil
}

func (v *Of[T]) UnmarshalYAML(n *yaml.Node) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to unmarshal poly-contract value of '%s' from YAML; %w", stringsext.TypeStringFor[T](), err)
			err = errs.JoinIfNot(err, ErrPoly)
		}
	}()

	defer func() { err = errs.Join(err, errs.PanicErr(recover())) }()

	// nolint:tagliatelle
	carrier := struct {
		ContractName string    `yaml:".type"`
		Data         yaml.Node `yaml:".data"`
	}{}

	if err = n.Decode(&carrier); err != nil {
		return
	}

	unmarshaler, err := yamlUnmarshaler[T](carrier.ContractName)
	if err != nil {
		return
	}

	val, err := unmarshaler(&carrier.Data)
	if err != nil {
		return err
	}

	*v = Of[T]{Val: val}

	return
}

func (v Of[T]) String() string {
	sb := &strings.Builder{}

	switch valTyped := any(v.Val).(type) {
	case nil:
		sb.WriteString(placeholder.Nil)
	case fmt.Stringer:
		sb.WriteString(stringsext.Stringer(valTyped))
	default:
		sb.WriteByte('(')
		fmt.Fprintf(sb, "%v", v.Val)
		sb.WriteByte(')')
	}

	return sb.String()
}

func jsonUnmarshaler[T any](contractName string) (contract.JSONUnmarshalerFn[T], error) {
	reg, err := contract.DefaultRegistrar()
	if err != nil {
		return nil, err
	}

	unmarshalerAny, err := reg.JSONUnmarshaler(contractName)
	if err != nil {
		return nil, err
	}

	return func(b []byte) (res T, err error) {
			val, err := unmarshalerAny(b)
			if err != nil {
				return
			}

			switch valTyped := val.(type) {
			case T:
				res = valTyped
			case nil:
			default:
				err = fmt.Errorf("unable to convert unmarshaled value of type '%s' to required type '%s'",
					stringsext.TypeString(reflect.ValueOf(val).Type()),
					stringsext.TypeStringFor[T]())
			}

			return
		},
		nil
}

func yamlUnmarshaler[T any](contractName string) (contract.YAMLUnmarshalerFn[T], error) {
	reg, err := contract.DefaultRegistrar()
	if err != nil {
		return nil, err
	}

	unmarshalerAny, err := reg.YAMLUnmarshaler(contractName)
	if err != nil {
		return nil, err
	}

	return func(n *yaml.Node) (res T, err error) {
			val, err := unmarshalerAny(n)
			if err != nil {
				return
			}

			switch valTyped := val.(type) {
			case T:
				res = valTyped
			case nil:
			default:
				err = fmt.Errorf("unable to convert unmarshaled value of type '%s' to required type '%s'",
					stringsext.TypeString(reflect.ValueOf(val).Type()),
					stringsext.TypeStringFor[T]())
			}

			return
		},
		nil
}

func contractNameOf(typ reflect.Type) (string, error) {
	reg, err := contract.DefaultRegistrar()
	if err != nil {
		return "", err
	}

	return reg.ContractName(typ)
}

func Val[T any](val T) Of[T] { return Of[T]{Val: val} }
