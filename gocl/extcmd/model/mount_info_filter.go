package model

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"rn.m.vo7.io/vo7io/m/gocl/value/enum"
)

//nolint:tagliatelle
type MountInfoFilter struct {
	FileSystemUUID  *[]uuid.UUID                   `yaml:"file_system_uuid,omitempty"`
	FileSystemLabel *[]string                      `yaml:"file_system_label,omitempty"`
	FileSystemType  *[]enum.Of[FileSystemTypeCode] `yaml:"file_system_type_code,omitempty"`
	Target          *[]string                      `yaml:"target,omitempty"`
	OptionsAll      *MountOptions                  `yaml:"options_all,omitempty"`
	OptionsAny      *MountOptions                  `yaml:"options_any,omitempty"`
}

func (v MountInfoFilter) Empty() bool {
	return (v.FileSystemUUID == nil || *v.FileSystemUUID == nil) &&
		(v.FileSystemLabel == nil || *v.FileSystemLabel == nil) &&
		(v.FileSystemType == nil || *v.FileSystemType == nil) &&
		(v.Target == nil || *v.Target == nil) &&
		v.OptionsAll == nil &&
		v.OptionsAny == nil
}

func (v MountInfoFilter) String() string {
	sb := strings.Builder{}

	w := func(n, v string) {
		if sb.Len() > 1 {
			sb.WriteString(", ")
		}

		sb.WriteString(n)
		sb.WriteString(": ")
		sb.WriteString(v)
	}

	sb.WriteByte('(')

	if v.FileSystemUUID != nil && *v.FileSystemUUID != nil {
		if len(*v.FileSystemUUID) == 1 {
			w("file system UUID", (*v.FileSystemUUID)[0].String())
		} else {
			w("file system UUID", "<"+strconv.FormatInt(int64(len(*v.FileSystemUUID)), 10)+" items>")
		}
	}

	if v.FileSystemLabel != nil && *v.FileSystemLabel != nil {
		if len(*v.FileSystemLabel) == 1 {
			w("file system label", (*v.FileSystemLabel)[0])
		} else {
			w("file system label", "<"+strconv.FormatInt(int64(len(*v.FileSystemLabel)), 10)+" items>")
		}
	}

	if v.FileSystemType != nil && *v.FileSystemType != nil {
		if len(*v.FileSystemType) == 1 {
			w("file system type", (*v.FileSystemType)[0].String())
		} else {
			w("file system type", "<"+strconv.FormatInt(int64(len(*v.FileSystemType)), 10)+" items>")
		}
	}

	if v.Target != nil && *v.Target != nil {
		if len(*v.Target) == 1 {
			w("target", (*v.Target)[0])
		} else {
			w("target", "<"+strconv.FormatInt(int64(len(*v.Target)), 10)+" items>")
		}
	}

	if v.OptionsAll != nil {
		w("all of options", v.OptionsAll.String())
	}

	if v.OptionsAny != nil {
		w("any of options", v.OptionsAny.String())
	}

	sb.WriteByte(')')

	return sb.String()
}
