package md

import (
	"fmt"
	"strconv"

	"google.golang.org/grpc/metadata"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/misc/slicesext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type Key = attrname.Name

func ExtractLastString(md metadata.MD, key Key) (string, bool) {
	vals, exists := md[string(key)]

	if !exists {
		return "", false
	}

	str, ok := slicesext.Last(vals)
	if !ok {
		return "", false
	}

	return str, true
}

func ExtractLastInt64(md metadata.MD, key Key) (int64, bool, error) {
	vals, exists := md[string(key)]

	if !exists {
		return 0, false, nil
	}

	intStr, ok := slicesext.LastNonDefault(vals)
	if !ok {
		return 0, false, nil
	}

	parsedInt, err := strconv.ParseInt(intStr, 10, 64)
	if err != nil {
		return 0,
			false,
			errs.BadArgsWrapWith(fmt.Errorf("invalid value of metadata; metadata value can't be parsed as integer; metadata: %s; value: %s; err: %w", key, intStr, err))
	}

	return parsedInt, true, nil
}
