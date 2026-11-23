package status

import "google.golang.org/grpc/status"

type AsGRPCStatus = interface{ GRPCStatus() *status.Status }
