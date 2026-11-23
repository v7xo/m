package pg

import (
	"github.com/jackc/pgtype"

	"rn.m.vo7.io/vo7io/m/gocl/storage/pgmap"
)

func Int64SliceToPgInt8(v []int64) (res pgtype.Int8Array) {
	res.Status = pgtype.Present
	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}

	res.Elements = make([]pgtype.Int8, len(v))

	for i := range v {
		res.Elements[i] = pgtype.Int8{Status: pgtype.Present, Int: v[i]}
	}

	return
}

func Int64PtrSliceToPgInt8(v []*int64) pgtype.Int8Array {
	var res pgtype.Int8Array

	res.Status = pgtype.Present
	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}

	res.Elements = make([]pgtype.Int8, len(v))

	for i := range v {
		if v[i] == nil {
			res.Elements[i] = pgtype.Int8{Status: pgtype.Null}
		} else {
			res.Elements[i] = pgtype.Int8{Status: pgtype.Present, Int: *v[i]}
		}
	}

	return res
}

func Int32SliceToPgInt4(v []int32) (res pgtype.Int4Array) {
	res.Status = pgtype.Present
	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}

	res.Elements = make([]pgtype.Int4, len(v))

	for i := range v {
		res.Elements[i] = pgtype.Int4{Status: pgtype.Present, Int: v[i]}
	}

	return
}

func Int32PtrSliceToPgInt4(v []*int32) (res pgtype.Int4Array) {
	res.Status = pgtype.Present
	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}

	res.Elements = make([]pgtype.Int4, len(v))

	for i := range v {
		if v[i] == nil {
			res.Elements[i] = pgtype.Int4{Status: pgtype.Null}
		} else {
			res.Elements[i] = pgtype.Int4{Status: pgtype.Present, Int: *v[i]}
		}
	}

	return
}

func Int2ArrayFrom[T interface{ ~int16 | ~int32 | ~int64 }](v []T) (res pgtype.Int2Array) {
	return pgmap.Int2ArrayFrom(v)
}

func Int2ArrayFromFn[T interface{ ~int16 | ~int32 | ~int64 }](length int, fn func(i int) T) (res pgtype.Int2Array) {
	return pgmap.Int2ArrayFromFn(length, fn)
}

func Int2ArrayFromPtr[T interface{ ~int16 | ~int32 | ~int64 }](v []*T) (res pgtype.Int2Array) {
	return pgmap.Int2ArrayFromPtr(v)
}

func Int2ArrayFromPtrFn[T interface{ ~int16 | ~int32 | ~int64 }](length int, fn func(i int) *T) (res pgtype.Int2Array) {
	return pgmap.Int2ArrayFromPtrFn(length, fn)
}

func Int4ArrayFrom[T interface{ ~int16 | ~int32 | ~int64 }](v []T) (res pgtype.Int4Array) {
	return pgmap.Int4ArrayFrom(v)
}

func Int4ArrayFromFn[T interface{ ~int16 | ~int32 | ~int64 }](length int, fn func(i int) T) (res pgtype.Int4Array) {
	return pgmap.Int4ArrayFromFn(length, fn)
}

func Int4ArrayFromPtr[T interface{ ~int16 | ~int32 | ~int64 }](v []*T) (res pgtype.Int4Array) {
	return pgmap.Int4ArrayFromPtr(v)
}

func Int4ArrayFromPtrFn[T interface{ ~int16 | ~int32 | ~int64 }](length int, fn func(i int) *T) (res pgtype.Int4Array) {
	return pgmap.Int4ArrayFromPtrFn(length, fn)
}

func Int8ArrayFrom[T interface{ ~int16 | ~int32 | ~int64 }](v []T) (res pgtype.Int8Array) {
	return pgmap.Int8ArrayFrom(v)
}

func Int8ArrayFromFn[T interface{ ~int16 | ~int32 | ~int64 }](length int, fn func(i int) T) (res pgtype.Int8Array) {
	return pgmap.Int8ArrayFromFn(length, fn)
}

func Int8ArrayFromPtr[T interface{ ~int16 | ~int32 | ~int64 }](v []*T) (res pgtype.Int8Array) {
	return pgmap.Int8ArrayFromPtr(v)
}

func Int8ArrayFromPtrFn[T interface{ ~int16 | ~int32 | ~int64 }](length int, fn func(i int) *T) (res pgtype.Int8Array) {
	return pgmap.Int8ArrayFromPtrFn(length, fn)
}
