package model

import (
	"errors"
)

type CryptCipherParams struct {
	encoded string
}

func MustDecodeCryptCipherParams(val string) CryptCipherParams {
	res, err := DecodeCryptCipherParams(val)
	if err != nil {
		panic(err)
	}

	return res
}

func DecodeCryptCipherParams(val string) (res CryptCipherParams, err error) {
	if val == "" {
		return res, errors.New("encoded string is empty")
	}

	return CryptCipherParams{encoded: val}, nil
}

func (v CryptCipherParams) Encode() string { return v.encoded }

func (v CryptCipherParams) Empty() bool { return v.encoded == "" }

func (v CryptCipherParams) String() string { return v.encoded }
