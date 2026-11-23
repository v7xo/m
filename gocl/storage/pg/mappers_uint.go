package pg

import (
	"fmt"
	"math"

	"github.com/jackc/pgtype"

	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

func Uint64SliceToPgInt8(v []uint64) (pgtype.Int8Array, error) {
	var res pgtype.Int8Array

	res.Status = pgtype.Present
	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}

	res.Elements = make([]pgtype.Int8, len(v))

	for i := range v {
		valAsInt, err := Uint64ToInt64(v[i])
		if err != nil {
			return pgtype.Int8Array{}, fmt.Errorf("failed to map uint64 slice item; item pos.: %d; err: %w", i, err)
		}

		res.Elements[i] = pgtype.Int8{Status: pgtype.Present, Int: valAsInt}
	}

	return res, nil
}

func Uint64PtrSliceToPgInt8(v []*uint64) (pgtype.Int8Array, error) {
	var res pgtype.Int8Array

	res.Status = pgtype.Present
	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}

	res.Elements = make([]pgtype.Int8, len(v))

	for i := range v {
		if v[i] == nil {
			res.Elements[i] = pgtype.Int8{Status: pgtype.Null}

			continue
		}

		valAsInt, err := Uint64ToInt64(*v[i])
		if err != nil {
			return pgtype.Int8Array{}, fmt.Errorf("failed to map uint64 slice item; item pos.: %d; err: %w", i, err)
		}

		res.Elements[i] = pgtype.Int8{Status: pgtype.Present, Int: valAsInt}
	}

	return res, nil
}

func Uint32SliceToPgInt4(v []uint32) (pgtype.Int4Array, error) {
	var res pgtype.Int4Array

	res.Status = pgtype.Present
	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}

	res.Elements = make([]pgtype.Int4, len(v))

	for i := range v {
		valAsInt, err := Uint32ToInt32(v[i])
		if err != nil {
			return pgtype.Int4Array{}, fmt.Errorf("failed to map uint64 slice item; item pos.: %d; err: %w", i, err)
		}

		res.Elements[i] = pgtype.Int4{Status: pgtype.Present, Int: valAsInt}
	}

	return res, nil
}

func Uint32PtrSliceToPgInt4(v []*uint32) (pgtype.Int4Array, error) {
	var res pgtype.Int4Array

	res.Status = pgtype.Present
	res.Dimensions = []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}}

	res.Elements = make([]pgtype.Int4, len(v))

	for i := range v {
		if v[i] == nil {
			res.Elements[i] = pgtype.Int4{Status: pgtype.Null}

			continue
		}

		valAsInt, err := Uint32ToInt32(*v[i])
		if err != nil {
			return pgtype.Int4Array{}, fmt.Errorf("failed to map uint32 slice item; item pos.: %d; err: %w", i, err)
		}

		res.Elements[i] = pgtype.Int4{Status: pgtype.Present, Int: valAsInt}
	}

	return res, nil
}

func Uint64ToInt64(val uint64) (int64, error) {
	if uint64(math.MaxInt64) < val {
		return 0, errs.BadArgsWrapWith(fmt.Errorf("uint64 value can't be mapped to int64 value, overflow; value: %d", val))
	}

	return int64(val), nil
}

func Uint32ToInt32(val uint32) (int32, error) {
	if uint32(math.MaxInt32) < val {
		return 0, errs.BadArgsWrapWith(fmt.Errorf("uint32 value can't be mapped to int32 value, overflow; value: %d", val))
	}

	return int32(val), nil
}
