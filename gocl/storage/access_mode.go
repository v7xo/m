package storage

import "github.com/jackc/pgx/v5"

type AccessMode byte

//nolint:iotamixing
const (
	AccessModeReadWrite AccessMode = iota
	AccessModeReadOnly

	AccessModeDefault = AccessModeReadWrite
)

func (v AccessMode) ToPg() pgx.TxAccessMode {
	switch v {
	case AccessModeReadWrite:
		return pgx.ReadWrite
	case AccessModeReadOnly:
		return pgx.ReadOnly
	default:
		return pgx.ReadWrite
	}
}
