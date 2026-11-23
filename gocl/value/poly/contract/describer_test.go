//go:build test_unit

package contract_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/value/poly/contract"
	"rn.m.vo7.io/vo7io/m/gocl/value/poly/contract_thelper"
)

type (
	NamespaceHydrogen struct {
		hydrogen contract.Namespace[NamespaceHelium]
	}
	NamespaceHelium struct {
		helium contract.Namespace[contract.AsPackage]
	}
	NamespaceRadon struct {
		radon contract.Namespace[contract.AsType] `polynamespace:",snake_case"`
	}
	LocalNameNeon struct {
		neon contract.LocalName[contract.AsType] `polylocal:",snake_case"`
	}

	Moscow struct {
		moscow contract.Name[NamespaceHydrogen, LocalNameNeon]
	}
	Boston struct {
		boston contract.Name[contract.AsTag, LocalNameNeon] `polynamespace:"north-america/,snake_case"`
	}
	CapeTown struct {
		capeTown contract.Name[contract.AsPackage, LocalNameNeon]
	}
	Joburg struct {
		joburg contract.Name[contract.AsType, LocalNameNeon] `polynamespace:",snake_case"`
	}
)

func (LocalNameNeon) IsPolyContractLocalNameLocator()     {}
func (NamespaceHelium) IsPolyContractNamespaceLocator()   {}
func (NamespaceHydrogen) IsPolyContractNamespaceLocator() {}
func (NamespaceRadon) IsPolyContractNamespaceLocator()    {}

func Test_Describer_NRedirect_LRedirect(t *testing.T) {
	ctx := thelper.Pre(t)

	describer := contract_thelper.NewDescriber(t, ctx)

	contract, err := describer.Describe(ctx, reflect.TypeFor[Moscow]())
	require.NoError(t, err)
	require.Equal(t, "rn.m.vo7.io/vo7io/m/gocl/value/poly/contract_test/local_name_neon", contract.Name)
	require.Equal(t, "rn.m.vo7.io/vo7io/m/gocl/value/poly/contract_test/", contract.Namespace)
	require.Equal(t, "local_name_neon", contract.LocalName)

	t.Logf("\n[INFO] contract:\n%s", contract.StringMore())
}

func Test_Describer_NAsTag_LRedirect(t *testing.T) {
	ctx := thelper.Pre(t)

	describer := contract_thelper.NewDescriber(t, ctx)

	contract, err := describer.Describe(ctx, reflect.TypeFor[Boston]())
	require.NoError(t, err)
	require.Equal(t, "north_america/local_name_neon", contract.Name)
	require.Equal(t, "north_america/", contract.Namespace)
	require.Equal(t, "local_name_neon", contract.LocalName)

	t.Logf("\n[INFO] contract:\n%s", contract.StringMore())
}

func Test_Describer_NAsPackage_LRedirect(t *testing.T) {
	ctx := thelper.Pre(t)

	describer := contract_thelper.NewDescriber(t, ctx)

	contract, err := describer.Describe(ctx, reflect.TypeFor[CapeTown]())
	require.NoError(t, err)
	require.Equal(t, "rn.m.vo7.io/vo7io/m/gocl/value/poly/contract_test/local_name_neon", contract.Name)
	require.Equal(t, "rn.m.vo7.io/vo7io/m/gocl/value/poly/contract_test/", contract.Namespace)
	require.Equal(t, "local_name_neon", contract.LocalName)

	t.Logf("\n[INFO] contract:\n%s", contract.StringMore())
}

func Test_Describer_NAsType_LRedirect(t *testing.T) {
	ctx := thelper.Pre(t)

	describer := contract_thelper.NewDescriber(t, ctx)

	contract, err := describer.Describe(ctx, reflect.TypeFor[Joburg]())
	require.NoError(t, err)
	require.Equal(t, "joburg/local_name_neon", contract.Name)
	require.Equal(t, "joburg/", contract.Namespace)
	require.Equal(t, "local_name_neon", contract.LocalName)

	t.Logf("\n[INFO] contract:\n%s", contract.StringMore())
}
