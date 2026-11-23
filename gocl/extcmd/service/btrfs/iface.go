package btrfs

import (
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/btrfs/filesystem"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/btrfs/subvol"
)

type Service interface {
	Subvol() subvol.Subvol
	FileSystem() filesystem.FileSystem
}
