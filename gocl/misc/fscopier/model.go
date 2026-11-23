package fscopier

import (
	"regexp"
	"strconv"
)

var shellFilePathRegexp = regexp.MustCompile(`^.+\.sh$`)

type PathFilterAction uint

const (
	PathFilterActionUnspecified PathFilterAction = iota
	PathFilterActionIgnore
	PathFilterActionContinue
	PathFilterActionDeleteExisting
)

func (v PathFilterAction) String() string {
	switch v {
	case PathFilterActionUnspecified:
		return ""
	case PathFilterActionIgnore:
		return "ignore"
	case PathFilterActionContinue:
		return "continue"
	case PathFilterActionDeleteExisting:
		return "delete_existing"
	default:
		return strconv.Itoa(int(v))
	}
}

type ExistingAction string

const (
	ExistingActionFail      = ExistingAction("")
	ExistingActionOverwrite = ExistingAction("overwrite")
	ExistingActionSkip      = ExistingAction("skip")
	ExistingActionFilter    = ExistingAction("filter")
)

type (
	ContentFilterFn  func(src, dst string, content []byte) ([]byte, error)
	ExistingFilterFn func(src, dst string) (ExistingAction, error)
	NameFilterFn     func(srcPath, srcName, dstDir, dstName string) (string, error)
	PathFilterFn     func(src, dst string, file bool) (PathFilterAction, error)
)

type copyItem struct {
	srcPath string
	dstPath string

	isDir         bool
	dstMustBeDel  bool
	dstMustBeSkip bool
}
