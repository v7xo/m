package metrics

import (
	"context"
	"time"

	"rn.m.vo7.io/vo7io/m/gocl/app/metricstub"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
	querypkg "rn.m.vo7.io/vo7io/m/gocl/storage/query"
	apipkg "rn.m.vo7.io/vo7io/m/gocl/storageapi"
)

type impl struct {
	inner apipkg.SQLAPI
}

func (s *impl) StorageType() storage.StorageType {
	return s.inner.StorageType()
}

func (s *impl) StorageName() string {
	return s.inner.StorageName()
}

func (s *impl) Close(ctx context.Context) (err error) {
	stats := querypkg.NewQueryExecStats(querypkg.CloseName)

	startTs := time.Now().UTC()
	err = s.inner.Close(ctx)

	stats.SetExec(time.Since(startTs))

	s.produceQueryExecMetrics(stats, err)

	return
}

func (s *impl) CloseWithLog(ctx context.Context, log logext.Logger) {
	var err error

	if log == nil {
		log = s.Log()
	}

	defer func() {
		if err != nil {
			log.Err(ctx, err).Msg("error while closing storage API instance")
		}
	}()

	stats := querypkg.NewQueryExecStats(querypkg.CloseName)

	startTs := time.Now().UTC()
	err = s.inner.Close(ctx)

	stats.SetExec(time.Since(startTs))

	s.produceQueryExecMetrics(stats, err)
}

func (s *impl) IsOperable() bool {
	return s.inner.IsOperable()
}

func (s *impl) IsTxSupported() bool {
	return s.inner.IsTxSupported()
}

func (s *impl) HasTx() (has bool, iso storage.IsolationLevel) {
	return s.inner.HasTx()
}

func (s *impl) BeginTxIsoAccess(ctx context.Context, iso storage.IsolationLevel, access storage.AccessMode) (err error) {
	stats := querypkg.NewQueryExecStats(querypkg.TxBeginName)
	ctx = stats.AssignToContext(ctx)

	err = s.inner.BeginTxIsoAccess(ctx, iso, access)

	s.produceQueryExecMetrics(stats, err)

	return
}

func (s *impl) BeginTxIso(ctx context.Context, iso storage.IsolationLevel) (err error) {
	stats := querypkg.NewQueryExecStats(querypkg.TxBeginName)
	ctx = stats.AssignToContext(ctx)

	err = s.inner.BeginTxIso(ctx, iso)

	s.produceQueryExecMetrics(stats, err)

	return
}

func (s *impl) BeginTx(ctx context.Context) (err error) {
	stats := querypkg.NewQueryExecStats(querypkg.TxBeginName)
	ctx = stats.AssignToContext(ctx)

	err = s.inner.BeginTx(ctx)

	s.produceQueryExecMetrics(stats, err)

	return
}

func (s *impl) CommitTx(ctx context.Context) (err error) {
	stats := querypkg.NewQueryExecStats(querypkg.TxCommitName)
	ctx = stats.AssignToContext(ctx)

	err = s.inner.CommitTx(ctx)

	s.produceQueryExecMetrics(stats, err)

	return
}

func (s *impl) RollbackTx(ctx context.Context) (err error) {
	stats := querypkg.NewQueryExecStats(querypkg.TxRollbackName)
	ctx = stats.AssignToContext(ctx)

	err = s.inner.RollbackTx(ctx)

	s.produceQueryExecMetrics(stats, err)

	return
}

func (s *impl) ExecSQL(ctx context.Context, query querypkg.Named, dst any, args ...any) (err error) {
	stats := querypkg.NewQueryExecStats(query.Name())
	ctx = stats.AssignToContext(ctx)

	err = s.inner.ExecSQL(ctx, query, dst, args...)

	s.produceQueryExecMetrics(stats, err)

	return
}

func (s *impl) ExecSQLNonQuery(ctx context.Context, query querypkg.Named, args ...any) (err error) {
	stats := querypkg.NewQueryExecStats(query.Name())
	ctx = stats.AssignToContext(ctx)

	err = s.inner.ExecSQLNonQuery(ctx, query, args...)

	s.produceQueryExecMetrics(stats, err)

	return
}

func (s *impl) ExecSQLOneRow(ctx context.Context, query querypkg.Named, dst any, args ...any) (err error) {
	stats := querypkg.NewQueryExecStats(query.Name())
	ctx = stats.AssignToContext(ctx)

	err = s.inner.ExecSQLOneRow(ctx, query, dst, args...)

	s.produceQueryExecMetrics(stats, err)

	return
}

func (s *impl) Log() logext.Logger {
	return s.inner.Log()
}

func (s *impl) Ping(ctx context.Context) (err error) {
	stats := querypkg.NewQueryExecStats(querypkg.PingName)
	ctx = stats.AssignToContext(ctx)

	err = s.inner.Ping(ctx)

	s.produceQueryExecMetrics(stats, err)

	return
}

func (s *impl) Name() string {
	return s.inner.Name()
}

func (s *impl) produceQueryExecMetrics(stats querypkg.ExecStats, err error) {
	dbType, apiName, opStatus := s.StorageType().String(), s.Name(), errs.TranslateStatus(err).String()

	queryName, _ := stats.Name()
	queryName = placeholder.EmptyIf(queryName)

	totalExecMs := int64(0)
	totalExecPub := false

	if dur, ok := stats.AcquireConn(); ok {
		totalExecMs, totalExecPub = dur.Milliseconds(), true

		metricstub.
			StorageAPIQueryExecStatsAcquireConnSeconds.
			WithLabelVals(dbType, apiName, opStatus).
			Observe(dur.Seconds())
	}

	if dur, ok := stats.Exec(); ok {
		totalExecMs += dur.Milliseconds()
		totalExecPub = true

		metricstub.
			StorageAPIQueryExecStatsExecSeconds.
			WithLabelVals(dbType, apiName, queryName, opStatus).
			Observe(dur.Seconds())
	}

	if dur, ok := stats.Scan(); ok {
		totalExecMs += dur.Milliseconds()
		totalExecPub = true

		metricstub.
			StorageAPIQueryExecStatsScanSeconds.
			WithLabelVals(dbType, apiName, queryName, opStatus).
			Observe(float64(dur.Seconds()))
	}

	if totalExecPub {
		metricstub.
			StorageAPIQueryExecStatsTotalSeconds.
			WithLabelVals(dbType, apiName, queryName, opStatus).
			Observe((time.Duration(totalExecMs) * time.Millisecond).Seconds())
	}

	if val, ok := stats.RowsAffected(); ok {
		metricstub.
			StorageAPIQueryExecStatsRowsAffectedGaugeVec.
			WithLabelVals(dbType, apiName, queryName, opStatus).
			Set(float64(val))
	}
}

func (s *impl) ExecTxFn(ctx context.Context, fn func() error) error {
	return apipkg.ExecTxFn(ctx, s, fn)
}

func (s *impl) ExecTxIsoFn(ctx context.Context, iso storage.IsolationLevel, fn func() error) error {
	return apipkg.ExecTxIsoFn(ctx, s, iso, fn)
}

func (s *impl) ExecTxIsoAccessFn(ctx context.Context, iso storage.IsolationLevel, access storage.AccessMode, fn func() error) error {
	return apipkg.ExecTxIsoAccessFn(ctx, s, iso, access, fn)
}

func NewSQL(inner apipkg.SQLAPI) (apipkg.SQLAPI, error) {
	return &impl{inner: inner}, nil
}

func MustNewSQL(inner apipkg.SQLAPI) apipkg.SQLAPI {
	res, err := NewSQL(inner)
	if err != nil {
		panic(err)
	}

	return res
}
