//go:build test_unit

package enum_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	clhlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/value/enum"
)

func Test_Enum_YAML(t *testing.T) {
	clhlp.Pre(t)

	type Enum1 int32

	const (
		enum1Unspecified Enum1 = iota
		enum1A
		enum1B
	)

	err := enum.Register[Enum1](
		map[Enum1]string{
			enum1Unspecified: "",
			enum1A:           "a",
			enum1B:           "b",
		},
	)
	require.NoError(t, err)

	type dataT struct {
		Foo string         `yaml:"foo"`
		Bar enum.Of[Enum1] `yaml:"bar"`
	}

	// ------ case 1 ------
	data := dataT{
		Foo: "foo",
		Bar: enum.Of[Enum1]{Val: enum1B},
	}

	dataYAML, err := yaml.Marshal(data)
	require.NoError(t, err)

	dataYAMLExpectation := `foo: foo
bar: b
`

	require.Equal(t, dataYAMLExpectation, string(dataYAML))

	var dataUnmarshaled dataT

	err = yaml.Unmarshal(dataYAML, &dataUnmarshaled)
	require.NoError(t, err)

	require.Equal(t, data, dataUnmarshaled)

	// ------ case 2 (unknown name) ------
	data = dataT{
		Foo: "foo",
		Bar: enum.Of[Enum1]{Val: 20},
	}

	dataYAML, err = yaml.Marshal(data)
	require.NoError(t, err)

	dataYAMLExpectation = `foo: foo
bar: <unknown(20)>
`

	require.Equal(t, dataYAMLExpectation, string(dataYAML))

	dataUnmarshaled = dataT{}

	err = yaml.Unmarshal(dataYAML, &dataUnmarshaled)
	require.NoError(t, err)

	require.Equal(t, data, dataUnmarshaled)

	// ------ case 3 (empty, unspecified) ------
	data = dataT{
		Foo: "foo",
		Bar: enum.Of[Enum1]{},
	}

	dataYAML, err = yaml.Marshal(data)
	require.NoError(t, err)

	dataYAMLExpectation = `foo: foo
bar: ""
`

	require.Equal(t, dataYAMLExpectation, string(dataYAML))

	dataUnmarshaled = dataT{}

	err = yaml.Unmarshal(dataYAML, &dataUnmarshaled)
	require.NoError(t, err)

	require.Equal(t, data, dataUnmarshaled)
}

func Test_FlagEnum_YAML(t *testing.T) {
	clhlp.Pre(t)

	type EnumA int32

	const (
		EnumA1 EnumA = 1 << iota
		EnumA2
		EnumA4
		EnumA8
		EnumAUnknown1
		EnumAUnknown2

		EnumA0 EnumA = 0
	)

	err := enum.RegisterFlag[EnumA](
		map[EnumA]string{
			// intentionally skip EnummAUnknown1, EnumAUnknown2.
			EnumA0: "",
			EnumA1: "a1",
			EnumA2: "a2",
			EnumA4: "a4",
			EnumA8: "a8",
		},
	)
	require.NoError(t, err)

	type userType struct {
		Foo string         `yaml:"foo"`
		Bar enum.Of[EnumA] `yaml:"bar"`
	}

	// ------ case 1 ------
	data := userType{
		Foo: "foo",
		Bar: enum.Of[EnumA]{Val: EnumA2 | EnumA4},
	}

	dataYAML, err := yaml.Marshal(data)
	require.NoError(t, err)

	dataYAMLExpectation := `foo: foo
bar: a2,a4
`

	require.Equal(t, dataYAMLExpectation, string(dataYAML))

	var dataUnmarshaled userType

	err = yaml.Unmarshal(dataYAML, &dataUnmarshaled)
	require.NoError(t, err)

	require.Equal(t, data, dataUnmarshaled)

	// ------ case 2 (unknown name) ------
	data = userType{
		Foo: "foo",
		Bar: enum.Of[EnumA]{Val: EnumA1 | EnumA4 | EnumAUnknown1 | EnumAUnknown2},
	}

	dataYAML, err = yaml.Marshal(data)
	require.NoError(t, err)

	dataYAMLExpectation = `foo: foo
bar: a1,a4,<unknown(48)>
`

	require.Equal(t, dataYAMLExpectation, string(dataYAML))

	dataUnmarshaled = userType{}

	err = yaml.Unmarshal(dataYAML, &dataUnmarshaled)
	require.NoError(t, err)

	require.Equal(t, data, dataUnmarshaled)

	// ------ case 3 (empty, unspecified) ------
	data = userType{
		Foo: "foo",
		Bar: enum.Of[EnumA]{},
	}

	dataYAML, err = yaml.Marshal(data)
	require.NoError(t, err)

	dataYAMLExpectation = `foo: foo
bar: ""
`

	require.Equal(t, dataYAMLExpectation, string(dataYAML))

	dataUnmarshaled = userType{}

	err = yaml.Unmarshal(dataYAML, &dataUnmarshaled)
	require.NoError(t, err)

	require.Equal(t, data, dataUnmarshaled)
}

func Test_Enum_JSON(t *testing.T) {
	clhlp.Pre(t)

	type Enum1 int32

	const (
		enum1Unspecified Enum1 = iota
		enum1A
		enum1B
	)

	err := enum.Register[Enum1](
		map[Enum1]string{
			enum1Unspecified: "",
			enum1A:           "a",
			enum1B:           "b",
		},
	)
	require.NoError(t, err)

	type dataT struct {
		Foo string         `json:"foo"`
		Bar enum.Of[Enum1] `json:"bar"`
	}

	// ------ case 1 ------
	data := dataT{
		Foo: "foo",
		Bar: enum.Of[Enum1]{Val: enum1B},
	}

	dataJSON, err := json.Marshal(data)
	require.NoError(t, err)

	dataJSONExpectation := `{"foo":"foo","bar":"b"}`

	require.Equal(t, dataJSONExpectation, string(dataJSON))

	var dataUnmarshaled dataT

	err = json.Unmarshal(dataJSON, &dataUnmarshaled)
	require.NoError(t, err)

	require.Equal(t, data, dataUnmarshaled)

	// ------ case 2 (unknown name) ------
	data = dataT{
		Foo: "foo",
		Bar: enum.Of[Enum1]{Val: 20},
	}

	dataJSON, err = json.Marshal(data)
	require.NoError(t, err)

	dataJSONExpectation = `{"foo":"foo","bar":"\u003cunknown(20)\u003e"}`

	require.Equal(t, dataJSONExpectation, string(dataJSON))

	dataUnmarshaled = dataT{}

	err = json.Unmarshal(dataJSON, &dataUnmarshaled)
	require.NoError(t, err)

	require.Equal(t, data, dataUnmarshaled)

	// ------ case 3 (empty, unspecified) ------
	data = dataT{
		Foo: "foo",
		Bar: enum.Of[Enum1]{},
	}

	dataJSON, err = json.Marshal(data)
	require.NoError(t, err)

	dataJSONExpectation = `{"foo":"foo","bar":""}`

	require.Equal(t, dataJSONExpectation, string(dataJSON))

	dataUnmarshaled = dataT{}

	err = json.Unmarshal(dataJSON, &dataUnmarshaled)
	require.NoError(t, err)

	require.Equal(t, data, dataUnmarshaled)
}

func Test_FlagEnum_JSON(t *testing.T) {
	clhlp.Pre(t)

	type EnumA int32

	const (
		EnumA1 EnumA = 1 << iota
		EnumA2
		EnumA4
		EnumA8
		EnumAUnknown1
		EnumAUnknown2

		EnumA0 EnumA = 0
	)

	err := enum.RegisterFlag[EnumA](
		map[EnumA]string{
			// intentionally skip EnummAUnknown1, EnumAUnknown2.
			EnumA0: "",
			EnumA1: "a1",
			EnumA2: "a2",
			EnumA4: "a4",
			EnumA8: "a8",
		},
	)
	require.NoError(t, err)

	type userType struct {
		Foo string         `json:"foo"`
		Bar enum.Of[EnumA] `json:"bar"`
	}

	// ------ case 1a ------
	data := userType{
		Foo: "foo",
		Bar: enum.Of[EnumA]{Val: EnumA2 | EnumA4},
	}

	dataJSON, err := json.Marshal(data)
	require.NoError(t, err)

	dataJSONExpectation := `{"foo":"foo","bar":"a2,a4"}`

	require.Equal(t, dataJSONExpectation, string(dataJSON))

	var dataUnmarshaled userType

	err = json.Unmarshal(dataJSON, &dataUnmarshaled)
	require.NoError(t, err)

	require.Equal(t, data, dataUnmarshaled)

	// ------ case 1b (num) ------
	data = userType{
		Foo: "foo",
		Bar: enum.Of[EnumA]{Val: EnumA8 | EnumA4 | EnumA2},
	}

	dataJSON, err = json.Marshal(data)
	require.NoError(t, err)

	dataJSONExpectation = `{"foo":"foo","bar":"a2,a4,a8"}`

	require.Equal(t, dataJSONExpectation, string(dataJSON))

	err = json.Unmarshal(dataJSON, &dataUnmarshaled)
	require.NoError(t, err)

	require.Equal(t, data, dataUnmarshaled)

	dataJSONExpectation = `{"foo":"foo","bar":14}`

	err = json.Unmarshal(dataJSON, &dataUnmarshaled)
	require.NoError(t, err)

	require.Equal(t, data, dataUnmarshaled)

	// ------ case 2 (unknown name) ------
	data = userType{
		Foo: "foo",
		Bar: enum.Of[EnumA]{Val: EnumA1 | EnumA4 | EnumAUnknown1 | EnumAUnknown2},
	}

	dataJSON, err = json.Marshal(data)
	require.NoError(t, err)

	dataJSONExpectation = `{"foo":"foo","bar":"a1,a4,\u003cunknown(48)\u003e"}`

	require.Equal(t, dataJSONExpectation, string(dataJSON))

	dataUnmarshaled = userType{}

	err = json.Unmarshal(dataJSON, &dataUnmarshaled)
	require.NoError(t, err)

	require.Equal(t, data, dataUnmarshaled)

	// ------ case 3 (empty, unspecified) ------
	data = userType{
		Foo: "foo",
		Bar: enum.Of[EnumA]{},
	}

	dataJSON, err = json.Marshal(data)
	require.NoError(t, err)

	dataJSONExpectation = `{"foo":"foo","bar":""}`

	require.Equal(t, dataJSONExpectation, string(dataJSON))

	dataUnmarshaled = userType{}

	err = json.Unmarshal(dataJSON, &dataUnmarshaled)
	require.NoError(t, err)

	require.Equal(t, data, dataUnmarshaled)
}
