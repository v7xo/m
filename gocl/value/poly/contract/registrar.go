package contract

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/value/valonce"
)

type (
	Registrar interface {
		Register(ctx context.Context, typ reflect.Type, opts RegistrationOptions) error

		ContractName(typ reflect.Type) (string, error)
		YAMLUnmarshaler(contractName string) (YAMLUnmarshalerFn[any], error)
		JSONUnmarshaler(contractName string) (JSONUnmarshalerFn[any], error)
	}

	RegistrationOptions interface {
		Registrar() Registrar
		YAMLUnmarshaler() YAMLUnmarshalerFn[any]
		JSONUnmarshaler() JSONUnmarshalerFn[any]
		HasCustomUnmarshaler() bool
	}
)

type (
	registrarRecord struct {
		contract             *Contract
		yamlUnmarshaler      YAMLUnmarshalerFn[any]
		jsonUnmarshaler      JSONUnmarshalerFn[any]
		hasCustomUnmarshaler bool
	}
	registrar struct {
		fopt.LogOpt
		types *sync.Map
		names *sync.Map

		desc Describer
	}
	registrarTypesVal = valonce.Of[*registrarRecord]
	registrarNamesVal = *registrarRecord
)

func (v *registrarRecord) equal(b *registrarRecord) bool {
	if v == nil {
		return b == nil
	} else if b == nil || b.hasCustomUnmarshaler || v.hasCustomUnmarshaler {
		return false
	}

	return v.contract.Type == b.contract.Type
}

func (v *registrarRecord) String() string {
	if v == nil {
		return placeholder.Nil
	}

	buf := make([]byte, 0, 64)

	buf = append(buf, '(')

	buf = append(buf, "contract: "...)
	if v.contract == nil {
		buf = append(buf, placeholder.Nil...)
	} else {
		buf = append(buf, v.contract.String()...)

		buf = append(buf, ", type: "...)
		buf = append(buf, stringsext.TypeString(v.contract.Type)...)
	}

	buf = append(buf, ')')

	return string(buf)
}

func (s *registrar) Register(ctx context.Context, typ reflect.Type, opts RegistrationOptions) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w; failed to register contract for type '%s'; %w", ErrRegistration, stringsext.TypeString(typ), err)
		}
	}()

	if err = s.validateOpts(opts); err != nil {
		return
	}

	prepared, exists := s.types.Load(typ)
	if !exists {
		prepared, _ = s.types.LoadOrStore(typ, s.prepareRecordForTypeOnce(typ, opts))
	}

	rec, err := prepared.(registrarTypesVal).Eval(ctx) //nolint:forcetypeassert
	if err != nil {
		return err
	}

	otherRecVal, alreadyRegistered := s.names.LoadOrStore(rec.contract.Name, rec)
	if alreadyRegistered {
		otherRec := otherRecVal.(registrarNamesVal) //nolint:forcetypeassert

		switch {
		case otherRec.equal(rec):
			break
		default:
			return fmt.Errorf("%w; contract already registered as '%s' and unable reregister it as '%s'",
				errs.ErrAlreadyExists,
				otherRec,
				rec)
		}
	}

	return nil
}

func (s *registrar) ContractName(typ reflect.Type) (string, error) {
	prepared, exists := s.types.Load(typ)
	if !exists {
		return "", s.contractNotRegisteredYetErrForType(typ)
	}

	rec, ok := prepared.(registrarTypesVal).TryGet() //nolint:forcetypeassert
	if !ok {
		return "", s.contractNotRegisteredYetErrForType(typ)
	}

	return rec.contract.Name, nil
}

func (s *registrar) YAMLUnmarshaler(contractName string) (YAMLUnmarshalerFn[any], error) {
	val, exists := s.names.Load(contractName)
	if !exists {
		return nil, s.contractNotRegisteredYetErrForName(contractName)
	}

	unmarshaler := val.(registrarNamesVal).yamlUnmarshaler //nolint:forcetypeassert
	if unmarshaler == nil {
		return nil, fmt.Errorf("%w; %w; no YAML unmarshaler function associated with contract name '%s'", ErrContract, errs.ErrNotFound, contractName)
	}

	return unmarshaler, nil
}

func (s *registrar) JSONUnmarshaler(contractName string) (JSONUnmarshalerFn[any], error) {
	val, exists := s.names.Load(contractName)
	if !exists {
		return nil, s.contractNotRegisteredYetErrForName(contractName)
	}

	unmarshaler := val.(registrarNamesVal).jsonUnmarshaler //nolint:forcetypeassert
	if unmarshaler == nil {
		return nil, fmt.Errorf("%w; %w; no YAML unmarshaler function associated with contract name '%s'", ErrContract, errs.ErrNotFound, contractName)
	}

	return unmarshaler, nil
}

func (*registrar) validateOpts(opts RegistrationOptions) error {
	if opts == nil {
		return errors.New("registration options can't be a nil")
	}

	if opts.YAMLUnmarshaler() == nil {
		return errors.New("registration options not specify YAML unmarshaler (nil)")
	}

	return nil
}

func (*registrar) contractNotRegisteredYetErrForType(typ reflect.Type) error {
	return fmt.Errorf("%w; %w; contract for type '%s' not registered", ErrContract, errs.ErrNotFound, stringsext.TypeString(typ))
}

func (*registrar) contractNotRegisteredYetErrForName(name string) error {
	return fmt.Errorf("%w; %w; contract name '%s' not registered", ErrContract, errs.ErrNotFound, name)
}

func (s *registrar) prepareRecordForTypeOnce(typ reflect.Type, opts RegistrationOptions) registrarTypesVal {
	return valonce.New(
		func(ctx context.Context) (_ *registrarRecord, err error) {
			defer func() {
				if err != nil {
					err = fmt.Errorf("failed to prepare registration; %w", err)

					s.Log().Err(ctx, err).Send()

					return
				}

				s.Log().Trace(ctx).Msgf("succeeded to prepare registration for poly-contract of type '%s'", stringsext.TypeString(typ))
			}()

			c, err := s.desc.Describe(ctx, typ)
			if err != nil {
				return
			}

			return &registrarRecord{
				contract:             &c,
				yamlUnmarshaler:      opts.YAMLUnmarshaler(),
				jsonUnmarshaler:      opts.JSONUnmarshaler(),
				hasCustomUnmarshaler: opts.HasCustomUnmarshaler(),
			}, nil
		})
}

func NewRegistrar(opts ...RegistrarOptionFn) (_ Registrar, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[registrar](), err)
			err = errs.JoinIfNot(err, ErrContract)
		}
	}()

	impl := &registrar{
		types: &sync.Map{},
		names: &sync.Map{},
	}
	if err = fopt.Apply(impl, opts...); err != nil {
		return
	}

	if impl.desc == nil {
		return nil, errors.New("no describer specified")
	}

	if err = impl.SetLogScope(logScope); err != nil {
		return nil, err
	}

	return impl, nil
}

type RegistrarOptionFn func(*registrar) error

func (o RegistrarOptionFn) WithLog(val logext.Logger) RegistrarOptionFn { return fopt.WithLog(o, val) }

func (o RegistrarOptionFn) WithDescriber(val Describer) RegistrarOptionFn {
	return fopt.Chain(o, func(t *registrar) { t.desc = val })
}

func RegistrarOption() RegistrarOptionFn { return fopt.Nop[registrar] }
