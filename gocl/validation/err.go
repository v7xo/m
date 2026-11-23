package validation

import (
	"fmt"

	vldr "github.com/go-playground/validator/v10"
)

type Error struct{ errorF func() string }

func (err Error) Error() string { return err.errorF() }

func errFromGoPlaygroundValidator(err vldr.ValidationErrors) error {
	return Error{errorF: err.Error}
}

func Err(msg string) error {
	return Error{errorF: func() string { return msg }}
}

func Errf(msg string, args ...any) error {
	m := fmt.Sprintf(msg, args...)

	return Error{errorF: func() string { return m }}
}
