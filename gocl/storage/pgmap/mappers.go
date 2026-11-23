package pgmap

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

func BoolArrayFrom[T interface{ ~bool }](v []T) (res pgtype.BoolArray) {
	if v == nil {
		res.Status = pgtype.Null

		return
	}

	res.Status = pgtype.Present

	if len(v) == 0 {
		return
	}

	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}
	res.Elements = make([]pgtype.Bool, len(v))

	for i := range v {
		res.Elements[i] = pgtype.Bool{Status: pgtype.Present, Bool: bool(v[i])}
	}

	return
}

func BoolArrayFromFn[T interface{ ~bool }](length int, fn func(i int) T) (res pgtype.BoolArray) {
	sl := make([]T, length)

	for i := range length {
		sl[i] = fn(i)
	}

	return BoolArrayFrom(sl)
}

func BoolArrayFromPtr[T interface{ ~bool }](v []*T) (res pgtype.BoolArray) {
	if v == nil {
		res.Status = pgtype.Null

		return
	}

	res.Status = pgtype.Present

	if len(v) == 0 {
		return
	}

	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}
	res.Elements = make([]pgtype.Bool, len(v))

	for i := range v {
		if v[i] == nil {
			res.Elements[i] = pgtype.Bool{Status: pgtype.Null}

			continue
		}

		res.Elements[i] = pgtype.Bool{Status: pgtype.Present, Bool: bool(*v[i])}
	}

	return
}

func BoolArrayFromPtrFn[T interface{ ~bool }](length int, fn func(i int) *T) (res pgtype.BoolArray) {
	sl := make([]*T, length)

	for i := range length {
		sl[i] = fn(i)
	}

	return BoolArrayFromPtr(sl)
}

func Int2ArrayFrom[T interface{ ~int16 | ~int32 | ~int64 }](v []T) (res pgtype.Int2Array) {
	if v == nil {
		res.Status = pgtype.Null

		return
	}

	res.Status = pgtype.Present

	if len(v) == 0 {
		return
	}

	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}
	res.Elements = make([]pgtype.Int2, len(v))

	for i := range v {
		res.Elements[i] = pgtype.Int2{Status: pgtype.Present, Int: int16(v[i])}
	}

	return
}

func Int2ArrayFromFn[T interface{ ~int16 | ~int32 | ~int64 }](length int, fn func(i int) T) (res pgtype.Int2Array) {
	sl := make([]T, length)

	for i := range length {
		sl[i] = fn(i)
	}

	return Int2ArrayFrom(sl)
}

func Int2ArrayFromPtr[T interface{ ~int16 | ~int32 | ~int64 }](v []*T) (res pgtype.Int2Array) {
	if v == nil {
		res.Status = pgtype.Null

		return
	}

	res.Status = pgtype.Present

	if len(v) == 0 {
		return
	}

	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}
	res.Elements = make([]pgtype.Int2, len(v))

	for i := range v {
		if v[i] == nil {
			res.Elements[i] = pgtype.Int2{Status: pgtype.Null}

			continue
		}

		res.Elements[i] = pgtype.Int2{Status: pgtype.Present, Int: int16(*v[i])}
	}

	return
}

func Int2ArrayFromPtrFn[T interface{ ~int16 | ~int32 | ~int64 }](length int, fn func(i int) *T) (res pgtype.Int2Array) {
	sl := make([]*T, length)

	for i := range length {
		sl[i] = fn(i)
	}

	return Int2ArrayFromPtr(sl)
}

func Int4ArrayFrom[T interface{ ~int16 | ~int32 | ~int64 }](v []T) (res pgtype.Int4Array) {
	if v == nil {
		res.Status = pgtype.Null

		return
	}

	res.Status = pgtype.Present

	if len(v) == 0 {
		return
	}

	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}
	res.Elements = make([]pgtype.Int4, len(v))

	for i := range v {
		res.Elements[i] = pgtype.Int4{Status: pgtype.Present, Int: int32(v[i])}
	}

	return
}

func Int4ArrayFromFn[T interface{ ~int16 | ~int32 | ~int64 }](length int, fn func(i int) T) (res pgtype.Int4Array) {
	sl := make([]T, length)

	for i := range length {
		sl[i] = fn(i)
	}

	return Int4ArrayFrom(sl)
}

func Int4ArrayFromPtr[T interface{ ~int16 | ~int32 | ~int64 }](v []*T) (res pgtype.Int4Array) {
	if v == nil {
		res.Status = pgtype.Null

		return
	}

	res.Status = pgtype.Present

	if len(v) == 0 {
		return
	}

	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}
	res.Elements = make([]pgtype.Int4, len(v))

	for i := range v {
		if v[i] == nil {
			res.Elements[i] = pgtype.Int4{Status: pgtype.Null}

			continue
		}

		res.Elements[i] = pgtype.Int4{Status: pgtype.Present, Int: int32(*v[i])}
	}

	return
}

func Int4ArrayFromPtrFn[T interface{ ~int16 | ~int32 | ~int64 }](length int, fn func(i int) *T) (res pgtype.Int4Array) {
	sl := make([]*T, length)

	for i := range length {
		sl[i] = fn(i)
	}

	return Int4ArrayFromPtr(sl)
}

func Int8ArrayFrom[T interface{ ~int16 | ~int32 | ~int64 }](v []T) (res pgtype.Int8Array) {
	if v == nil {
		res.Status = pgtype.Null

		return
	}

	res.Status = pgtype.Present

	if len(v) == 0 {
		return
	}

	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}
	res.Elements = make([]pgtype.Int8, len(v))

	for i := range v {
		res.Elements[i] = pgtype.Int8{Status: pgtype.Present, Int: int64(v[i])}
	}

	return
}

func Int8ArrayFromFn[T interface{ ~int16 | ~int32 | ~int64 }](length int, fn func(i int) T) (res pgtype.Int8Array) {
	sl := make([]T, length)

	for i := range length {
		sl[i] = fn(i)
	}

	return Int8ArrayFrom(sl)
}

func Int8ArrayFromPtr[T interface{ ~int16 | ~int32 | ~int64 }](v []*T) (res pgtype.Int8Array) {
	if v == nil {
		res.Status = pgtype.Null

		return
	}

	res.Status = pgtype.Present

	if len(v) == 0 {
		return
	}

	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}
	res.Elements = make([]pgtype.Int8, len(v))

	for i := range v {
		if v[i] == nil {
			res.Elements[i] = pgtype.Int8{Status: pgtype.Null}

			continue
		}

		res.Elements[i] = pgtype.Int8{Status: pgtype.Present, Int: int64(*v[i])}
	}

	return
}

func Int8ArrayFromPtrFn[T interface{ ~int16 | ~int32 | ~int64 }](length int, fn func(i int) *T) (res pgtype.Int8Array) {
	sl := make([]*T, length)

	for i := range length {
		sl[i] = fn(i)
	}

	return Int8ArrayFromPtr(sl)
}

func TextArrayFrom[T interface{ ~string }](v []T) (res pgtype.TextArray) {
	if v == nil {
		res.Status = pgtype.Null

		return
	}

	res.Status = pgtype.Present

	if len(v) == 0 {
		return
	}

	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}
	res.Elements = make([]pgtype.Text, len(v))

	for i := range v {
		res.Elements[i] = pgtype.Text{Status: pgtype.Present, String: string(v[i])}
	}

	return
}

func TextArrayFromFn[T interface{ ~string }](length int, fn func(i int) T) (res pgtype.TextArray) {
	sl := make([]T, length)

	for i := range length {
		sl[i] = fn(i)
	}

	return TextArrayFrom(sl)
}

func TextArrayFromPtr[T interface{ ~string }](v []*T) (res pgtype.TextArray) {
	if v == nil {
		res.Status = pgtype.Null

		return
	}

	res.Status = pgtype.Present

	if len(v) == 0 {
		return
	}

	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}
	res.Elements = make([]pgtype.Text, len(v))

	for i := range v {
		if v[i] == nil {
			res.Elements[i] = pgtype.Text{Status: pgtype.Null}

			continue
		}

		res.Elements[i] = pgtype.Text{Status: pgtype.Present, String: string(*v[i])}
	}

	return
}

func TextArrayFromPtrFn[T interface{ ~string }](length int, fn func(i int) *T) (res pgtype.TextArray) {
	sl := make([]*T, length)

	for i := range length {
		sl[i] = fn(i)
	}

	return TextArrayFromPtr(sl)
}

func UUIDArrayFrom(v []uuid.UUID) (res pgtype.UUIDArray) {
	if v == nil {
		res.Status = pgtype.Null

		return
	}

	res.Status = pgtype.Present

	if len(v) == 0 {
		return
	}

	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}
	res.Elements = make([]pgtype.UUID, len(v))

	for i := range v {
		res.Elements[i] = pgtype.UUID{Status: pgtype.Present, Bytes: v[i]}
	}

	return
}

// nolint:predeclared
func UUIDArrayFromFn(len int, fn func(i int) uuid.UUID) (res pgtype.UUIDArray) {
	sl := make([]uuid.UUID, len)

	for i := range len {
		sl[i] = fn(i)
	}

	return UUIDArrayFrom(sl)
}

func UUIDArrayFromPtr(v []*uuid.UUID) (res pgtype.UUIDArray) {
	if v == nil {
		res.Status = pgtype.Null

		return
	}

	res.Status = pgtype.Present

	if len(v) == 0 {
		return
	}

	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}
	res.Elements = make([]pgtype.UUID, len(v))

	for i := range v {
		if v[i] == nil {
			res.Elements[i] = pgtype.UUID{Status: pgtype.Null}

			continue
		}

		res.Elements[i] = pgtype.UUID{Status: pgtype.Present, Bytes: *v[i]}
	}

	return
}

func UUIDArrayFromPtrFn(length int, fn func(i int) *uuid.UUID) (res pgtype.UUIDArray) {
	sl := make([]*uuid.UUID, length)

	for i := range length {
		sl[i] = fn(i)
	}

	return UUIDArrayFromPtr(sl)
}

func TimestamptzArrayFrom(v []time.Time) (res pgtype.TimestamptzArray) {
	if v == nil {
		res.Status = pgtype.Null

		return
	}

	res.Status = pgtype.Present

	if len(v) == 0 {
		return
	}

	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}
	res.Elements = make([]pgtype.Timestamptz, len(v))

	for i := range v {
		res.Elements[i] = pgtype.Timestamptz{Status: pgtype.Present, Time: v[i]}
	}

	return
}

func TimestamptzArrayFromFn(length int, fn func(i int) time.Time) (res pgtype.TimestamptzArray) {
	sl := make([]time.Time, length)

	for i := range length {
		sl[i] = fn(i)
	}

	return TimestamptzArrayFrom(sl)
}

func TimestamptzArrayFromPtr(v []*time.Time) (res pgtype.TimestamptzArray) {
	if v == nil {
		res.Status = pgtype.Null

		return
	}

	res.Status = pgtype.Present

	if len(v) == 0 {
		return
	}

	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}
	res.Elements = make([]pgtype.Timestamptz, len(v))

	for i := range v {
		if v[i] == nil {
			res.Elements[i] = pgtype.Timestamptz{Status: pgtype.Null}

			continue
		}

		res.Elements[i] = pgtype.Timestamptz{Status: pgtype.Present, Time: *v[i]}
	}

	return
}

func TimestamptzArrayFromPtrFn(length int, fn func(i int) *time.Time) (res pgtype.TimestamptzArray) {
	sl := make([]*time.Time, length)

	for i := range length {
		sl[i] = fn(i)
	}

	return TimestamptzArrayFromPtr(sl)
}

func ByteaArrayFrom[T ~[]byte](v []T) (res pgtype.ByteaArray) {
	if v == nil {
		res.Status = pgtype.Null

		return
	}

	res.Status = pgtype.Present

	if len(v) == 0 {
		return
	}

	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}
	res.Elements = make([]pgtype.Bytea, len(v))

	for i := range v {
		res.Elements[i] = pgtype.Bytea{Status: pgtype.Present, Bytes: v[i]}
	}

	return
}

func ByteaArrayFromFn[T ~[]byte](length int, fn func(i int) T) (res pgtype.ByteaArray) {
	sl := make([]T, length)

	for i := range length {
		sl[i] = fn(i)
	}

	return ByteaArrayFrom(sl)
}

func ByteaArrayFromPtr[T ~[]byte](v []*T) (res pgtype.ByteaArray) {
	if v == nil {
		res.Status = pgtype.Null

		return
	}

	res.Status = pgtype.Present

	if len(v) == 0 {
		return
	}

	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}
	res.Elements = make([]pgtype.Bytea, len(v))

	for i := range v {
		if v[i] == nil || *v[i] == nil {
			res.Elements[i] = pgtype.Bytea{Status: pgtype.Null}

			continue
		}

		res.Elements[i] = pgtype.Bytea{Status: pgtype.Present, Bytes: *v[i]}
	}

	return
}

func ByteaArrayFromPtrFn[T ~[]byte](length int, fn func(i int) *T) (res pgtype.ByteaArray) {
	sl := make([]*T, length)

	for i := range length {
		sl[i] = fn(i)
	}

	return ByteaArrayFromPtr(sl)
}
