package contract

import (
	"fmt"
	"math/bits"
	"reflect"
	"strings"

	clstrings "rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/value/enum"
)

// Contract represents info describing poly-contract.
type Contract struct {
	NamespaceLocator *AttrLocator
	LocalNameLocator *AttrLocator
	Type             reflect.Type
	LocalName        string
	Namespace        string
	Name             string
	NameAttr         reflect.StructField
}

func (v Contract) String() string {
	sb := &strings.Builder{}

	sb.WriteByte('(')

	sb.WriteString("name: ")
	sb.WriteString(placeholder.UnspecifiedIfEmpty(v.Name))

	sb.WriteByte(')')

	return sb.String()
}

func (v Contract) StringMore() string {
	sb := &strings.Builder{}

	sb.WriteByte('(')

	sb.WriteString("name: ")
	sb.WriteString(placeholder.UnspecifiedIfEmpty(v.Name))

	sb.WriteString(", local name: ")
	sb.WriteString(placeholder.UnspecifiedIfEmpty(v.LocalName))

	if v.NamespaceLocator != nil {
		sb.WriteString(",\nnamespace locator:\n")
		sb.WriteString(v.NamespaceLocator.String())
	}

	if v.LocalNameLocator != nil {
		sb.WriteString(",\nlocal name locator:\n")
		sb.WriteString(v.LocalNameLocator.String())
	}

	sb.WriteByte(')')

	return sb.String()
}

type AttrLocator struct {
	Backward *AttrLocator
	Forward  *AttrLocator

	AttrTag      *Tag
	LocalNameTag *Tag
	NamespaceTag *Tag
	Origin       reflect.Type
	Raw          string
	AttrField    reflect.StructField
	Kind         LocatorKind
	AttrKind     AttrKind
}

func (v *AttrLocator) String() string {
	if v == nil {
		return placeholder.Nil
	}

	sb := &strings.Builder{}

	sb.WriteByte('[')

	i := 0

	c := v
	for c != nil {
		if i != 0 {
			sb.WriteString(",\n<- ")
		}

		sb.WriteByte('(')

		sb.WriteString("kind: ")
		sb.WriteString(c.Kind.String())

		if c.Raw != "" {
			sb.WriteString(", raw: ")
			sb.WriteString(c.Raw)
		}

		if c.AttrField.Type != nil && c.AttrField.Name != "_" && c.AttrField.Name != "" {
			sb.WriteString(", field: ")
			sb.WriteString(c.AttrField.Name)
		}

		if c.AttrTag != nil {
			sb.WriteString(", tag: ")
			sb.WriteString(c.AttrTag.String())
		}

		if c.Origin != nil {
			sb.WriteString(", origin: ")
			sb.WriteString(clstrings.TypeString(c.Origin))
		}

		sb.WriteByte(')')

		c = c.Backward
		i++
	}

	sb.WriteByte(']')

	return sb.String()
}

type TagFlag int64

const (
	TagFlagSnakeCase TagFlag = 1 << iota
	TagFlagKebabCase
	TagFlagCamelCase

	TagFlagNone     TagFlag = 0
	TagFlagCaseMask         = TagFlagKebabCase | TagFlagSnakeCase | TagFlagCamelCase
)

var _ = enum.MustRegisterFlag(
	map[TagFlag]string{
		TagFlagNone:      "",
		TagFlagSnakeCase: "snake_case",
		TagFlagKebabCase: "kebab_case",
		TagFlagCamelCase: "camel_case",
	})

func (v TagFlag) String() string {
	return enum.String(v)
}

func ParseTagFlag(val string) (TagFlag, error) { return enum.Parse[TagFlag](val) }

func (v TagFlag) Validate() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("invalid use of flags '%s'; %w", v, err)
		}
	}()

	if bits.OnesCount64(uint64(v&TagFlagCaseMask)) > 1 {
		return fmt.Errorf("only one of '%s' can be used", v&TagFlagCaseMask)
	}

	return nil
}

type Tag struct {
	Raw   string
	Value string
	Flags TagFlag
}

func (v *Tag) String() string {
	if v == nil {
		return placeholder.Nil
	}

	buf := make([]byte, 0, 64)

	buf = append(buf, '(')

	buf = append(buf, "flags: "...)
	buf = append(buf, v.Flags.String()...)

	if v.Value != "" {
		buf = append(buf, ", value: "...)
		buf = append(buf, v.Value...)
	}

	if v.Raw != "" {
		buf = append(buf, ", raw: "...)
		buf = append(buf, v.Raw...)
	}

	buf = append(buf, ')')

	return string(buf)
}
