package storageapi

import (
	"context"
	"errors"
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
)

func ExecTxFn(ctx context.Context, api TxAPI, fn func() error) error {
	return execTxFn(ctx, api, storage.IsolationLevelUnspecified, storage.AccessModeDefault, fn)
}

func ExecTxIsoFn(ctx context.Context, api TxAPI, iso storage.IsolationLevel, fn func() error) error {
	return execTxFn(ctx, api, iso, storage.AccessModeDefault, fn)
}

func ExecTxIsoAccessFn(ctx context.Context, api TxAPI, iso storage.IsolationLevel, access storage.AccessMode, fn func() error) error {
	return execTxFn(ctx, api, iso, access, fn)
}

func execTxFn(ctx context.Context, api TxAPI, iso storage.IsolationLevel, access storage.AccessMode, fn func() error) (err error) {
	log := api.Log()

	switch {
	case iso.IsUnspecified():
		iso = storage.IsolationLevelInheritedOrDefault
	}

	if err = api.BeginTxIsoAccess(ctx, iso, access); err != nil {
		err = fmt.Errorf("failed to begin transaction; err: %w", err)

		log.Err(ctx, err).Send()

		return
	}

	err = fn()
	if err != nil {
		err = errors.Join(err, errs.ErrTxRollingback)

		if rollbackErr := api.RollbackTx(ctx); rollbackErr != nil {
			err = errors.Join(err, fmt.Errorf("error while rolling back transaction; err: %w", rollbackErr))

			log.Err(ctx, err).Send()
		}

		return
	}

	if commitErr := api.CommitTx(ctx); commitErr != nil {
		err = fmt.Errorf("error while committing transaction; err: %w", commitErr)

		log.Err(ctx, err).Send()
	}

	return
}
