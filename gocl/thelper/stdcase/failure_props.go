package stdcase

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/status"

	"rn.m.vo7.io/vo7io/m/gocl/status/clerr"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	clgrpcstatus "rn.m.vo7.io/vo7io/m/gocl/status/grpc"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	"rn.m.vo7.io/vo7io/m/gocl/status/opstatus"
)

// FailureProps represensts test case failure conditions to be checked.
type FailureProps struct {
	// ShouldFailAsErr indicates "As" condition for expected error.
	// Senseless when ShouldFail is false.
	ShouldFailAsErr any
	// ShouldFailIsErr indicates "Is" condition for expected error.
	// Senseless when ShouldFail is false.
	ShouldFailIsErr error
	// ShouldGRPCStatus denotes "gRPC status" condition (see also [clgrpcstatus.Translate]).
	// When specified as nil, then condition not checked.
	// Senseless when ShouldFail is false.
	ShouldGRPCStatus *status.Status
	// ShouldOpStatusName indicates operation status name condition for expected error (see also [clerr.TranslateOpStatusName]).
	// When specified as [opst.Unspecified] then condition not checked.
	// Senseless when ShouldFail is false.
	ShouldOpStatusName opstatus.Name
	// ShouldLevel denotest "error level" condition (see also [errs.TranslateLevel]).
	// When specified as [level.Unspecified] then condition not checked.
	// Senseless when ShouldFail is false.
	ShouldLevel level.Name
	// ShouldFail indicates whether error expected or not.
	ShouldFail bool
}

func (f FailureProps) RequireConformError(t *testing.T, err error) {
	t.Helper()

	f.RequireConformErrorMsgf(t, err, "")
}

//nolint:perfsprint
func (f FailureProps) RequireConformErrorMsgf(t *testing.T, err error, msg string, args ...any) {
	t.Helper()

	userMsg := ""
	if len(msg) != 0 {
		userMsg = fmt.Sprintf("; %s", fmt.Sprintf(msg, args...))
	}

	if f.ShouldFail {
		require.Error(t, err, fmt.Sprintf("no test case error whereas it expected%s", userMsg))

		if !f.ShouldOpStatusName.IsUnspecified() {
			require.Equal(t, f.ShouldOpStatusName, clerr.TranslateStatus(err), fmt.Sprintf("actual test case error violates operation status condition%s", userMsg))
		}

		if !f.ShouldLevel.IsUnspecified() {
			require.Equal(t, f.ShouldLevel, errs.TranslateLevel(err), fmt.Sprintf("actual test case error violates \"error level\" condition%s", userMsg))
		}

		if f.ShouldFailAsErr != nil {
			require.ErrorAs(t, err, f.ShouldFailAsErr, fmt.Sprintf("actual test case error violates \"As\" condition%s", userMsg))
		}

		if f.ShouldFailIsErr != nil {
			require.ErrorIs(t, err, f.ShouldFailIsErr, fmt.Sprintf("actual test case error violates \"Is\" condition%s", userMsg))
		}

		if f.ShouldGRPCStatus != nil {
			st, _ := clgrpcstatus.Translate(err)
			require.Equal(t, f.ShouldGRPCStatus, st, fmt.Sprintf("actual test case error violates \"gRPC status\" condition%s", userMsg))
		}

		return
	}

	require.NoError(t, err, fmt.Sprintf("test case error whereas it not expected%s", userMsg))
}
