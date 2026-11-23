package uuidnumhash

import (
	"github.com/google/uuid"
)

func SHA256Uint64HashMapperSlice[TIn any, TOut any](
	in []TIn,
	mapper func(pos int, in TIn, h Hash) TOut,
) (res []TOut) {
	if len(in) == 0 {
		return
	}

	res = make([]TOut, len(in))
	h := NewSHA256()

	for i := range in {
		res[i] = mapper(i, in[i], h)
	}

	return
}

func SHA256Uint64HashSlice(val []uuid.UUID) (res []uint64) {
	if len(val) == 0 {
		return
	}

	res = NewSHA256().Uint64Slice(val)

	return
}

func SHA256Uint64Hash(val uuid.UUID) (res uint64) {
	res = NewSHA256().Uint64(val)

	return
}

func SHA256Uint64HashMapKey[TVal any](val map[uuid.UUID]TVal) (res map[uint64]TVal) {
	if len(val) == 0 {
		return
	}

	res = make(map[uint64]TVal, len(val))
	h := NewSHA256()

	for k, v := range val {
		res[h.Uint64(k)] = v
	}

	return
}
