package query

import (
	"context"
	"fmt"
	"time"

	clcontext "rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type ExecStats struct {
	name string

	acquireConnDur time.Duration
	execDur        time.Duration
	scanDur        time.Duration

	rowsAffected int
	flags        int32
}

const (
	acquireConnDurSetQueryExecStatsFlag int32 = 1
	execDurSetQueryExecStatsFlag        int32 = 2
	rowsAffectedSetQueryExecStatsFlag   int32 = 4
	scanDurSetQueryExecStatsFlag        int32 = 8
)

func (s *ExecStats) Name() (string, bool) {
	if s == nil {
		return "", false
	}

	return s.name, true
}

func (s *ExecStats) AcquireConn() (time.Duration, bool) {
	const setFlag = acquireConnDurSetQueryExecStatsFlag

	if s == nil {
		return 0, false
	}

	return s.acquireConnDur, s.flags&setFlag == setFlag
}

func (s *ExecStats) RowsAffected() (int, bool) {
	const setFlag = rowsAffectedSetQueryExecStatsFlag

	if s == nil {
		return 0, false
	}

	return s.rowsAffected, s.flags&setFlag == setFlag
}

func (s *ExecStats) Scan() (time.Duration, bool) {
	const setFlag = scanDurSetQueryExecStatsFlag

	if s == nil {
		return 0, false
	}

	return s.scanDur, s.flags&setFlag == setFlag
}

func (s *ExecStats) Exec() (time.Duration, bool) {
	const setFlag = execDurSetQueryExecStatsFlag

	if s == nil {
		return 0, false
	}

	return s.execDur, s.flags&setFlag == setFlag
}

func (s *ExecStats) SetAcquireConn(d time.Duration) {
	if s == nil {
		return
	}

	s.acquireConnDur = d
	s.flags |= acquireConnDurSetQueryExecStatsFlag
}

func (s *ExecStats) SetRowsAffected(v int) {
	if s == nil {
		return
	}

	s.rowsAffected = v
	s.flags |= rowsAffectedSetQueryExecStatsFlag
}

func (s *ExecStats) SetScan(d time.Duration) {
	if s == nil {
		return
	}

	s.scanDur = d
	s.flags |= scanDurSetQueryExecStatsFlag
}

func (s *ExecStats) SetExec(d time.Duration) {
	if s == nil {
		return
	}

	s.execDur = d
	s.flags |= execDurSetQueryExecStatsFlag
}

func (s *ExecStats) AssignToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, clcontext.KeyStorageQueryExecStats, s)
}

func ExecStatsFromContext(ctx context.Context) (*ExecStats, error) {
	if ctx == nil {
		return nil, nil //nolint:nilnil
	}

	ctxVal := ctx.Value(clcontext.KeyStorageQueryExecStats)
	if ctxVal == nil {
		return nil, nil //nolint:nilnil
	}

	if ctxValAsStats, ok := ctxVal.(*ExecStats); ok {
		return ctxValAsStats, nil
	}

	return nil,
		fmt.Errorf(
			"such type of context's value is not expected; value key: %s; actual value type: %T; expected value type: %T; err: %w",
			clcontext.KeyStorageQueryExecStats,
			ctxVal,
			&ExecStats{},
			errs.ErrUnexpectedTypeOfValueInContext)
}

func NewQueryExecStats(queryName string) ExecStats {
	return ExecStats{
		name: queryName,
	}
}
