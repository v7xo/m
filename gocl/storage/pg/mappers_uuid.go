package pg

import (
	"github.com/google/uuid"
	"github.com/jackc/pgtype"

	"rn.m.vo7.io/vo7io/m/gocl/storage/pgmap"
)

func UUIDArrayFrom(v []uuid.UUID) (res pgtype.UUIDArray) {
	return pgmap.UUIDArrayFrom(v)
}

// nolint:predeclared
func UUIDArrayFromFn(len int, fn func(i int) uuid.UUID) (res pgtype.UUIDArray) {
	return pgmap.UUIDArrayFromFn(len, fn)
}
