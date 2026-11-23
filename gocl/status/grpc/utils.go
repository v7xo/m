package status

import (
	"errors"

	"google.golang.org/grpc/status"

	clpbt "rn.m.vo7.io/vo7io/m/gocl/pbt/types"
)

func Translate(err error) (s *status.Status, exactly bool) {
	if err == nil {
		exactly = true

		return
	}

	var asGRPCStatus AsGRPCStatus

	if errors.As(err, &asGRPCStatus) {
		s = asGRPCStatus.GRPCStatus()
		exactly = true

		return
	}

	s = status.FromContextError(err)

	return
}

func CustomError(st *status.Status) *clpbt.Error {
	if st == nil {
		return nil
	}

	for _, det := range st.Details() {
		if custErr, ok := det.(*clpbt.Error); ok {
			return custErr
		}
	}

	return nil
}

func ChangeMessage(st *status.Status, msg string) *status.Status {
	if st == nil {
		return nil
	}

	res := status.New(st.Code(), msg)

	if custErr := CustomError(st); custErr != nil {
		detailed, detailedErr := res.WithDetails(custErr)
		if detailedErr == nil {
			res = detailed
		}
	}

	return res
}
