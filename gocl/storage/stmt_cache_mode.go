package storage

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type StmtCacheMode string

const (
	StmtCacheModeStorageTypeDefault = StmtCacheMode("default")
	StmtCacheModeDescribe           = StmtCacheMode("describe")
	StmtCacheModeUnspecified        = StmtCacheMode("")
)

func (v StmtCacheMode) String() string {
	return string(v)
}

func (v StmtCacheMode) IsUnspecified() bool {
	return v == ""
}

func (v StmtCacheMode) IsStorageTypeDefault() bool {
	return strings.EqualFold(string(v), string(StmtCacheModeStorageTypeDefault))
}

func (v StmtCacheMode) IsDescribe() bool {
	return strings.EqualFold(string(v), string(StmtCacheModeDescribe))
}

func (v StmtCacheMode) AsPgxQueryExecMode() (pgx.QueryExecMode, error) {
	switch {
	case v.IsDescribe():
		return pgx.QueryExecModeCacheDescribe, nil
	case v.IsStorageTypeDefault():
		return DefaultPostgreSQLStmtCacheMode.AsPgxQueryExecMode()
	case v.IsUnspecified():
		return 0, errors.New("statement cache mode not speicifed")
	default:
		return 0, fmt.Errorf("unknown statement cache mode; mode: %s", v)
	}
}
