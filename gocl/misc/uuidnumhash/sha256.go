package uuidnumhash

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"

	"github.com/google/uuid"
)

type sha256Impl struct {
	h    hash.Hash
	hBuf []byte
}

func (r *sha256Impl) Uint64Slice(val []uuid.UUID) (res []uint64) {
	if len(val) == 0 {
		return
	}

	res = make([]uint64, len(val))
	h := r.h
	hBuf := r.hBuf

	for i := range val {
		h.Reset()

		hBuf = hBuf[:0]

		h.Write(val[i][:])

		hBuf = h.Sum(hBuf)
		res[i] = binary.BigEndian.Uint64(hBuf[24:]) //nolint:revive
	}

	return
}

func (r *sha256Impl) Uint64(val uuid.UUID) (res uint64) {
	if len(val) == 0 {
		return
	}

	h := r.h
	hBuf := r.hBuf

	h.Reset()

	hBuf = hBuf[:0]

	h.Write(val[:])

	hBuf = h.Sum(hBuf)
	res = binary.BigEndian.Uint64(hBuf[24:]) //nolint:revive

	return
}

func NewSHA256() Hash {
	return &sha256Impl{
		h:    sha256.New(),
		hBuf: make([]byte, 0, 32), //nolint:revive,mnd
	}
}
