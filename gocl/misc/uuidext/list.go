package uuidext

import (
	"bytes"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
)

type List uuid.UUIDs

func (v List) Uniq() map[uuid.UUID]struct{} {
	if v == nil {
		return nil
	}

	ret := make(map[uuid.UUID]struct{}, len(v))

	for i := range len(v) {
		ret[v[i]] = struct{}{}
	}

	return ret
}

func (v List) String() string {
	if v == nil {
		return placeholder.Nil
	}

	sb := &strings.Builder{}

	sb.WriteByte('[')

	sb.WriteString("count: ")
	sb.WriteString(strconv.Itoa(len(v)))

	const maxCount = 32

	for i := range len(v) {
		sb.WriteString(", ")

		if i < maxCount {
			sb.WriteString(v[i].String())

			continue
		}

		sb.WriteRune(stringsext.TruncateElipsisSym)

		break
	}

	sb.WriteByte(']')

	return sb.String()
}

func (v *List) FromString[S ~[]E, E ~string](val S) (err error) {
	if v == nil {
		return nil
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to parse '%s' from '%s'; %w", typestring.TypeStringForExpr(v), typestring.TypeStringForExpr(val), err)
		}
	}()

	if val == nil {
		*v = nil

		return nil
	}

	cp := make(List, len(val))

	for i := range len(val) {
		val, err := uuid.Parse(string(val[i]))
		if err != nil {
			return fmt.Errorf("failed to parse item '%s' at '%d' position; %w", stringsext.TruncateDefault(string(val[i])), i, err)
		}

		cp[i] = val
	}

	*v = cp

	return nil
}

func (v List) Sort(desc bool) {
	if desc {
		slices.SortFunc(v, func(a, b uuid.UUID) int { return bytes.Compare(b[:], a[:]) })

		return
	}

	slices.SortFunc(v, func(a, b uuid.UUID) int { return bytes.Compare(a[:], b[:]) })
}

func (v List) BinarySearch(val uuid.UUID, desc bool) (int, bool) {
	if desc {
		return slices.BinarySearchFunc(v, val, func(a, b uuid.UUID) int { return bytes.Compare(b[:], a[:]) })
	}

	return slices.BinarySearchFunc(v, val, func(a, b uuid.UUID) int { return bytes.Compare(a[:], b[:]) })
}
