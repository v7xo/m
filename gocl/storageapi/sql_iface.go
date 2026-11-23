package storageapi

import (
	"context"

	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/storage/query"
)

var _ = errs.ErrNotFound

type (
	SQLExecutor interface {
		// ExecSQL executes specified query and put results into specified destination dst.
		ExecSQL(ctx context.Context, query query.Named, dst any, args ...any) error

		// ExecSQLNonQuery executes specified query without any results.
		//
		// This method useful for cases when query to be executed doesn't imply a rowset as result.
		ExecSQLNonQuery(ctx context.Context, query query.Named, args ...any) error

		// ExecSQLOneRow executes specified query and expects only one row as a result of execution.
		//
		// If result rowset is empty then [errs.ErrNotFound] returned.
		//
		// If result rowset contains more than one row then error returned.
		ExecSQLOneRow(ctx context.Context, query query.Named, dst any, args ...any) error
	}

	// SQLAPI represents base specification of SQL storage API with common-use methods.
	//
	// nolint:interfacebloat
	SQLAPI interface {
		SQLExecutor
		TxAPI
	}

	SQLAPIPointer[T any] interface {
		SQLAPI
		*T
	}

	// CustomSQLAPI is generic type constraint for custom storage API implementations.
	CustomSQLAPI[TImplementation SQLAPI] interface {
		SetBaseAPI(v SQLAPI) error

		*TImplementation
	}
)
