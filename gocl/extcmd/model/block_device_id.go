package model

import (
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type BlockDeviceID interface {
	IsBlockDeviceID()

	BackingDevice() BlockDeviceID
	IsMapper() bool
	MapperName() string
	Path() string
	String() string
	Type() BlockDeviceTypeCode
}

type BlockDeviceIDList []BlockDeviceID

func (v BlockDeviceIDList) ToPathSlice() []string {
	if v == nil {
		return nil
	}

	res := make([]string, len(v))

	for i := range len(v) {
		if v[i] != nil {
			res[i] = v[i].Path()
		}
	}

	return res
}

func (v BlockDeviceIDList) String() string {
	if v == nil {
		return placeholder.Nil
	} else if len(v) == 0 {
		return placeholder.Empty
	}

	sb := strings.Builder{}

	sb.WriteByte('[')

	sb.WriteString("count: ")
	sb.WriteString(stringsext.FormatInt(len(v)))

	for i := range len(v) {
		sb.WriteString(", ")

		if i == 32 {
			sb.WriteRune(stringsext.TruncateElipsisSym)

			break
		}

		sb.WriteString(stringsext.Stringer(v[i]))
	}

	sb.WriteByte(']')

	return sb.String()
}

var _ BlockDeviceID = blockDeviceID{}

type blockDeviceID struct {
	b           BlockDeviceID
	path        string
	mapperName  string
	typ         BlockDeviceTypeCode
	initialized bool
}

func (v blockDeviceID) IsBlockDeviceID()             {}
func (v blockDeviceID) BackingDevice() BlockDeviceID { return v.b }
func (v blockDeviceID) IsMapper() bool               { return v.mapperName != "" }
func (v blockDeviceID) MapperName() string           { return v.mapperName }
func (v blockDeviceID) Path() string                 { return v.path }
func (v blockDeviceID) Type() BlockDeviceTypeCode    { return v.typ }

func (v blockDeviceID) String() string {
	if !v.initialized {
		return placeholder.NotInitialized
	}

	sb := strings.Builder{}

	sb.WriteByte('(')

	sb.WriteString("path: ")
	sb.WriteString(placeholder.UnspecifiedIfEmpty(v.path))

	sb.WriteString(", type: ")
	sb.WriteString(v.typ.String())

	if v.b != nil {
		sb.WriteString(", backing: ")
		sb.WriteString(v.b.String())
	}

	sb.WriteByte(')')

	return sb.String()
}

// Deprecated: use [NewBlockDeviceIDV2] instead.
func NewBlockDeviceID(path string, typ BlockDeviceTypeCode, opts ...BlockDeviceIDOptionFn) BlockDeviceID {
	res := blockDeviceID{initialized: true, path: path, typ: typ}

	for i := range len(opts) {
		if opts[i] == nil {
			continue
		}

		opts[i](&res)
	}

	return res
}

func NewBlockDeviceIDV2(opts ...BlockDeviceIDOptionFn) BlockDeviceID {
	res := blockDeviceID{initialized: true, typ: BlockDeviceTypeCodeUnspecified}

	for i := range len(opts) {
		if opts[i] == nil {
			continue
		}

		opts[i](&res)
	}

	return res
}

type BlockDeviceIDOptionFn func(*blockDeviceID)

func BlockDeviceIDOption() BlockDeviceIDOptionFn { return func(bdi *blockDeviceID) {} }

func (o BlockDeviceIDOptionFn) MapperName(val string) BlockDeviceIDOptionFn {
	return func(bdi *blockDeviceID) {
		o(bdi)

		bdi.mapperName = val
		bdi.path = "/dev/mapper/" + val
	}
}

func (o BlockDeviceIDOptionFn) Path(val string) BlockDeviceIDOptionFn {
	return func(bdi *blockDeviceID) {
		o(bdi)

		bdi.path = val

		if cut, ok := strings.CutPrefix(val, "/dev/mapper/"); ok {
			bdi.mapperName = cut
		} else {
			bdi.mapperName = ""
		}
	}
}

func (o BlockDeviceIDOptionFn) Type(val BlockDeviceTypeCode) BlockDeviceIDOptionFn {
	return func(bdi *blockDeviceID) { o(bdi); bdi.typ = val }
}

func (o BlockDeviceIDOptionFn) Backing(val BlockDeviceID) BlockDeviceIDOptionFn {
	return func(bdi *blockDeviceID) { o(bdi); bdi.b = val }
}
