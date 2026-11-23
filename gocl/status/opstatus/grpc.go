package opstatus

import (
	"fmt"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	unknownGRPCStatus = status.New(codes.Unknown, "something went wrong")

	toGRPCStatusMapMu sync.Mutex
	toGRPCStatusMap   = map[Name]*status.Status{
		AlreadyExists:              status.New(codes.AlreadyExists, AlreadyExists.String()),
		AlreadyLocked:              status.New(codes.FailedPrecondition, AlreadyLocked.String()),
		AmbiguousMatch:             status.New(codes.Aborted, AlreadyLocked.String()),
		BadArgs:                    status.New(codes.InvalidArgument, BadArgs.String()),
		CallerCancel:               status.New(codes.Canceled, CallerCancel.String()),
		CallerRateLimit:            status.New(codes.ResourceExhausted, CallerRateLimit.String()),
		CallerTimeout:              status.New(codes.DeadlineExceeded, CallerTimeout.String()),
		Cancel:                     status.New(codes.Aborted, Cancel.String()),
		Conflict:                   status.New(codes.FailedPrecondition, Conflict.String()),
		DataContractInvalidVersion: status.New(codes.InvalidArgument, DataContractInvalidVersion.String()),
		DataContractViolation:      status.New(codes.InvalidArgument, DataContractViolation.String()),
		DBConstraintViolation:      status.New(codes.InvalidArgument, DBConstraintViolation.String()),
		DBFail:                     status.New(codes.Internal, InternalFail.String()),
		Fail:                       unknownGRPCStatus,
		IdempotencyViolation:       status.New(codes.FailedPrecondition, IdempotencyViolation.String()),
		InternalFail:               status.New(codes.Internal, InternalFail.String()),
		NotFound:                   status.New(codes.NotFound, NotFound.String()),
		NotImplemented:             status.New(codes.Unimplemented, NotImplemented.String()),
		OK:                         status.New(codes.OK, OK.String()),
		Timeout:                    status.New(codes.DeadlineExceeded, Timeout.String()),
		TooManyRequests:            status.New(codes.ResourceExhausted, TooManyRequests.String()),
		Unauthned:                  status.New(codes.Unauthenticated, Unauthned.String()),
		Unauthzed:                  status.New(codes.PermissionDenied, Unauthzed.String()),
		UnknownStatus:              unknownGRPCStatus,
		Unspecified:                unknownGRPCStatus,
	}

	// index - codes.Code.
	fromGRPCStatusCodeMap = func() []Name {
		res := make([]Name, 17) // nolint:mnd

		res[codes.Aborted] = Cancel
		res[codes.AlreadyExists] = AlreadyExists
		res[codes.Canceled] = Cancel
		res[codes.DataLoss] = InternalFail
		res[codes.DeadlineExceeded] = Timeout
		res[codes.FailedPrecondition] = Fail
		res[codes.Internal] = InternalFail
		res[codes.InvalidArgument] = BadArgs
		res[codes.NotFound] = NotFound
		res[codes.OK] = OK
		res[codes.OutOfRange] = InternalFail
		res[codes.PermissionDenied] = Unauthzed
		res[codes.ResourceExhausted] = InternalFail
		res[codes.Unauthenticated] = Unauthned
		res[codes.Unavailable] = InternalFail
		res[codes.Unimplemented] = NotImplemented
		res[codes.Unknown] = Fail

		return res
	}()
)

func (v Name) GRPCStatus() (res *status.Status) {
	toGRPCStatusMapMu.Lock()

	if inMap, ok := toGRPCStatusMap[v]; ok {
		res = inMap
	} else {
		res = status.New(codes.Unknown, v.String())

		toGRPCStatusMap[v] = res
	}

	toGRPCStatusMapMu.Unlock()

	return
}

func MustRegisterGRPCStatusMap(n Name, status *status.Status, reset bool) {
	if res := RegisterGRPCStatusMap(n, status, reset); !res {
		panic(fmt.Errorf("specified status name have an gRPC status mapping already; status name: %s", n))
	}
}

func RegisterGRPCStatusMap(name Name, status *status.Status, reset bool) (res bool) {
	toGRPCStatusMapMu.Lock()

	if e, ok := toGRPCStatusMap[name]; ok {
		switch {
		case e == status:
			res = true
		case reset:
			toGRPCStatusMap[name] = status
			res = true
		default:
			res = false
		}
	} else {
		toGRPCStatusMap[name] = status
		res = true
	}

	toGRPCStatusMapMu.Unlock()

	return
}

func FromGRPCStatusCode(c codes.Code) (res Name) {
	defer func() {
		if panicObj := recover(); panicObj != nil {
			res = Fail
		}
	}()

	return fromGRPCStatusCodeMap[c]
}
