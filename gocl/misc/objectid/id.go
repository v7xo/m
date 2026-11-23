package objectid

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

const (
	MaxLen       = 511
	RegexPattern = `^[A-Za-z][A-Za-z0-9_\.-]{0,510}$`
)

var regex = regexp.MustCompile(RegexPattern)

// ID represents string unique identifier for any object.
type ID struct{ v string }

func (v ID) IsZero() bool { return v.v == "" }

func (v ID) Equal(y ID) bool { return v.v == y.v }

func (v ID) CompareTo(y ID) int { return strings.Compare(v.v, y.v) }

func (v ID) AsString() string { return v.v }

func (v ID) String() string {
	if v.v == "" {
		return placeholder.Unspecified
	}

	return v.v
}

func (v ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.v)
}

func (v *ID) UnmarshalJSON(b []byte) error {
	var val string

	if err := json.Unmarshal(b, &val); err != nil {
		return err
	}

	return v.FromString(val)
}

func (v ID) MarshalYAML() (any, error) {
	return v.v, nil
}

func (v *ID) UnmarshalYAML(n *yaml.Node) error {
	var val string

	if err := n.Decode(&val); err != nil {
		return err
	}

	return v.FromString(val)
}

func (v *ID) FromString(val string) (err error) {
	if v == nil {
		return
	} else if val == "" {
		*v = ID{}

		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to parse object ID from string; %w", err)
		}
	}()

	if !regex.MatchString(val) {
		return fmt.Errorf("specified string '%s' doesn't match expected pattern '%s'", val, RegexPattern)
	}

	*v = ID{v: val}

	return
}

type List []ID

func Parse(val string) (ID, error) {
	if val == "" {
		return ID{}, nil
	}

	if !regex.MatchString(val) {
		return ID{},
			fmt.Errorf("%w; specified string '%s' doesn't match expected pattern '%s' to be a valid object ID",
				errs.ErrBadArgs,
				val,
				RegexPattern)
	}

	return ID{v: val}, nil
}

func MustParse(val string) ID {
	if objID, err := Parse(val); err != nil {
		panic(err)
	} else {
		return objID
	}
}
