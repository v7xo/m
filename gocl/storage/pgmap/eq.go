package pgmap

import "github.com/jackc/pgtype"

func EqTimestamptz(a, b pgtype.Timestamptz) bool {
	return a.Status == b.Status &&
		a.InfinityModifier == b.InfinityModifier &&
		a.Time.UnixMicro() == b.Time.UnixMicro()
}

func EqTimestamp(a, b pgtype.Timestamp) bool {
	return a.Status == b.Status &&
		a.InfinityModifier == b.InfinityModifier &&
		a.Time.UnixMicro() == b.Time.UnixMicro()
}
