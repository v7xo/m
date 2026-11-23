package pg

import (
	"github.com/jackc/pgtype"

	"rn.m.vo7.io/vo7io/m/gocl/storage/pgmap"
)

func BoolArrayFrom[T interface{ ~bool }](v []T) (res pgtype.BoolArray) {
	return pgmap.BoolArrayFrom(v)
}

func BoolArrayFromFn[T interface{ ~bool }](length int, fn func(i int) T) (res pgtype.BoolArray) {
	return pgmap.BoolArrayFromFn(length, fn)
}

func BoolArrayFromPtr[T interface{ ~bool }](v []*T) (res pgtype.BoolArray) {
	return pgmap.BoolArrayFromPtr(v)
}

func BoolArrayFromPtrFn[T interface{ ~bool }](length int, fn func(i int) *T) (res pgtype.BoolArray) {
	return pgmap.BoolArrayFromPtrFn(length, fn)
}
