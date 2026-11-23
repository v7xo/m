package opstatus

import "strings"

type Name string

const (
	AlreadyExists              Name = "already_exists"
	AlreadyLocked              Name = "already_locked"
	AmbiguousMatch             Name = "ambiguous_match"
	BadArgs                    Name = "bad_args"
	CallerCancel               Name = "caller_cancel"
	CallerRateLimit            Name = "caller_rate_limit"
	CallerTimeout              Name = "caller_timeout"
	Cancel                     Name = "cancel"
	Conflict                   Name = "conflict"
	DataContractInvalidVersion Name = "data_contract_invalid_version"
	DataContractViolation      Name = "data_contract_violation"
	DBConstraintViolation      Name = "db_constraint_violation"
	DBFail                     Name = "internal_db_fail"
	Fail                       Name = "fail"
	IdempotencyViolation       Name = "idempotency_violation"
	InternalFail               Name = "internal_fail"
	NotFound                   Name = "not_found"
	NotImplemented             Name = "not_implemented"
	OK                         Name = "OK"
	Timeout                    Name = "timeout"
	TooManyRequests            Name = "too_many_requests"
	Unauthned                  Name = "unauthenticated"
	Unauthzed                  Name = "unauthorized"
	UnknownStatus              Name = "unknown"
	Unspecified                Name = ""
	// WARN: for new members modify translation maps (toGRPCStatusMap, etc).
)

func (v Name) IsUnspecified() bool { return v == Unspecified }
func (v Name) IsFail() bool        { return strings.EqualFold(string(Fail), string(v)) }
func (v Name) String() string      { return string(v) }
