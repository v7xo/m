package pg

import (
	"github.com/jackc/pgtype"

	"rn.m.vo7.io/vo7io/m/gocl/storage/pgmap"
)

func EqTimestamptz(a, b pgtype.Timestamptz) bool {
	return pgmap.EqTimestamptz(a, b)
}

func EqTimestamp(a, b pgtype.Timestamp) bool {
	return pgmap.EqTimestamp(a, b)
}
