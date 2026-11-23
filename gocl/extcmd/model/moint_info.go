package model

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/google/uuid"

	"rn.m.vo7.io/vo7io/m/gocl/misc/ptrof"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type MountInfo struct {
	Source                  BlockDeviceID
	FileSystemLabel         *string
	FileSystemRoot          string
	Target                  string
	Options                 MountOptions
	FileSystemSize          DataSize
	FileSystemSizeAvailable DataSize
	FileSystemType          FileSystemTypeCode
	FileSystemUUID          uuid.UUID
}

func (v MountInfo) String() string {
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

	w("target", placeholder.UnspecifiedIfEmpty(v.Target))
	w("type", v.FileSystemType.String())

	if v.FileSystemLabel != nil {
		w("label", *v.FileSystemLabel)
	}

	if v.FileSystemUUID != uuid.Nil {
		w("uuid", v.FileSystemUUID.String())
	}

	sb.WriteByte(')')

	return sb.String()
}

func (v MountInfo) MatchFilter(f MountInfoFilter) bool {
	if f.FileSystemUUID != nil && *f.FileSystemUUID != nil {
		vals := *f.FileSystemUUID
		found := false

		for i := range len(vals) {
			if vals[i] == v.FileSystemUUID {
				found = true

				break
			}
		}

		if !found {
			return false
		}
	}

	if f.FileSystemLabel != nil && *f.FileSystemLabel != nil {
		vals := *f.FileSystemLabel
		found := false

		switch {
		case v.FileSystemLabel != nil:
			label := *v.FileSystemLabel

			for i := range len(vals) {
				if vals[i] == label {
					found = true

					break
				}
			}
		}

		if !found {
			return false
		}
	}

	if f.FileSystemType != nil && *f.FileSystemType != nil {
		vals := *f.FileSystemType
		found := false

		for i := range len(vals) {
			if vals[i].Val == v.FileSystemType {
				found = true

				break
			}
		}

		if !found {
			return false
		}
	}

	if f.Target != nil && *f.Target != nil {
		vals := *f.Target
		found := false

		for i := range len(vals) {
			if vals[i] == v.Target {
				found = true

				break
			}
		}

		if !found {
			return false
		}
	}

	if f.OptionsAll != nil && !v.Options.HasAll(*f.OptionsAll) {
		return false
	}

	if f.OptionsAny != nil && !v.Options.HasAny(*f.OptionsAny) {
		return false
	}

	return true
}

type MountInfoList []MountInfo

func (v *MountInfoList) FromFindmntOutputReaderJSON(r io.Reader) (err error) {
	if v == nil {
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to map mount info model from 'findmnt' command JSON output; err: %w", err)
		}
	}()

	//nolint:tagliatelle
	type (
		outputElementType struct {
			Label *string `json:"label"`

			FSRoot  string `json:"fsroot"`
			FSType  string `json:"fstype"`
			Options string `json:"options"`
			Source  string `json:"source"`
			Target  string `json:"target"`

			Children []outputElementType `json:"children"`

			UUID          uuid.UUID `json:"uuid"`
			Size          int64     `json:"size"`
			SizeAvailable int64     `json:"avail"`
		}
		outputType struct {
			FileSystems []outputElementType `json:"filesystems"`
		}
	)

	var (
		output  outputType
		decoder = json.NewDecoder(r)
	)

	if err = decoder.Decode(&output); err != nil {
		return
	}

	if output.FileSystems == nil {
		*v = nil

		return
	} else if len(output.FileSystems) == 0 {
		*v = MountInfoList{}

		return
	}

	cp := make(MountInfoList, 0, len(output.FileSystems))

	for i := 0; i < len(output.FileSystems); i++ {
		e := output.FileSystems[i]

		if len(e.Target) == 0 {
			return fmt.Errorf("target mount point not specified in output (element '%d')", i)
		}

		if len(e.Source) == 0 {
			return fmt.Errorf("source device path not specified in output (element '%d')", i)
		}

		if len(e.FSType) == 0 {
			return fmt.Errorf("file system type not specified in output (element '%d')", i)
		}

		if len(e.FSRoot) == 0 {
			return fmt.Errorf("file system root not specified in output (element '%d')", i)
		}

		mi := MountInfo{
			FileSystemRoot:          e.FSRoot,
			FileSystemLabel:         ptrof.Copy(e.Label),
			FileSystemSize:          DataSize{Unit: DataSizeUnitB, UnitAmount: e.Size},
			FileSystemSizeAvailable: DataSize{Unit: DataSizeUnitB, UnitAmount: e.SizeAvailable},
			Source:                  NewBlockDeviceID(e.Source, BlockDeviceTypeCodeUnspecified, nil),
			Target:                  e.Target,
			FileSystemUUID:          e.UUID,
		}

		if err = mi.FileSystemType.FromString(e.FSType); err != nil {
			return fmt.Errorf("failed to map file system type (element '%d'); err: %w", i, err)
		}

		if err = mi.Options.FromString(e.Options); err != nil {
			return fmt.Errorf("failed to map mount options (element '%d'); err: %w", i, err)
		}

		cp = append(cp, mi)

		output.FileSystems = append(output.FileSystems, e.Children...)
	}

	*v = cp

	return
}

func (v MountInfoList) First(sel func(MountInfo) bool) (MountInfo, bool) {
	for i := range len(v) {
		if sel(v[i]) {
			return v[i], true
		}
	}

	return MountInfo{}, false
}

func (v MountInfoList) All(sel func(MountInfo) bool) MountInfoList {
	if v == nil {
		return nil
	} else if len(v) == 0 {
		return MountInfoList{}
	}

	res := make(MountInfoList, 0, len(v))

	for i := range len(v) {
		if sel(v[i]) {
			res = append(res, v[i])
		}
	}

	return res
}

func (v MountInfoList) MatchFilter(f MountInfoFilter) MountInfoList {
	if v == nil {
		return nil
	} else if f.Empty() {
		return slices.Clone(v)
	}

	res := make(MountInfoList, 0, len(v))

	for i := range len(v) {
		if v[i].MatchFilter(f) {
			res = append(res, v[i])
		}
	}

	return res
}

func (v MountInfoList) FileSystemRootLookup() MountInfoFileSystemRootLookup {
	if v == nil {
		return nil
	}

	res := make(MountInfoFileSystemRootLookup, len(v))

	for _, mi := range v {
		res[mi.FileSystemRoot] = append(res[mi.FileSystemRoot], mi)
	}

	return res
}

func (v MountInfoList) TargetLookup() MountInfoTargetLookup {
	if v == nil {
		return nil
	}

	res := make(MountInfoTargetLookup, len(v))

	for _, mi := range v {
		res[mi.Target] = append(res[mi.Target], mi)
	}

	return res
}

func (v MountInfoList) FileSystemUUIDLookup() MountInfoFileSystemUUIDLookup {
	if v == nil {
		return nil
	}

	res := make(MountInfoFileSystemUUIDLookup, len(v))

	for _, mi := range v {
		res[mi.FileSystemUUID] = append(res[mi.FileSystemUUID], mi)
	}

	return res
}

type (
	MountInfoFileSystemRootLookup map[string]MountInfoList
	MountInfoFileSystemUUIDLookup map[uuid.UUID]MountInfoList
	MountInfoTargetLookup         map[string]MountInfoList
)
