package confset

import (
	"fmt"
	"reflect"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/confacet"
	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/diag/endpoint"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type set struct {
	confs map[reflect.Type]flenv.Conf
}

var (
	_                          Set = &set{}
	logConfTyp                     = reflect.TypeFor[*logext.Conf]()
	diagnosticsEndpointConfTyp     = reflect.TypeFor[*endpoint.Conf]()
)

func (s *set) isConfSet()        {}
func (s *set) Confs() ConfGetter { return ConfGetter{set: s} }

func (s *set) LogConf() (*logext.Conf, error) {
	res, ok := s.confs[logConfTyp]
	if !ok {
		return nil, nil
	}

	switch resTyped := res.(type) {
	case *logext.Conf:
		return resTyped, nil
	default:
		return nil, fmt.Errorf("%w; unexpected condition occurred: logging configuration object type '%s' mismatch required type '%s'",
			errs.ErrTypeMismatch,
			typestring.TypeStringOfVal(res),
			typestring.TypeStringFor[*logext.Conf]())
	}
}

func (s *set) DiagnosticsEndpointConf() (*endpoint.Conf, error) {
	res, ok := s.confs[diagnosticsEndpointConfTyp]
	if !ok {
		return nil, nil
	}

	switch resTyped := res.(type) {
	case *endpoint.Conf:
		return resTyped, nil
	default:
		return nil, fmt.Errorf("%w; unexpected condition occurred: logging configuration object type '%s' mismatch required type '%s'",
			errs.ErrTypeMismatch,
			typestring.TypeStringOfVal(res),
			typestring.TypeStringFor[*endpoint.Conf]())
	}
}

func (s *set) BindFlagSetPrefixDefault(prefix string, flset *pflag.FlagSet, def confacet.Set) error {
	for _, conf := range s.confs {
		if err := flenv.CallFlagBinder(conf, prefix, flset, def); err != nil {
			return err
		}
	}

	return nil
}

func (s *set) PrepareUse() error {
	for _, conf := range s.confs {
		if t, ok := conf.(flenv.UsePreparer); ok {
			if err := t.PrepareUse(); err != nil {
				return fmt.Errorf("failed to transfor '%s'; %w", typestring.TypeStringOfVal(conf), err)
			}
		}
	}

	return nil
}

func (s *set) Validate(v validation.Validator) error {
	if s == nil {
		return fmt.Errorf("%w; value type '%s'", errs.ErrNilValue, typestring.TypeStringForExpr(s))
	}

	for _, conf := range s.confs {
		if err := conf.Validate(v); err != nil {
			return fmt.Errorf("validation of '%s' failed; %w", typestring.TypeStringOfVal(conf), err)
		}
	}

	return nil
}
