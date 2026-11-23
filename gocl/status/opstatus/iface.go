package opstatus

import (
	st "rn.m.vo7.io/vo7io/m/gocl/status/grpc"
)

type Status = interface {
	String()
	st.AsGRPCStatus
}
