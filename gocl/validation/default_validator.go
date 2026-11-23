package validation

import (
	"errors"
	"fmt"
	"sync"

	validatorpkg "github.com/go-playground/validator/v10"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
)

type (
	Validator interface {
		Struct(val any) error
	}

	validatorV10Wrapper struct{ impl *validatorpkg.Validate }
)

//nolint:gochecknoglobals
var (
	singleton     Validator
	singletonOnce sync.Once
)

func (s *validatorV10Wrapper) Struct(val any) error {
	err := s.impl.Struct(val)
	if err == nil {
		return nil
	}

	validatorPkgErr := validatorpkg.ValidationErrors{}
	if errors.As(err, &validatorPkgErr) {
		return fmt.Errorf("validation failed (type '%s'); %w", stringsext.TypeStringOfVal(val), errFromGoPlaygroundValidator(validatorPkgErr))
	}

	return fmt.Errorf("value validation failed (type '%s'); %w", stringsext.TypeStringOfVal(val), err)
}

// DefaultValidator returns default validator singleton.
func DefaultValidator() Validator {
	singletonOnce.Do(func() { singleton = &validatorV10Wrapper{impl: validatorpkg.New()} })

	return singleton
}
