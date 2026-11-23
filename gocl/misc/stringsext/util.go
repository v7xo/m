package stringsext

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unsafe"

	"rn.m.vo7.io/vo7io/m/gocl/misc/ptrof"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
)

type TruncateMode int

//nolint:iotamixing
const (
	MiddleTruncateMode TruncateMode = iota
	StartTruncateMode
	EndTruncateMode

	TruncateElipsisSym    = rune('…')
	TruncateElipsis       = "[…]"
	TruncateElipsisLen    = len(TruncateElipsis)
	TruncateDefaultMaxLen = 2048
)

func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	} else if maxLen <= TruncateElipsisLen {
		return TruncateElipsis
	}

	res := make([]byte, 0, TruncateElipsisLen+maxLen)

	payloadLen := maxLen - TruncateElipsisLen
	res = append(res, []byte(s[:payloadLen/2])...) // nolint:mnd

	var trailer []byte
	if payloadLen%2 == 0 {
		trailer = []byte(s[len(s)-(payloadLen/2):]) // nolint:mnd
	} else {
		trailer = []byte(s[len(s)-(payloadLen/2)-1:]) // nolint:mnd
	}

	res = append(res, TruncateElipsis...)
	res = append(res, trailer...)

	return string(res)
}

func TruncateDefault(s string) string {
	return Truncate(s, TruncateDefaultMaxLen)
}

func TruncateEnd(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen]
}

func SplitSpace(s string) []string {
	const nilIndex = -1

	if len(s) == 0 {
		return nil
	}

	wordStartIndex := nilIndex
	res := make([]string, 0, 4) //nolint:mnd,revive

	for j, sym := range s {
		switch {
		case unicode.IsSpace(sym):
			if wordStartIndex == nilIndex {
				continue
			}

			res = append(res, s[wordStartIndex:j])

			wordStartIndex = nilIndex
		default:
			if wordStartIndex == nilIndex {
				wordStartIndex = j
			}
		}
	}

	if wordStartIndex != nilIndex {
		res = append(res, s[wordStartIndex:])
	}

	return res
}

func FormatUint[T ~uint | ~uint64 | ~uint32 | ~uint16 | ~uint8](val T) string {
	return strconv.FormatUint(uint64(val), 10)
}

func FormatUintPtr[
	T ~uint | ~uint64 | ~uint32 | ~uint16 | ~uint8,
	TPointer interface{ *T }](val TPointer) string {
	if val == nil {
		return placeholder.Nil
	}

	return strconv.FormatUint(uint64(*val), 10)
}

func FormatInt[T ~int | ~int64 | ~int32 | ~int16 | ~int8](val T) string {
	return strconv.FormatInt(int64(val), 10)
}

func FormatIntPtr[
	T ~int | ~int64 | ~int32 | ~int16 | ~int8,
	TPointer interface{ *T }](val TPointer) string {
	if val == nil {
		return placeholder.Nil
	}

	return strconv.FormatInt(int64(*val), 10)
}

func String[T ~string, TPointer interface{ *T }](val TPointer) string {
	if val == nil {
		return placeholder.Nil
	}

	return string(*val)
}

func Stringer(val interface{ String() string }) string {
	if ptrof.IsNil(val) {
		return placeholder.Nil
	}

	return val.String()
}

func FormatIntSlice[S ~[]E, E ~int | ~int64 | ~int32 | ~int16 | ~int8](val S, sep string) string {
	buf := formatIntSlice(val, sep)
	if len(buf) == 0 {
		return ""
	}

	return unsafe.String(unsafe.SliceData(buf), len(buf))
}

func formatIntSlice[S ~[]E, E ~int | ~int64 | ~int32 | ~int16 | ~int8](val S, sep string) []byte {
	if len(val) == 0 {
		return nil
	}

	buf := []byte{}

	for i := range len(val) {
		if i != 0 {
			buf = append(buf, sep...)
		}

		buf = append(buf, strconv.FormatInt(int64(val[i]), 10)...)
	}

	return buf
}

func Sha256IntSlice[S ~[]E, E ~int | ~int64 | ~int32 | ~int16 | ~int8](val S) []byte {
	buf := formatIntSlice(val, ",")
	if len(buf) == 0 {
		buf = []byte(placeholder.Empty)
	}

	h := sha256.Sum256(buf)

	return h[:]
}

func FormatUintSlice[S ~[]E, E ~uint | ~uint64 | ~uint32 | ~uint16 | ~uint8](val S, sep string) string {
	buf := formatUintSlice(val, sep)
	if len(buf) == 0 {
		return ""
	}

	return unsafe.String(unsafe.SliceData(buf), len(buf))
}

func formatUintSlice[S ~[]E, E ~uint | ~uint64 | ~uint32 | ~uint16 | ~uint8](val S, sep string) []byte {
	if len(val) == 0 {
		return nil
	}

	buf := []byte{}

	for i := range len(val) {
		if i != 0 {
			buf = append(buf, sep...)
		}

		buf = append(buf, strconv.FormatUint(uint64(val[i]), 10)...)
	}

	return buf
}

func Sha256UintSlice[S ~[]E, E ~uint | ~uint64 | ~uint32 | ~uint16 | ~uint8](val S) []byte {
	buf := formatUintSlice(val, ",")
	if len(buf) == 0 {
		buf = []byte(placeholder.Empty)
	}

	h := sha256.Sum256(buf)

	return h[:]
}

func Sha256Hex(val string) string {
	h := sha256.Sum256([]byte(val))

	return hex.EncodeToString(h[:])
}

func ConcatBytes(s ...[]byte) string {
	if len(s) == 0 {
		return ""
	}

	sb := &strings.Builder{}

	for i := range len(s) {
		sb.Write(s[i])
	}

	return sb.String()
}

func TypeStringFor[T any]() string { return typestring.TypeString(reflect.TypeFor[T]()) }

func TypeStringForExpr[T any](expr T) string { return typestring.TypeStringForExpr(expr) }

func TypeStringOfVal(val any) string { return typestring.TypeStringOfVal(val) }

func TypeString(typ reflect.Type) string { return typestring.TypeString(typ) }

func EqualFold[T ~string](a, b T) bool { return strings.EqualFold(string(a), string(b)) }
