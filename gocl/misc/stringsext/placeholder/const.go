package placeholder

import "rn.m.vo7.io/vo7io/m/gocl/status/opstatus"

const (
	ErrDetailedMessageOmitted = "<error detailed message omitted (see logs)>"

	Auto           = "<auto>"
	Default        = "<default>"
	Deprecated     = "<deprecated>"
	Empty          = "<empty>"
	Err            = "<error>"
	Failed         = "<failed>"
	Infinity       = "<infinity>"
	Missing        = "<missing>"
	Nil            = "<nil>"
	NoName         = "<no name>"
	NotApplicable  = "<not applicable>"
	NotAvailable   = "<n/a>"
	NotInitialized = "<not initialized>"
	Omitted        = "<omitted>"
	PII            = "<PII>"
	PIIRedacted    = "<PII, redacted>"
	Redacted       = "<redacted>"
	TooLarge       = "<too large>"
	Unassigned     = "<unassigned>"
	Uninitialized  = NotInitialized
	Unknown        = "<unknown>"
	Unspecified    = "<unspecified>"

	Overlimit        = "…"
	OverlimitLen int = 1 // correlates with [Overlimit]

	CallerCancelOpStatus  = opstatus.CallerCancel
	CallerTimeoutOpStatus = opstatus.CallerTimeout
	CancelOpStatus        = opstatus.Cancel
	FailOpStatus          = opstatus.Fail
	SuccessOpStatus       = opstatus.OK
	TimeoutOpStatus       = opstatus.Timeout
	UnknownOpStatus       = opstatus.UnknownStatus
)
