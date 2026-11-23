package uuidnumhash

import "github.com/google/uuid"

type Hash = interface {
	Uint64Slice(u []uuid.UUID) []uint64
	Uint64(u uuid.UUID) uint64
}
