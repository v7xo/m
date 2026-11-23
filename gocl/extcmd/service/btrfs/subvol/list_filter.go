package subvol

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"rn.m.vo7.io/vo7io/m/gocl/misc/uuidext"
)

type ListFilter struct {
	IsSnapshot   *bool
	UUID         uuidext.List
	ReachPathDir []string
}

func (v ListFilter) fn() func(Summary) bool {
	uuidFn := func(Summary) bool { return true }

	if v.UUID != nil {
		if len(v.UUID) == 0 {
			uuidFn = func(Summary) bool { return false }
		} else {
			m := make(map[uuid.UUID]struct{}, len(v.UUID))

			for i := range len(v.UUID) {
				m[v.UUID[i]] = struct{}{}
			}

			uuidFn = func(x Summary) bool {
				_, ok := m[x.UUID]

				return ok
			}
		}
	}

	isSnapshotFn := func(Summary) bool { return true }

	if v.IsSnapshot != nil {
		isSnapshot := *v.IsSnapshot

		isSnapshotFn = func(x Summary) bool { return x.IsCreatedAsSnapshot() == isSnapshot }
	}

	reachPathDirFn := func(Summary) bool { return true }

	if v.ReachPathDir != nil {
		if len(v.ReachPathDir) == 0 {
			reachPathDirFn = func(Summary) bool { return false }
		} else {
			m := make(map[string]struct{}, len(v.ReachPathDir))

			for i := range len(v.ReachPathDir) {
				m[v.ReachPathDir[i]] = struct{}{}
			}

			reachPathDirFn = func(x Summary) bool {
				_, ok := m[filepath.Dir(x.ReachPath)]

				return ok
			}
		}
	}

	return func(x Summary) bool { return isSnapshotFn(x) && uuidFn(x) && reachPathDirFn(x) }
}

func (v ListFilter) String() string {
	sb := &strings.Builder{}

	w := func(n, v string) {
		if sb.Len() > 1 {
			sb.WriteString(", ")
		}

		sb.WriteString(n)
		sb.WriteString(": ")
		sb.WriteString(v)
	}

	sb.WriteByte('(')

	if v.UUID != nil {
		w("UUID", v.UUID.String())
	}

	if v.IsSnapshot != nil {
		w("is snapshot", strconv.FormatBool(*v.IsSnapshot))
	}

	sb.WriteByte(')')

	return sb.String()
}
