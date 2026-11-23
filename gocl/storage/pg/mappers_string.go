package pg

import (
	"github.com/jackc/pgtype"

	"rn.m.vo7.io/vo7io/m/gocl/storage/pgmap"
)

func StringSliceToPgVarchar(v []string) (res pgtype.VarcharArray) {
	res.Status = pgtype.Present
	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}

	res.Elements = make([]pgtype.Varchar, len(v))

	for i := range v {
		res.Elements[i] = pgtype.Varchar{Status: pgtype.Present, String: v[i]}
	}

	return
}

func StringPtrSliceToPgVarchar(v []*string) (res pgtype.VarcharArray) {
	res.Status = pgtype.Present
	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}

	res.Elements = make([]pgtype.Varchar, len(v))

	for i := range v {
		if v[i] == nil {
			res.Elements[i] = pgtype.Varchar{Status: pgtype.Null}
		} else {
			res.Elements[i] = pgtype.Varchar{Status: pgtype.Present, String: *v[i]}
		}
	}

	return
}

func StringSliceToPgText(v []string) (res pgtype.TextArray) {
	res.Status = pgtype.Present
	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}

	res.Elements = make([]pgtype.Text, len(v))

	for i := range v {
		res.Elements[i] = pgtype.Text{Status: pgtype.Present, String: v[i]}
	}

	return
}

func StringPtrSliceToPgText(v []*string) (res pgtype.TextArray) {
	res.Status = pgtype.Present
	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}

	res.Elements = make([]pgtype.Text, len(v))

	for i := range v {
		if v[i] == nil {
			res.Elements[i] = pgtype.Text{Status: pgtype.Null}
		} else {
			res.Elements[i] = pgtype.Text{Status: pgtype.Present, String: *v[i]}
		}
	}

	return
}

func TextArrayFrom[T interface{ ~string }](v []T) (res pgtype.TextArray) {
	return pgmap.TextArrayFrom(v)
}

func TextArrayFromFn[T interface{ ~string }](length int, fn func(i int) T) (res pgtype.TextArray) {
	return pgmap.TextArrayFromFn(length, fn)
}

func TextArrayFromPtr[T interface{ ~string }](v []*T) (res pgtype.TextArray) {
	return pgmap.TextArrayFromPtr(v)
}

func TextArrayFromPtrFn[T interface{ ~string }](length int, fn func(i int) *T) (res pgtype.TextArray) {
	return pgmap.TextArrayFromPtrFn(length, fn)
}
