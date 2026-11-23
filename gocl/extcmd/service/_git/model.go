package git

import (
	"fmt"
	"strconv"
)

type WorkingTreeStatusCode uint8

const (
	WorkingTreeStatusCodeUnspecified WorkingTreeStatusCode = iota
	WorkingTreeStatusCodeAdded
	WorkingTreeStatusCodeChanged
	WorkingTreeStatusCodeCopied
	WorkingTreeStatusCodeDeleted
	WorkingTreeStatusCodeNotChanged
	WorkingTreeStatusCodeOther
	WorkingTreeStatusCodeRenamed
	WorkingTreeStatusCodeUnmerged
	WorkingTreeStatusCodeUntracked
)

func (v WorkingTreeStatusCode) String() string {
	switch v {
	case WorkingTreeStatusCodeUnspecified:
		return "Unspecified"
	case WorkingTreeStatusCodeAdded:
		return "Added"
	case WorkingTreeStatusCodeChanged:
		return "Changed"
	case WorkingTreeStatusCodeCopied:
		return "Copied"
	case WorkingTreeStatusCodeDeleted:
		return "Deleted"
	case WorkingTreeStatusCodeNotChanged:
		return "NotChanged"
	case WorkingTreeStatusCodeOther:
		return "Other"
	case WorkingTreeStatusCodeRenamed:
		return "Renamed"
	case WorkingTreeStatusCodeUnmerged:
		return "Unmerged"
	case WorkingTreeStatusCodeUntracked:
		return "Untracked"
	default:
		return strconv.FormatInt(int64(v), 10)
	}
}

type WorkingTreeStatusEntry struct {
	PathRel string
	PathAbs string
	Code    WorkingTreeStatusCode
}

func (v WorkingTreeStatusEntry) String() string {
	return fmt.Sprintf("(path-relative: %s, code: %s)", v.PathRel, v.Code)
}
