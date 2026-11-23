package osutil

import "rn.m.vo7.io/vo7io/m/gocl/value/enum"

type Remove int8

const (
	RemoveYes Remove = (1 << iota)
	RemoveAll

	RemoveNo = Remove(0)
)

var _ = enum.MustRegisterFlag(
	map[Remove]string{
		RemoveAll: "all",
		RemoveYes: "yes",
		RemoveNo:  "no",
	},
)

func (v Remove) Yes() bool      { return v&RemoveYes == RemoveYes }
func (v Remove) All() bool      { return v&RemoveAll == RemoveAll }
func (v Remove) String() string { return enum.String(v) }
