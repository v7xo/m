package prommetric

import (
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/appmeta/info"
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

func validateUserLabelName(ln []string) error {
	for _, labelName := range ln {
		switch labelName {
		case string(attrname.AppName):
			fallthrough
		case string(attrname.AppCmdName):
			return fmt.Errorf("specified label name is reserved for predefined use and can't be used as custom label; name: %s", labelName)
		}
	}

	return nil
}

func prependPredefinedLabelName(names []string, appName bool) []string {
	prepLen := 1
	if appName {
		prepLen++
	}

	names = append(names, make([]string, prepLen)...)
	copy(names[prepLen:], names[0:])

	idx := 0

	if appName {
		names[0] = string(attrname.AppName)
		idx++
	}

	names[idx] = string(attrname.AppCmdName)

	return names
}

func prependPredefinedLabelVal(vals []string, appName bool) []string {
	prepLen := 1
	if appName {
		prepLen++
	}

	vals = append(vals, make([]string, prepLen)...)
	copy(vals[prepLen:], vals[0:])

	idx := 0

	if appName {
		vals[0] = info.Name()
		idx++
	}

	vals[idx] = info.CmdName()

	return vals
}

func normalizeLabelVal(lv []string) []string {
	for i := range lv {
		if len(lv[i]) == 0 {
			lv[i] = placeholder.Empty
		}
	}

	return lv
}
