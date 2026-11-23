package typestring

import (
	"reflect"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

func TypeStringFor[T any]() string {
	return TypeString(reflect.TypeFor[T]())
}

func TypeStringForExpr[T any](expr T) string {
	return TypeString(reflect.TypeFor[T]())
}

func TypeStringOfVal(val any) string {
	valRefl := reflect.ValueOf(val)

	if !valRefl.IsValid() {
		return TypeString(nil)
	}

	return TypeString(valRefl.Type())
}

func TypeString(typ reflect.Type) string {
	if typ == nil {
		return placeholder.Nil
	}

	const pointerSpecifier = "*"

	pointerPrepos := ""

	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
		pointerPrepos = pointerSpecifier
	}

	pkg := typ.PkgPath()
	if pkg != "" {
		pkg = ", " + pkg
	}

	n := typ.Name()
	if n == "" {
		n = typ.String()
	}

	return "{" + pointerPrepos + n + pkg + "}"
}
