package model

import (
	"errors"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"

	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
)

type (
	MountOptionKey      string
	kvopts              = stringsext.KVOpts
	MountOptions        struct{ kvopts }
	mountOptionsBuilder struct{ dict stringsext.KVDictString }
)

const (
	MountOptionKeyDiscard   = MountOptionKey("discard")
	MountOptionKeyNoATime   = MountOptionKey("noatime")
	MountOptionKeyNoDev     = MountOptionKey("nodev")
	MountOptionKeyNoSUID    = MountOptionKey("nosuid")
	MountOptionKeyReadOnly  = MountOptionKey("ro")
	MountOptionKeyReadWrite = MountOptionKey("rw")
	MountOptionKeySSD       = MountOptionKey("ssd")
	MountOptionKeySubvol    = MountOptionKey("subvol")
	MountOptionKeySubvolID  = MountOptionKey("subvolid")
	MountOptionKeyNoDiscard = MountOptionKey("nodiscard")
)

type MountOptionFn func(*mountOptionsBuilder) error

func MountOption() MountOptionFn { return fopt.Nop[mountOptionsBuilder] }

func (o MountOptionFn) NoDiscard() MountOptionFn { return o.Discard(false) }

func (o MountOptionFn) Discard(val bool) MountOptionFn {
	return fopt.Chain(
		o,
		func(t *mountOptionsBuilder) {
			if val {
				t.dict.Del(string(MountOptionKeyNoDiscard))
				t.dict.Set(string(MountOptionKeyDiscard), nil)

				return
			}

			t.dict.Del(string(MountOptionKeyDiscard))
			t.dict.Set(string(MountOptionKeyNoDiscard), nil)
		})
}

func (o MountOptionFn) NoDev() MountOptionFn { return o.Dev(false) }

func (o MountOptionFn) Dev(val bool) MountOptionFn {
	return fopt.Chain(
		o,
		func(t *mountOptionsBuilder) {
			if val {
				t.dict.Del(string(MountOptionKeyNoDev))

				return
			}

			t.dict.Set(string(MountOptionKeyNoDev), nil)
		})
}

func (o MountOptionFn) ReadOnly() MountOptionFn {
	return fopt.Chain(
		o,
		func(t *mountOptionsBuilder) {
			t.dict.Del(string(MountOptionKeyReadWrite))
			t.dict.Set(string(MountOptionKeyReadOnly), nil)
		})
}

func (o MountOptionFn) ReadWrite(val bool) MountOptionFn {
	return fopt.Chain(
		o,
		func(t *mountOptionsBuilder) {
			t.dict.Del(string(MountOptionKeyReadOnly))
			t.dict.Set(string(MountOptionKeyReadWrite), nil)
		})
}

func (o MountOptionFn) NoATime() MountOptionFn { return o.ATime(false) }

func (o MountOptionFn) ATime(val bool) MountOptionFn {
	return fopt.Chain(
		o,
		func(t *mountOptionsBuilder) {
			if val {
				t.dict.Del(string(MountOptionKeyNoATime))

				return
			}

			t.dict.Set(string(MountOptionKeyNoATime), nil)
		})
}

func (o MountOptionFn) NoSUID() MountOptionFn { return o.SUID(false) }

func (o MountOptionFn) SUID(val bool) MountOptionFn {
	return fopt.Chain(
		o,
		func(t *mountOptionsBuilder) {
			if val {
				t.dict.Del(string(MountOptionKeyNoSUID))

				return
			}

			t.dict.Set(string(MountOptionKeyNoSUID), nil)
		})
}

func (o MountOptionFn) BtrfsSubvol(val string) MountOptionFn {
	return fopt.ChainErr(
		o,
		func(t *mountOptionsBuilder) error {
			if val == "" {
				return errors.New("subvolume name can't be an empty string")
			}

			t.dict.Set(string(MountOptionKeySubvol), stringsext.ListString{val})

			return nil
		})
}

func (o MountOptionFn) BtrfsSubvolID(val int64) MountOptionFn {
	return fopt.ChainErr(
		o,
		func(t *mountOptionsBuilder) error {
			const (
				mainSubvolID    = 5
				userSubvolIDMin = 256
			)

			if val != mainSubvolID && val < userSubvolIDMin {
				return fmt.Errorf("BTRFS subvolume ID '%d' is out of range, must be either '%d' or any greater than '%d'", val, mainSubvolID, userSubvolIDMin-1)
			}

			t.dict.Set(string(MountOptionKeySubvolID), stringsext.ListString{strconv.FormatInt(val, 10)})

			return nil
		})
}

func (o MountOptionFn) BtrfsSubvolIDString(val string) MountOptionFn {
	return fopt.ChainErr(
		o,
		func(t *mountOptionsBuilder) error {
			if val == "" {
				return errors.New("subvolume ID can't be an empty string")
			}

			id, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return fmt.Errorf("failed to parse '%s' as BTRFS subvolume ID (int64); %w", val, err)
			}

			return MountOption().BtrfsSubvolID(id)(t)
		})
}

func (v MountOptions) With(opts ...MountOptionFn) (MountOptions, error) {
	nonNilOpts := make([]MountOptionFn, 0, len(opts))

	for i := range len(opts) {
		if opts[i] == nil {
			continue
		}

		nonNilOpts = append(nonNilOpts, opts[i])
	}

	if len(nonNilOpts) == 0 {
		return v, nil
	}

	builder := &mountOptionsBuilder{dict: v.CloneKVDict()}
	if builder.dict == nil {
		builder.dict = make(stringsext.KVDictString)
	}

	if err := fopt.Apply(builder, nonNilOpts...); err != nil {
		return MountOptions{}, fmt.Errorf("failed to construct a new mount options; %w", err)
	}

	if err := v.kvopts.FromKVDict(builder.dict); err != nil {
		return MountOptions{}, fmt.Errorf("failed to construct a new mount options; %w", err)
	}

	return v, nil
}

func (v MountOptions) MustWith(opts ...MountOptionFn) MountOptions {
	r, err := v.With(opts...)
	if err != nil {
		panic(err)
	}

	return r
}

func (v MountOptions) ReadOnly() (bool, bool) {
	if _, ok := v.Get(string(MountOptionKeyReadOnly)); ok {
		return true, true
	} else if _, ok = v.Get(string(MountOptionKeyReadWrite)); ok {
		return false, true
	}

	return false, false
}

func (v MountOptions) NoATime() bool {
	_, ok := v.Get(string(MountOptionKeyNoATime))

	return ok
}

func (v MountOptions) NoDev() bool {
	_, ok := v.Get(string(MountOptionKeyNoATime))

	return ok
}

func (v MountOptions) SubvolIDInt64() (int64, bool, error) {
	val, ok := v.SubvolID()
	if !ok {
		return 0, false, nil
	}

	valInt64, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, true, fmt.Errorf("failed to parse int64 from subvolume ID option '%s' of mount info object '%s'; %w", val, v, err)
	}

	return valInt64, true, nil
}

func (v *MountOptions) UnmarshalJSON(b []byte) error     { return v.kvopts.UnmarshalJSON(b) }
func (v *MountOptions) UnmarshalYAML(n *yaml.Node) error { return v.kvopts.UnmarshalYAML(n) }
func (v MountOptions) HasAll(o MountOptions) bool        { return v.kvopts.HasAll(o.kvopts) }
func (v MountOptions) HasAny(o MountOptions) bool        { return v.kvopts.HasAny(o.kvopts) }

func (v MountOptions) Subvol() (string, bool) {
	l, ok := v.Get(string(MountOptionKeySubvol))
	li, _ := l.LastOrZero()

	return li, ok
}

func (v MountOptions) SubvolID() (string, bool) {
	l, ok := v.Get(string(MountOptionKeySubvolID))
	li, _ := l.LastOrZero()

	return li, ok
}

func (v MountOptions) ToMountCommandString() string { return v.Formatted() }

func (v *MountOptions) MustFromString(val string) {
	if err := v.FromString(val); err != nil {
		panic(err)
	}
}

func (v *MountOptions) FromString(val string) (err error) {
	if v == nil {
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to map '%s' from '%s'; %w", typestring.TypeStringForExpr(v), typestring.TypeStringForExpr(val), err)
		}
	}()

	cp := MountOptions{}

	if err = cp.kvopts.FromString(val); err != nil {
		return err
	}

	*v = cp

	return nil
}

func EmptyMountOptions() MountOptions { return MountOptions{} }

func NewMountOptions(opts ...MountOptionFn) (_ MountOptions, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", typestring.TypeStringFor[MountOptions](), err)
		}
	}()

	res := MountOptions{}

	if res, err = res.With(opts...); err != nil {
		return MountOptions{}, err
	}

	return res, nil
}

func MustNewMountOptions(opts ...MountOptionFn) MountOptions {
	res, err := NewMountOptions(opts...)
	if err != nil {
		panic(err)
	}

	return res
}
