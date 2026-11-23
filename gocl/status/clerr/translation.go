package clerr

import (
	"context"
	"errors"
	"os"

	"rn.m.vo7.io/vo7io/m/gocl/misc/ptrof"
	status "rn.m.vo7.io/vo7io/m/gocl/status/grpc"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

func TranslateStatus(err error) opstatus.Name {
	if ptrof.IsNil(err) {
		return opstatus.OK
	}

	// if error defines explicit status name.
	var asStatusName opstatus.StatusName
	if errors.As(err, &asStatusName) {
		return asStatusName.StatusName()
	}

	if errors.Is(err, context.DeadlineExceeded) || os.IsTimeout(err) {
		return opstatus.Timeout
	}

	if errors.Is(err, context.Canceled) {
		return opstatus.Cancel
	}

	if asGRPCStatus := status.AsGRPCStatus(nil); errors.As(err, &asGRPCStatus) {
		return opstatus.FromGRPCStatusCode(asGRPCStatus.GRPCStatus().Code())
	}

	return opstatus.Fail
}
