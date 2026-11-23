package factory

import (
	"context"
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
)

type APIConstructorFn[TAPI storageapi.SQLAPI] func(ctx context.Context, b storageapi.SQLAPI) (TAPI, error)

func CustomAPIConstructor[TAPI storageapi.SQLAPI, TCustom storageapi.CustomSQLAPI[TAPI]](_ context.Context, b storageapi.SQLAPI) (res TAPI, err error) {
	custom := TCustom(&res)
	if err = custom.SetBaseAPI(b); err != nil {
		err = fmt.Errorf("error while setting base storage API on instance of custom storage API; %w", err)
	}

	return
}

func BaseAPIConstructor(_ context.Context, baseAPI storageapi.SQLAPI) (storageapi.SQLAPI, error) {
	return baseAPI, nil
}
