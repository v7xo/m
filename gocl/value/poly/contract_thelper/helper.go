package contract_thelper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	clstring "rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/value/poly/contract"
)

func NewDescriber(t *testing.T, ctx context.Context, opts ...contract.DescriberOptionFn) contract.Describer {
	t.Helper()

	res, err := contract.NewDescriber(fopt.PrependFromContext(ctx, opts...))
	require.NoError(t, err, "unexpected error while creating instance of '%s'", clstring.TypeStringFor[contract.Describer]())
	require.NotNil(t, res, "unexpected nil as a result of call to constructor of '%s'", clstring.TypeStringFor[contract.Describer]())

	return res
}

func NewRegistrar(t *testing.T, ctx context.Context, opts ...contract.RegistrarOptionFn) contract.Registrar {
	t.Helper()

	res, err := contract.NewRegistrar(fopt.PrependFromContext(ctx, opts...))
	require.NoError(t, err, "unexpected error while creating instance of '%s'", clstring.TypeStringFor[contract.Registrar]())
	require.NotNil(t, res, "unexpected nil as a result of call to constructor of '%s'", clstring.TypeStringFor[contract.Registrar]())

	return res
}
