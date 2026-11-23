package stringsext

import (
	"bytes"
	"encoding"
	"fmt"
	"slices"
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type KVDict[K ~string, V ~string] map[K]List[V]

var (
	_ encoding.TextMarshaler   = KVDict[string, string]{}
	_ encoding.TextUnmarshaler = &KVDict[string, string]{}
)

func (v KVDict[K, V]) Del(k K)                 { delete(v, k) }
func (v KVDict[K, V]) Get(k K) (List[V], bool) { r, ok := v[k]; return r, ok }
func (v KVDict[K, V]) Set(k K, val List[V])    { v[k] = val }

func (v KVDict[K, V]) String() string {
	if v == nil {
		return placeholder.Nil
	} else if len(v) == 0 {
		return placeholder.Empty
	}

	sb := &strings.Builder{}

	sb.WriteByte('[')

	sb.WriteString("count: ")
	sb.WriteString(FormatInt(len(v)))

	c := 0

	for key, val := range v {
		sb.WriteString(", ")

		if c == 32 {
			sb.WriteRune(TruncateElipsisSym)

			break
		}

		sb.WriteByte('(')

		sb.WriteString("key: ")
		sb.WriteString(string(key))

		if len(val) != 0 {
			sb.WriteString(", values: ")
			sb.WriteString(val.String())
		}

		sb.WriteByte(')')

		c++
	}

	sb.WriteByte(']')

	return sb.String()
}

func (v KVDict[K, V]) MustMarshalText() []byte {
	r, err := v.MarshalText()
	if err != nil {
		panic(err)
	}

	return r
}

func (v KVDict[K, V]) MarshalText() (_ []byte, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to marshal '%s' (%s) as text; %w", v, TypeStringForExpr(v), err)
		}
	}()

	if len(v) == 0 {
		return []byte{}, nil
	}

	sb := bytes.NewBuffer(nil)
	buf := make([]byte, 0, 512) // nolint:mnd,revive

	keys := make([]K, 0, len(v))
	for key, list := range v {
		switch {
		case len(key) == 0:
			for j := range len(list) {
				if len(list[j]) != 0 {
					return nil, fmt.Errorf("unexpected miss of key (empty string) whereas there are values '%s' presented for this key", list)
				}
			}

			continue
		}

		keys = append(keys, key)
	}

	slices.Sort(keys)

	for _, key := range keys {
		list := v[key]

		if sb.Len() != 0 {
			sb.WriteByte(',')
		}

		keyEncoded := v.encode(string(key), buf)
		sb.Write(keyEncoded)

		for j := range len(list) {
			if j != 0 {
				sb.WriteByte(',')
				sb.Write(keyEncoded)
			}

			//if len(list[j]) == 0 {
			//	continue
			//}

			sb.WriteByte('=')
			sb.Write(v.encode(string(list[j]), buf))
		}
	}

	return sb.Bytes(), nil
}

func (KVDict[K, V]) encode(s string, buf []byte) []byte {
	bb := bytes.NewBuffer(buf[:0])
	q := false

	for _, sym := range s {
		switch sym {
		case ',', '=':
			bb.WriteRune(sym)

			q = true
		case '\\':
			bb.WriteRune('\\')
			bb.WriteRune('\\')
		case '"':
			bb.WriteRune('\\')
			bb.WriteRune('"')

			q = true
		default:
			bb.WriteRune(sym)
		}
	}

	res := bb.Bytes()

	if q {
		res = slices.Grow(res, 2)
		res = append(res, 0, 0)

		copy(res[1:], res)

		res[0], res[len(res)-1] = '"', '"'
	}

	return res
}

func (v *KVDict[K, V]) UnmarshalText(text []byte) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to unmarshal '%s' from '%s'; %w", TypeStringForExpr(v), TruncateDefault(string(text)), err)
		}
	}()

	parsed, _, err := ParseKVDict[K, V](string(text))
	if err != nil {
		return nil
	}

	*v = parsed

	return nil
}

// ParseKVDict parses a string into a key-value dictionary.
//
// The input string `s` is expected to be in a format where keys and values are separated by `=` and pairs are separated by `,`.
// Values can optionally be quoted with double quotes, allowing for special characters such as commas, equals signs, and quotes.
// Special characters can be escaped with a backslash `\` to be interpreted literally.
//
// The function returns:
//   - A [KVDictString] containing the parsed key-value pairs.
//   - A slice of strings (`keyord`) where each element corresponds to the starting key of each pair.
//   - An error if the parsing fails due to invalid syntax.
//
// Examples:
//
//	dict, keyord, err := ParseKVDict(`key1=value1,key2="value, with comma",key3=value3`)
//	if err != nil {
//	    log.Fatalf("Failed to parse KVDict: %v", err)
//	}
//	fmt.Println(dict)    // Output: map[key1:value1 key2:value, with comma key3:value3]
//	fmt.Println(keyord) // Output: [key1 key2 key3]
//
// Notes:
//   - The function supports escape sequences for special characters within keys and values.
//   - If a key or value is not terminated by a comma, it is considered the final pair.
//   - The function validates the syntax and returns an error if it encounters unexpected symbols or malformed input.
//
//nolint:gocognit,exhaustive,cyclop,maintidx
func ParseKVDict[K ~string, V ~string](s string) (dict KVDict[K, V], keyord List[K], err error) {
	if len(s) == 0 {
		return nil, nil, nil
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to parse '%s' as key/value pairs '%s'; %w", TruncateDefault(s), TypeStringForExpr(dict), err)
		}
	}()

	type scope byte

	const (
		scopeNone scope = iota
		scopePair
		scopeKey
		scopeVal
		scopeText
		scopeQuotedText
		scopeEscape
	)

	const (
		comma            = ","
		invalidSyntaxMsg = `%w; unexpected '%s' symbol (position %d)`
	)

	var (
		ctx            = []scope{scopePair}
		buf            = make([]rune, 0, 96) // nolint:mnd,revive
		symPos         int
		k              string
		needFinalComma bool
	)

	addPair := func(key K, val V, omitVal bool) {
		if dict == nil {
			dict = make(KVDict[K, V])
		}

		li := dict[key]

		if !omitVal {
			li = append(li, val)
		}

		dict[key] = li

		keyord = append(keyord, key)
	}

	for {
		var sym rune

		for _, sym = range s {
			symPos++

			switch sym {
			case '\\':
				switch ctx[0] {
				case scopeEscape:
					buf = append(buf, sym)
					ctx = ctx[1:]
				case scopeQuotedText:
					ctx = slices.Insert(ctx, 0, scopeEscape)
				default:
					return nil, nil, fmt.Errorf(invalidSyntaxMsg, errs.ErrInvalidSyntax, string(sym), symPos)
				}
			case '=':
				switch ctx[0] {
				case scopeEscape:
					buf = append(buf, sym)
					ctx = ctx[1:]
				case scopeQuotedText:
					buf = append(buf, sym)

					continue
				case scopeText:
					ctx = ctx[1:]
				case scopeKey:
				default:
					return nil, nil, fmt.Errorf(invalidSyntaxMsg, errs.ErrInvalidSyntax, string(sym), symPos)
				}

				switch ctx[0] {
				case scopeKey:
					ctx = ctx[1:]

					if len(buf) == 0 {
						return nil, nil, fmt.Errorf(invalidSyntaxMsg, errs.ErrInvalidSyntax, string(sym), symPos)
					}

					k = string(buf)
					buf = buf[:0]
					ctx = slices.Insert(ctx, 0, scopeVal)
				default:
					return nil, nil, fmt.Errorf(invalidSyntaxMsg, errs.ErrInvalidSyntax, string(sym), symPos)
				}
			case ',':
				switch ctx[0] {
				case scopeEscape:
					buf = append(buf, sym)
					ctx = ctx[1:]

					continue
				case scopeQuotedText:
					buf = append(buf, sym)

					continue
				case scopePair:
					keyord = append(keyord, K(k))

					continue
				case scopeText:
					ctx = ctx[1:]
				case scopeKey, scopeVal:
					break
				default:
					return nil, nil, fmt.Errorf(invalidSyntaxMsg, errs.ErrInvalidSyntax, string(sym), symPos)
				}

				switch ctx[0] {
				case scopeKey:
					ctx = ctx[1:]

					if len(buf) == 0 {
						return nil, nil, fmt.Errorf(invalidSyntaxMsg, errs.ErrInvalidSyntax, string(sym), symPos)
					}

					addPair(K(buf), "", true)

					buf = buf[:0]
				case scopeVal:
					ctx = ctx[1:]

					addPair(K(k), V(buf), false)

					buf = buf[:0]
				default:
					return nil, nil, fmt.Errorf(invalidSyntaxMsg, errs.ErrInvalidSyntax, string(sym), symPos)
				}
			case '"':
				switch ctx[0] {
				case scopeEscape:
					buf = append(buf, sym)
					ctx = ctx[1:]
				case scopeQuotedText:
					ctx = ctx[1:]
				case scopePair:
					ctx = slices.Insert(ctx, 0, scopeQuotedText, scopeKey)
				case scopeVal:
					ctx = slices.Insert(ctx, 0, scopeQuotedText)
				default:
					return nil, nil, fmt.Errorf(invalidSyntaxMsg, errs.ErrInvalidSyntax, string(sym), symPos)
				}
			default:
				switch ctx[0] {
				case scopeEscape:
					return nil, nil, fmt.Errorf("%w; it is unknown how to escape '\\%s' symbol (position %d)", errs.ErrInvalidSyntax, string(sym), symPos)
				case scopePair:
					ctx = slices.Insert(ctx, 0, scopeText, scopeKey)

					buf = append(buf, sym)
				case scopeKey, scopeVal:
					ctx = slices.Insert(ctx, 0, scopeText)

					buf = append(buf, sym)
				case scopeText, scopeQuotedText:
					buf = append(buf, sym)
				}
			}
		}

		if sym != ',' {
			s, needFinalComma = comma, true

			continue
		} else if !needFinalComma {
			keyord = append(keyord, "")
		}

		if len(ctx) != 1 || ctx[0] != scopePair {
			return nil, nil, fmt.Errorf(invalidSyntaxMsg, errs.ErrInvalidSyntax, string(sym), symPos)
		}

		break
	}

	return
}
