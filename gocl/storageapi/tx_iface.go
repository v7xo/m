package storageapi

import (
	"context"

	"rn.m.vo7.io/vo7io/m/gocl/storage"
)

type TxAPI interface {
	API

	// IsTxSupported indicates whether transactions are supported by this API instance.
	IsTxSupported() bool
	// HasTx indicates whether this API instance have open transaction (see BeginTx).
	HasTx() (has bool, iso storage.IsolationLevel)

	BeginTx(ctx context.Context) error
	BeginTxIso(ctx context.Context, iso storage.IsolationLevel) error
	BeginTxIsoAccess(ctx context.Context, iso storage.IsolationLevel, access storage.AccessMode) error
	CommitTx(ctx context.Context) error
	RollbackTx(ctx context.Context) error

	// ExecTxFn executes function fn in storage transaction.
	// If fn returns error then transaction will be rolled back, otherwise transaction will be committed.
	ExecTxFn(ctx context.Context, fn func() error) error
	// ExecTxIsoFn executes function fn in storage transaction.
	// If fn returns error then transaction will be rolled back, otherwise transaction will be committed.
	ExecTxIsoFn(ctx context.Context, iso storage.IsolationLevel, fn func() error) error
	// ExecTxIsoAccessFn executes function fn in storage transaction.
	// If fn returns error then transaction will be rolled back, otherwise transaction will be committed.
	ExecTxIsoAccessFn(ctx context.Context, iso storage.IsolationLevel, access storage.AccessMode, fn func() error) error
}
