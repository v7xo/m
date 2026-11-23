package storage

import "strings"

type StorageType string

func (f StorageType) IsTarantool() bool {
	return strings.EqualFold(string(TarantoolStorageType), string(f))
}

func (f StorageType) IsPostgreSQL() bool {
	return strings.EqualFold(string(PostgreSQLStorageType), string(f))
}

func (f StorageType) IsMySQL() bool {
	return strings.EqualFold(string(StorageTypeMySQL), string(f))
}

func (f StorageType) IsUnspecified() bool {
	return len(f) == 0
}

func (f StorageType) String() string {
	return string(f)
}
