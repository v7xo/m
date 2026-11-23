package placeholder

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
)

var ErrUnknownEnumInvalidFormat = clerr.Errf("unknown enum name string representation not valid")

func UnknownIfEmpty(v string) string {
	if len(v) == 0 {
		return Unknown
	}

	return v
}

func MissingIfEmpty(v string) string {
	if len(v) == 0 {
		return Missing
	}

	return v
}

func UnspecifiedIfEmpty(v string) string {
	if len(v) == 0 {
		return Unspecified
	}

	return v
}

func EmptyIf(v string) string {
	if len(v) == 0 {
		return Empty
	}

	return v
}

func UnknownEnumInt[T ~int | ~int64 | ~int32 | ~int16 | ~int8](v T) string {
	return "<unknown(" + strconv.FormatInt(int64(v), 10) + ")>"
}

func ParseUknownEnumInt[T ~int | ~int64 | ~int32 | ~int16 | ~int8](val string) (_ T, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("unable to parse int value from string '%s' representing unknown enum member; %w", val, err)
		}
	}()

	var zero T

	if cut, ok := strings.CutPrefix(val, "<unknown("); ok {
		switch cut, ok = strings.CutSuffix(cut, ")>"); ok {
		case true:
			var res64 int64

			res64, err = strconv.ParseInt(cut, 10, 64)
			if err != nil {
				return zero, err
			}

			if int64(T(res64)) != res64 {
				return zero, fmt.Errorf("parsed int '%d' overflows valid range for type '%s'", res64, reflect.TypeFor[T]())
			}

			return T(res64), nil
		}
	}

	return zero, ErrUnknownEnumInvalidFormat
}

func UnknownEnumUint[T ~uint | ~uint64 | ~uint32 | ~uint16 | ~uint8](v T) string {
	return "<unknown(" + strconv.FormatUint(uint64(v), 10) + ")>"
}
