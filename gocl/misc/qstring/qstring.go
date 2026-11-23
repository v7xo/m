package qstring

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"gopkg.in/yaml.v3"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/value/enum"
)

type (
	Syntax int32

	String struct {
		parsed *atomic.Pointer[parsed]
		raw    string
	}

	parsed struct {
		err    error
		re     *regexp.Regexp
		kvOpts stringsext.KVOpts
		li     []string
		syntax Syntax
	}
)

const (
	SyntaxPlain = Syntax(iota)
	SyntaxRE
	SyntaxKV
	SyntaxLi

	rePrefix = "re:"
	kvPrefix = "kv:"
	liPrefix = "li:"
)

var (
	_ = enum.MustRegister(
		map[Syntax]string{
			SyntaxPlain: "plain",
			SyntaxRE:    "re",
			SyntaxKV:    "kv",
			SyntaxLi:    "li",
		})

	testCache   = map[string]String{}
	testCacheMu = &sync.RWMutex{}
)

func (v Syntax) String() string { return enum.String(v) }

func (v String) Syntax() Syntax {
	switch {
	case strings.HasPrefix(v.raw, rePrefix):
		return SyntaxRE
	case strings.HasPrefix(v.raw, kvPrefix):
		return SyntaxKV
	case strings.HasPrefix(v.raw, liPrefix):
		return SyntaxLi
	default:
		return SyntaxPlain
	}
}

func (v String) parse() (*parsed, error) {
	if v.parsed == nil {
		return nil, fmt.Errorf("%w; value type: %s", errs.ErrUninitializedValue, typestring.TypeStringForExpr(v))
	}

	ld := v.parsed.Load()
	if ld != nil {
		return ld, nil
	}

	syntax := v.Syntax()

	switch syntax {
	case SyntaxPlain:
		ld = &parsed{syntax: SyntaxPlain}
	case SyntaxRE:
		compiled, compileErr := regexp.Compile(v.raw[len(rePrefix):])
		if compileErr != nil {
			ld = &parsed{
				err: fmt.Errorf("%w; failed to parse string '%s' as regular expression; %w",
					errs.ErrInvalidSyntax,
					stringsext.TruncateDefault(v.raw[len(rePrefix):]),
					compileErr),
				syntax: syntax,
			}

			break
		}

		ld = &parsed{re: compiled, syntax: syntax}
	case SyntaxKV:
		kvOpts := stringsext.KVOpts{}
		if parseErr := kvOpts.FromString(v.raw[len(kvPrefix):]); parseErr != nil {
			ld = &parsed{err: parseErr, syntax: syntax}

			break
		}

		ld = &parsed{kvOpts: kvOpts, syntax: syntax}
	case SyntaxLi:
		dict, keyord, parseErr := stringsext.ParseKVDictString(v.raw[len(liPrefix):])
		if parseErr != nil {
			ld = &parsed{err: parseErr, syntax: syntax}

			break
		}

		var li []string

		for i := range len(keyord) {
			if val := dict[keyord[i]]; len(val) != 0 {
				ld = &parsed{
					err: fmt.Errorf("%w; failed to parse string '%s' as comma separated list items",
						errs.ErrInvalidSyntax,
						stringsext.TruncateDefault(v.raw[len(liPrefix):])),
					syntax: syntax,
				}

				break
			}

			if keyord[i] == "" {
				continue
			}

			if li == nil {
				li = make([]string, 0, len(keyord))
			}

			li = append(li, keyord[i])
		}

		if ld == nil {
			ld = &parsed{li: li, syntax: syntax}
		}
	default:
		ld = &parsed{err: fmt.Errorf("%w; unknown/unsupported syntax '%s' to parse '%s'", errs.ErrNotSupported, syntax, stringsext.TruncateDefault(v.raw)), syntax: syntax}
	}

	if v.parsed.CompareAndSwap(nil, ld) {
		return ld, nil
	}

	return v.parsed.Load(), nil
}

func (v String) MustRegexp() *regexp.Regexp {
	r, err := v.Regexp()
	if err != nil {
		panic(err)
	}

	return r
}

func (v String) Regexp() (*regexp.Regexp, error) {
	p, err := v.parse()
	if err != nil {
		return nil, err
	}

	if p.syntax != SyntaxRE {
		return nil, fmt.Errorf("source string '%s' wasn't specified as reqular expression; use prefix '%s' to denote reqular expression", v, rePrefix)
	} else if p.err != nil {
		return nil, p.err
	}

	return p.re, nil
}

func (v String) MustKVOpts() stringsext.KVOpts {
	r, err := v.KVOpts()
	if err != nil {
		panic(err)
	}

	return r
}

func (v String) MustList() []string {
	r, err := v.List()
	if err != nil {
		panic(err)
	}

	return r
}

func (v String) List() ([]string, error) {
	p, err := v.parse()
	if err != nil {
		return nil, err
	}

	if p.syntax != SyntaxLi {
		return nil, fmt.Errorf("source string '%s' wasn't specified as comma separated list; use prefix '%s' to denote such syntax", v, liPrefix)
	} else if p.err != nil {
		return nil, p.err
	}

	return slices.Clone(p.li), nil
}

func (v String) KVOpts() (stringsext.KVOpts, error) {
	p, err := v.parse()
	if err != nil {
		return stringsext.KVOpts{}, err
	}

	if p.syntax != SyntaxKV {
		return stringsext.KVOpts{}, fmt.Errorf("source string '%s' wasn't specified as key-value options; use prefix '%s' to denote reqular expression", v, kvPrefix)
	} else if p.err != nil {
		return stringsext.KVOpts{}, p.err
	}

	return p.kvOpts, nil
}

func (v String) MustPlain() string {
	r, err := v.Plain()
	if err != nil {
		panic(err)
	}

	return r
}

func (v String) Plain() (string, error) {
	p, err := v.parse()
	if err != nil {
		return "", err
	}

	if p.syntax != SyntaxPlain {
		return "", fmt.Errorf("source string '%s' wasn't specified as plain string; it was specified in syntax '%s'", v, p.syntax)
	} else if p.err != nil {
		return "", p.err
	}

	return v.raw, nil
}

func (v String) MarshalYAML() (any, error) { return v.raw, nil }

func (v *String) UnmarshalYAML(n *yaml.Node) error {
	var raw string

	if err := n.Decode(&raw); err != nil {
		return err
	}

	return v.FromString(raw)
}

func (v String) MarshalJSON() ([]byte, error) { return json.Marshal(v.raw) }

func (v String) IsZero() bool { return v.raw == "" }

func (v *String) UnmarshalJSON(b []byte) error {
	var raw string

	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	return v.FromString(raw)
}

func (v String) MarshalText() ([]byte, error) { return []byte(v.raw), nil }

func (v *String) UnmarshalText(text []byte) error { return v.FromString(string(text)) }

func (v *String) FromString(val string) (err error) {
	if v == nil {
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to parse '%s' from '%s'; %w", typestring.TypeStringForExpr(v), typestring.TypeStringForExpr(val), err)
		}
	}()

	if testing.Testing() {
		testCacheMu.RLock()

		if e, ok := testCache[val]; ok {
			testCacheMu.RUnlock()

			*v = e

			return nil
		}

		testCacheMu.RUnlock()
	}

	cp := String{parsed: &atomic.Pointer[parsed]{}, raw: val}

	// TODO: it might be useful to optionally delay parsing (?).
	p, err := cp.parse()
	if err != nil {
		return err
	} else if p.err != nil {
		return p.err
	}

	if testing.Testing() {
		testCacheMu.Lock()

		if e, ok := testCache[val]; ok {
			testCacheMu.Unlock()

			*v = e

			return nil
		}

		testCache[val] = cp

		testCacheMu.Unlock()
	}

	*v = cp

	return nil
}

func (v String) String() string {
	return placeholder.UnspecifiedIfEmpty(stringsext.TruncateDefault(v.raw))
}

func Parse(s string) (res String, err error) {
	err = res.FromString(s)

	return
}

func MustParse(s string) (res String) {
	if err := res.FromString(s); err != nil {
		panic(err)
	}

	return
}
