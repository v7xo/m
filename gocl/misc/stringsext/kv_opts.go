package stringsext

import (
	"encoding/json"
	"fmt"
	"iter"
	"maps"
	"strings"

	"gopkg.in/yaml.v3"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type KVOpts struct {
	dict    KVDictString
	encoded string
}

func (v KVOpts) Flags(prefix string) []string {
	if v.dict == nil {
		return nil
	}

	res := make([]string, 0, len(v.dict)*2)

	for k, v := range v.dict {
		if k = strings.TrimSpace(k); k == "" {
			continue
		}

		if !strings.HasPrefix(k, prefix) {
			k = prefix + k
		}

		res = append(res, k)

		for i := range len(v) {
			if i != 0 {
				res = append(res, k)
			}

			if v[i] == "" {
				continue
			}

			res = append(res, v[i])
		}
	}

	return res
}

func (v KVOpts) IsZero() bool { return len(v.dict) == 0 }

func (v KVOpts) CloneKVDict() KVDictString { return maps.Clone(v.dict) }

func (v KVOpts) All() iter.Seq2[string, ListString] { return maps.All(v.dict) }

func (v KVOpts) HasAll(o KVOpts) bool {
	if len(o.dict) == 0 {
		return true
	}

	for oKey, oVal := range o.dict {
		if vVal, ok := v.dict[oKey]; !ok || !oVal.Equal(vVal) {
			return false
		}
	}

	return true
}

func (v KVOpts) HasAny(o KVOpts) bool {
	if len(o.dict) == 0 {
		return false
	}

	for oKey, oVal := range o.dict {
		if vVal, ok := v.dict[oKey]; ok && oVal.Equal(vVal) {
			return true
		}
	}

	return false
}

func (v KVOpts) Get(k string) (ListString, bool) { r, ok := v.dict[k]; return r, ok }

func (v KVOpts) Count() int { return len(v.dict) }

func (v KVOpts) MarshalYAML() (any, error) { return v.encoded, nil }

func (v KVOpts) MarshalText() ([]byte, error) { return []byte(v.encoded), nil }

func (v *KVOpts) UnmarshalText(text []byte) error { return v.FromString(string(text)) }

func (v *KVOpts) UnmarshalYAML(n *yaml.Node) error {
	var encoded string

	if err := n.Decode(&encoded); err != nil {
		return err
	}

	return v.FromString(encoded)
}

func (v KVOpts) MarshalJSON() ([]byte, error) { return json.Marshal(v.encoded) }

func (v *KVOpts) UnmarshalJSON(b []byte) error {
	var encoded string

	if err := json.Unmarshal(b, &encoded); err != nil {
		return err
	}

	return v.FromString(encoded)
}

func (v KVOpts) encode() (string, error) { r, err := v.dict.MarshalText(); return string(r), err }

func (v KVOpts) String() string { return placeholder.EmptyIf(v.encoded) }

func (v KVOpts) Formatted() string { return v.encoded }

func (v *KVOpts) MustFromKVDict(val KVDictString) {
	err := v.FromKVDict(val)
	if err != nil {
		panic(err)
	}
}

func (v *KVOpts) FromKVDict(val KVDictString) (err error) {
	if v == nil {
		return nil
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to make '%s' from '%s'; %w", TypeStringForExpr(v), TypeStringForExpr(val), err)
		}
	}()

	cp := KVOpts{dict: val}

	cp.encoded, err = cp.encode()
	if err != nil {
		return err
	}

	*v = cp

	return nil
}

func (v *KVOpts) FromString(val string) (err error) {
	if v == nil {
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to parse '%s' from '%s'; %w", TypeStringForExpr(v), TypeStringForExpr(val), err)
		}
	}()

	cp := KVOpts{encoded: val}

	if cp.dict, _, err = ParseKVDictString(cp.encoded); err != nil {
		return err
	}

	*v = cp

	return nil
}

func (v *KVOpts) MustFromString(val string) {
	if err := v.FromString(val); err != nil {
		panic(err)
	}
}
