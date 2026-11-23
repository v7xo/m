package uuidext

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func RandInt64() (int64, error) {
	randBytes, err := uuid.NewRandom()
	if err != nil {
		return 0, fmt.Errorf("failed to generate new random UUID; err: %w", err)
	}

	return int64(binary.LittleEndian.Uint64(randBytes[8:])), nil
}

func Rand4BHex() (string, error) {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate new random UUID; err: %w", err)
	}

	//nolint:revive
	return hex.EncodeToString(newUUID[:4]), nil
}

func MustRand4BHex() string {
	res, err := Rand4BHex()
	if err != nil {
		panic(err)
	}

	return res
}

func MustRand8BHex() string {
	res, err := Rand8BHex()
	if err != nil {
		panic(err)
	}

	return res
}

func Rand8BHex() (string, error) {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate new random UUID; err: %w", err)
	}

	//nolint:revive,mnd
	return hex.EncodeToString(newUUID[:8]), nil
}

func Obfuscate(v uuid.UUID) (res uuid.UUID) {
	if v == uuid.Nil {
		return
	}

	h := sha256.New()

	h.Write(v[:])

	hash := h.Sum(nil)

	copy(res[:], hash)

	return
}

func RequestIDHex() (string, error) {
	endianess := binary.BigEndian

	b := make([]byte, 4, 8)
	endianess.PutUint32(b, uint32(time.Now().Unix()))

	rnd := make([]byte, 4)

	_, err := rand.Read(rnd)
	if err != nil {
		return "", fmt.Errorf("failed to generate random %d bytes; err: %w", len(rnd), err)
	}

	b = append(b, rnd...)

	const groupLen int = 2

	if len(b) <= groupLen {
		return hex.EncodeToString(b), nil
	}

	enc := make([]byte, 0, len(b)*2+(len(b)/groupLen-1))

	i := 0
	for {
		end := min(i+groupLen, len(b))

		if len(enc) > 0 {
			enc = append(enc, '_')
		}

		enc = hex.AppendEncode(enc, b[i:end])
		i += groupLen

		if i < len(b) {
			continue
		}

		break
	}

	return string(enc), nil
}

func Slice(u uuid.UUID) []byte { return u[:] }

func MustNewV7() uuid.UUID {
	res, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}

	return res
}
