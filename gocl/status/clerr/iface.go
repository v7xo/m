package clerr

import (
	"google.golang.org/grpc/status"

	"rn.m.vo7.io/vo7io/m/gocl/status/errcode"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

type Error interface {
	error

	level.AsLevel
	errcode.AsCustomErrorCode
	opstatus.StatusName

	GRPCStatus() *status.Status
	WithLevel(l level.Name) Error
	WithStatus(n opstatus.Name) Error
	WithStatusLevel(n opstatus.Name, l level.Name) Error

	Is(e error) bool
	Unwrap() error
}
