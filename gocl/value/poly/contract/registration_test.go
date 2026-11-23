//go:build test_unit

package contract_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/value/poly"
	"rn.m.vo7.io/vo7io/m/gocl/value/poly/contract"
)

type Kitchenware interface {
	IsKitchenware()
}

type BlackKnife struct {
	_ contract.Name[NamespaceHelium, LocalNameNeon]

	Material string `yaml:"material"`
}

func (BlackKnife) IsKitchenware() {}

type RedOwen struct {
	_ contract.Name[NamespaceRadon, LocalNameNeon]

	MaxTemp int `yaml:"max_temp"`
}

func (RedOwen) IsKitchenware() {}

type GreenSpatula struct {
	_ contract.Name[contract.AsType, LocalNameNeon] `polynamespace:",snake_case"`

	Weight float32 `yaml:"weight"`
}

func (GreenSpatula) IsKitchenware() {}

func Test_Registration_Happy(t *testing.T) {
	thelper.Pre(t)

	hlpRegister[BlackKnife](t)
	hlpRegister[GreenSpatula](t)
	hlpRegister[RedOwen](t)

	n, err := contract.NameFor[BlackKnife]()
	require.NoError(t, err)
	require.Equal(t, "rn.m.vo7.io/vo7io/m/gocl/value/poly/contract_test/local_name_neon", n)

	n, err = contract.NameFor[GreenSpatula]()
	require.NoError(t, err)
	require.Equal(t, "green_spatula/local_name_neon", n)

	n, err = contract.NameFor[RedOwen]()
	require.NoError(t, err)
	require.Equal(t, "namespace_radon/local_name_neon", n)
}

func Test_YAMLMarshal_Happy(t *testing.T) {
	thelper.Pre(t)

	hlpRegister[BlackKnife](t)

	hlpRegister[GreenSpatula](t)
	hlpRegister[RedOwen](t)

	type house struct {
		Color       string                 `yaml:"color"`
		Square      int64                  `yaml:"square"`
		Kitchenware []poly.Of[Kitchenware] `yaml:"kitchenware"`
	}

	fixture := house{
		Color:  "yellow",
		Square: 155,
		Kitchenware: []poly.Of[Kitchenware]{
			{Val: BlackKnife{Material: "wood"}},
			{Val: GreenSpatula{Weight: .31}},
			{Val: RedOwen{MaxTemp: 350}},
		},
	}

	buf := bytes.NewBuffer(nil)
	enc := yaml.NewEncoder(buf)

	enc.SetIndent(2)

	err := enc.Encode(fixture)
	require.NoError(t, err)

	yamlPayloadExpected := `color: yellow
square: 155
kitchenware:
  - .data:
      material: wood
    .type: rn.m.vo7.io/vo7io/m/gocl/value/poly/contract_test/local_name_neon
  - .data:
      weight: 0.31
    .type: green_spatula/local_name_neon
  - .data:
      max_temp: 350
    .type: namespace_radon/local_name_neon
`
	require.Equal(t, yamlPayloadExpected, buf.String())

	dec := yaml.NewDecoder(buf)
	unmarshaled := house{}

	dec.KnownFields(true)

	err = dec.Decode(&unmarshaled)
	require.NoError(t, err)
	require.Equal(t, fixture, unmarshaled)
}

func hlpRegister[T any](t *testing.T) {
	t.Helper()

	err := poly.Register(poly.Options[T]())
	require.NoError(t, err)
}
