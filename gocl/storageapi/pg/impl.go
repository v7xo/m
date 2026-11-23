package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/georgysavva/scany/dbscan"
	"github.com/georgysavva/scany/v2/pgxscan"
	pgt "github.com/jackc/pgtype"
	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	clslices "rn.m.vo7.io/vo7io/m/gocl/misc/slicesext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
	"rn.m.vo7.io/vo7io/m/gocl/storage/pg"
	querypkg "rn.m.vo7.io/vo7io/m/gocl/storage/query"
	apipkg "rn.m.vo7.io/vo7io/m/gocl/storageapi"
)

type (
	impl struct {
		fopt.LogOpt
		scan                    *dbscan.API
		conn                    *pgx.Conn
		connCloser              PgxConnCloserFn
		connConf                *pgx.ConnConfig
		connConnector           PgxConnectorFn
		name                    string
		connConfDBName          string
		tx                      []txState
		commandTimeout          timeout.Timeout
		mu                      sync.Mutex
		closeRequestMark        atomic.Int32
		queryNamedParamDisabled bool
	}

	txState struct {
		tx  pgx.Tx
		iso storage.IsolationLevel
	}
)

const closeRequestYes = int32(1)

func (api *impl) StorageType() storage.StorageType { return storage.PostgreSQLStorageType }

func (api *impl) ExecSQLOneRow(ctx context.Context, query querypkg.Named, dst any, args ...any) error {
	return api.execSQL(ctx, query, dst, true, args...)
}

func (api *impl) ExecSQL(ctx context.Context, query querypkg.Named, dst any, args ...any) error {
	return api.execSQL(ctx, query, dst, false, args...)
}

func (api *impl) ExecSQLNonQuery(ctx context.Context, query querypkg.Named, args ...any) (err error) {
	var (
		stats         *querypkg.ExecStats
		execQueryText string
		execQueryArgs []any
	)

	queryText, err := query.Text()
	if err != nil {
		return
	}

	execCtx, execCtxCancel := api.makeExecCtx(ctx, query)
	defer execCtxCancel()

	if stats, err = querypkg.ExecStatsFromContext(ctx); err != nil {
		return
	}

	if execQueryText, execQueryArgs, err = api.formExecSQLPositionalArgs(query, args...); err != nil {
		return
	}

	startTs := time.Now().UTC()
	conn, err := api.requireConn(execCtx)

	stats.SetAcquireConn(time.Since(startTs))

	if err != nil {
		err = fmt.Errorf("error while acquiring DB connection; query name: %s; %w", query.Name(), err)

		return
	}

	startTs = time.Now().UTC()
	tag, err := conn.Exec(execCtx, execQueryText, execQueryArgs...)

	stats.SetExec(time.Since(startTs))

	if err != nil {
		if query.Anon() {
			err = fmt.Errorf("error while executing SQL \"no rows\" query; query text: %s; %w", queryText, err)
		} else {
			err = fmt.Errorf("error while executing SQL \"no rows\" query; query name: %s; %w", query.Name(), err)
		}

		err = normalizePostgresError(err)

		return
	}

	stats.SetRowsAffected(int(tag.RowsAffected()))

	return
}

func (api *impl) Ping(ctx context.Context) (err error) {
	var stats *querypkg.ExecStats

	execCtx, execCtxCancel := api.makeExecCtx(ctx, querypkg.Named{})
	defer execCtxCancel()

	if stats, err = querypkg.ExecStatsFromContext(ctx); err != nil {
		return
	}

	startTs := time.Now().UTC()
	conn, err := api.requireConn(execCtx)

	stats.SetAcquireConn(time.Since(startTs))

	if err != nil {
		err = fmt.Errorf("error while acquiring DB connection; %w", err)

		return
	}

	startTs = time.Now().UTC()
	err = conn.Ping(execCtx)

	stats.SetExec(time.Since(startTs))

	if err != nil {
		err = fmt.Errorf("ping for DB failed; %w", err)
	}

	return
}

func (api *impl) Close(ctx context.Context) (err error) {
	api.closeRequestMark.Store(closeRequestYes)

	var logEvt logext.Event

	defer func() {
		if logEvt != nil {
			logEvt.Send()
		}
	}()

	api.mu.Lock()

	if api.conn != nil {
		if api.connCloser != nil {
			err = api.connCloser(ctx, api.conn)

			switch {
			case err != nil:
				err = fmt.Errorf("error while closing DB connection; %w", err)
				logEvt = api.Log().Err(ctx, err)
			default:
				logEvt = api.Log().Trace(ctx).WithSendMsg("connection with storage closed")
			}
		}

		if err == nil {
			api.conn = nil
			api.connCloser = nil
			api.tx = []txState{}
		}
	}

	api.mu.Unlock()

	return
}

func (api *impl) acquireLockIfOperable() (err error) {
	if err = api.acquireLockIfNoCloseRequest(); err != nil || api.conn == nil {
		return
	}

	if api.conn.IsClosed() {
		err = fmt.Errorf("%w; storage connection is closed", errs.ErrNotOperable)

		api.mu.Unlock()
	} else if pgConn := api.conn.PgConn(); pgConn != nil && pgConn.IsBusy() {
		err = fmt.Errorf("%w; storage connection is busy", errs.ErrNotOperable)

		api.mu.Unlock()
	}

	return
}

func (api *impl) IsOperable() bool {
	if err := api.acquireLockIfOperable(); err != nil {
		return false
	}

	api.mu.Unlock()

	return true
}

func (api *impl) CloseWithLog(ctx context.Context, log logext.Logger) {
	var err error

	if log == nil || log.IsNop() {
		log = api.Log()
	}

	if err = api.Close(ctx); err != nil {
		log.Err(ctx, err).Msg("error while closing PostgreSQL storage API instance")
	}
}

func (api *impl) Name() string {
	return api.name
}

func (api *impl) StorageName() string {
	return api.connConfDBName
}

func (api *impl) BeginTx(ctx context.Context) (err error) {
	return api.BeginTxIso(ctx, storage.IsolationLevelInheritedOrDefault)
}

func (api *impl) BeginTxIso(ctx context.Context, iso storage.IsolationLevel) (err error) {
	return api.BeginTxIsoAccess(ctx, iso, storage.AccessModeDefault)
}

func (api *impl) BeginTxIsoAccess(ctx context.Context, iso storage.IsolationLevel, access storage.AccessMode) (err error) {
	stats, err := querypkg.ExecStatsFromContext(ctx)
	if err != nil {
		return
	}

	startTs := time.Now().UTC()
	conn, err := api.requireConn(ctx)

	stats.SetAcquireConn(time.Since(startTs))

	if err != nil {
		return err
	}

	if err = api.acquireLockIfOperable(); err != nil {
		return
	}

	//nolint:exhaustive
	switch iso {
	case storage.IsolationLevelInheritedOrDefault:
		if len(api.tx) != 0 {
			iso = api.tx[len(api.tx)-1].iso
		} else {
			iso = DefaultIsolationLevel
		}
	}

	txOpts, err := makeTxOptions(iso, access)
	if err != nil {
		api.mu.Unlock()

		return
	}

	startTs = time.Now().UTC()

	defer func() {
		stats.SetExec(time.Since(startTs))
	}()

	var tx pgx.Tx

	if len(api.tx) == 0 {
		tx, err = conn.BeginTx(ctx, txOpts)
		if err != nil {
			err = fmt.Errorf("error while beginning DB tx; %w", err)
			api.mu.Unlock()

			return
		}

		api.tx = append(api.tx, txState{tx: tx, iso: iso})

		api.mu.Unlock()

		return
	}

	parentTxState := api.tx[len(api.tx)-1]
	if iso != parentTxState.iso {
		err = fmt.Errorf(
			"unable to begin nested DB tx; requested isolation level for nested tx doesn't match to isolation level of parent tx; requested level: %s; level of parent tx: %s",
			iso,
			parentTxState.iso,
		)

		api.mu.Unlock()

		return
	}

	tx, err = parentTxState.tx.Begin(ctx)
	if err != nil {
		err = fmt.Errorf("error while beginning nested tx for DB; nesting level (0 - root level): %d; %w", len(api.tx), err)
		api.mu.Unlock()

		return
	}

	api.tx = append(api.tx, txState{tx: tx, iso: iso})

	api.mu.Unlock()

	return
}

func (api *impl) RollbackTx(ctx context.Context) (err error) {
	return api.rollbackCommitTx(ctx, false)
}

func (api *impl) CommitTx(ctx context.Context) (err error) {
	return api.rollbackCommitTx(ctx, true)
}

func (api *impl) HasTx() (has bool, iso storage.IsolationLevel) {
	api.mu.Lock()

	if has = len(api.tx) != 0; has {
		iso = api.tx[len(api.tx)-1].iso
	}

	api.mu.Unlock()

	return
}

func (api *impl) IsTxSupported() bool { return true }

func (api *impl) acquireLockIfNoCloseRequest() (err error) {
	if err = api.ensureNoCloseRequest(); err != nil {
		return
	}

	api.mu.Lock()

	if err = api.ensureNoCloseRequest(); err != nil {
		api.mu.Unlock()
	}

	return
}

func (api *impl) ensureNoCloseRequest() error {
	if api.closeRequestMark.Load() == closeRequestYes {
		return fmt.Errorf("close has been requested fot this storage API instance; %w", errs.ErrCloseRequested)
	}

	return nil
}

// nolint:funlen
func (api *impl) requireConn(ctx context.Context) (conn *pgx.Conn, err error) {
	if err = api.acquireLockIfOperable(); err != nil {
		return
	}

	conn = api.conn
	if conn == nil {
		connCloser := api.connCloser

		switch {
		case api.connConf != nil:
			api.shuffleConnCfgHosts()

			conn, err = pgx.ConnectConfig(ctx, api.connConf)
			if err == nil {
				connCloser = defaultPgxConnCloser
			}
		case api.connConnector != nil:
			conn, connCloser, err = api.connConnector(ctx)
		default:
			err = errors.New("neither connection configuration nor connection connector func specified")
		}

		if err != nil {
			err = fmt.Errorf("error while connecting to storage; %w", err)
			api.mu.Unlock()

			return
		}

		api.conn, api.connCloser = conn, connCloser

		err = api.ensureNoCloseRequest()
		if err != nil {
			conn = nil
			api.mu.Unlock()

			return
		}

		logEvt := api.Log().Trace(ctx).WithSendMsg("connection with storage established")

		defer func() {
			logEvt.Send()
		}()
	}

	api.mu.Unlock()

	return
}

func (api *impl) execSQL(ctx context.Context, query querypkg.Named, dst any, oneRow bool, args ...any) (err error) {
	var (
		stats         *querypkg.ExecStats
		execQueryText string
		execQueryArgs []any
		rows          pgx.Rows
	)

	queryText, err := query.Text()
	if err != nil {
		return
	}

	execCtx, execCtxCancel := api.makeExecCtx(ctx, query)
	defer execCtxCancel()

	if stats, err = querypkg.ExecStatsFromContext(ctx); err != nil {
		return
	}

	if execQueryText, execQueryArgs, err = api.formExecSQLPositionalArgs(query, args...); err != nil {
		return
	}

	startTs := time.Now().UTC()
	conn, err := api.requireConn(execCtx)

	stats.SetAcquireConn(time.Since(startTs))

	if err != nil {
		err = fmt.Errorf("error while acquiring DB connection; query name: %s; %w", query.Name(), err)

		return
	}

	startTs = time.Now().UTC()
	rows, err = conn.Query(execCtx, execQueryText, execQueryArgs...)

	stats.SetExec(time.Since(startTs))

	if err != nil {
		if query.Anon() {
			err = fmt.Errorf("error while executing SQL query; query text: %s; %w", queryText, err)
		} else {
			err = fmt.Errorf("error while executing SQL query; query name: %s; %w", query.Name(), err)
		}

		err = normalizePostgresError(err)

		return
	}

	if oneRow {
		startTs = time.Now().UTC()
		err = api.scanOne(rows, dst)

		stats.SetScan(time.Since(startTs))
	} else {
		startTs = time.Now().UTC()
		err = api.scanAll(rows, dst)

		stats.SetScan(time.Since(startTs))
	}

	stats.SetRowsAffected(int(rows.CommandTag().RowsAffected()))

	if err != nil {
		if query.Anon() {
			err = fmt.Errorf("error while doing scan of query result; query text: %s; %w", queryText, err)
		} else {
			err = fmt.Errorf("error while doing scan of query result; query name: %s; %w", query.Name(), err)
		}

		err = normalizePostgresError(err)
	}

	return
}

func (api *impl) scanAll(rows pgx.Rows, dst any) error {
	return api.scan.ScanAll(dst, pgxscan.NewRowsAdapter(rows))
}

func (api *impl) scanOne(rows pgx.Rows, dst any) error {
	err := api.scan.ScanOne(dst, pgxscan.NewRowsAdapter(rows))
	if dbscan.NotFound(err) {
		err = errs.NotFoundWrapWith(err)
	}

	return err
}

func (api *impl) formExecSQLPositionalArgs(query querypkg.Named, args ...any) (queryText string, positionalArgs []any, err error) {
	if queryText, err = query.Text(); err != nil {
		return
	}

	if len(args) == 0 {
		positionalArgs = []any{}

		return
	}

	if query.HasPositionalSyntax() {
		positionalArgs = args

		return
	}

	if api.queryNamedParamDisabled {
		err = errors.New("specified storage query defined with named parameters, but named query parameters disabled for this storage API instance")

		return
	}

	if queryText, positionalArgs, err = querypkg.ToPositionalSyntax(queryText, query.ParamQualifier(), querypkg.DollarSignPositionalSyntaxVariant, args...); err != nil {
		err = fmt.Errorf("error while binding named query params to positional query params; query name: %s; %w", query.Name(), err)
	}

	return
}

func (api *impl) rollbackCommitTx(ctx context.Context, commit bool) (err error) {
	stats, err := querypkg.ExecStatsFromContext(ctx)
	if err != nil {
		return
	}

	startTs := time.Now().UTC()

	defer func() {
		stats.SetExec(time.Since(startTs))
	}()

	if err = api.acquireLockIfOperable(); err != nil {
		return
	}

	if len(api.tx) == 0 {
		if commit {
			err = errs.ErrTxInconsistentCommit
		} else {
			err = errs.ErrTxInconsistentRollback
		}

		api.mu.Unlock()

		return
	}

	txHolder := api.tx[len(api.tx)-1]
	api.tx = api.tx[:len(api.tx)-1]

	if commit {
		err = txHolder.tx.Commit(ctx)
	} else {
		err = txHolder.tx.Rollback(ctx)
	}

	if err != nil {
		switch {
		case len(api.tx) == 0:
			if commit {
				err = fmt.Errorf("DB tx commit fail; %w", err)
			} else {
				err = fmt.Errorf("DB tx rollback fail; %w", err)
			}
		case commit:
			err = fmt.Errorf("DB nested tx commit fail; nesting level (0 - root level): %d; %w", len(api.tx), err)
		default:
			err = fmt.Errorf("DB nexted tx rollback fail; nesting level (0 - root level): %d; %w", len(api.tx), err)
		}
	}

	api.mu.Unlock()

	return
}

func (api *impl) makeExecCtx(p context.Context, q querypkg.Named) (ctx context.Context, cancel context.CancelFunc) {
	execTimeout := q.Timeout()
	if execTimeout.IsZero() {
		execTimeout = api.commandTimeout
	}

	ctx = p
	cancel = func() {}

	if !q.Empty() {
		ctx = querypkg.ExecPropsContext(ctx, q.Name(), execTimeout)
	}

	if !execTimeout.IsInfinite() {
		ctx, cancel = context.WithTimeout(ctx, execTimeout.Dur())
	}

	return
}

func (api *impl) shuffleConnCfgHosts() {
	if len(api.connConf.Fallbacks) == 0 {
		return
	}

	addresses := make([]*pgconn.FallbackConfig, 0, len(api.connConf.Fallbacks)+1)
	addresses = append(addresses, &pgconn.FallbackConfig{
		Host: api.connConf.Host,
		Port: api.connConf.Port,
	})

	for _, fallback := range api.connConf.Fallbacks {
		addresses = append(addresses, &pgconn.FallbackConfig{
			Host: fallback.Host,
			Port: fallback.Port,
		})
	}

	clslices.Shuf(addresses)
	api.connConf.Host = addresses[0].Host
	api.connConf.Port = addresses[0].Port
	api.connConf.Fallbacks = addresses[1:]
}

func normalizePostgresError(err error) error {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pg.SQLState(pgErr.Code) {
		case pg.SQLState23514, pg.SQLState23505:
			err = errs.Wrap(errs.ErrConstraintViolation, err)
		case pg.SqlState40001:
			err = errs.Wrap(errs.ErrTxSerializationFailure, err)
		}
	}

	return err
}

func makeTxOptions(iso storage.IsolationLevel, access storage.AccessMode) (res pgx.TxOptions, err error) {
	switch iso {
	case storage.IsolationLevelUnspecified:
		err = fmt.Errorf("specified tx isolation level is not valid; iso: %s", iso)
	case storage.IsolationLevelReadUncommitted:
		err = fmt.Errorf("specified tx isolation level is not supported; iso: %s; %w", iso, errs.ErrNotSupported)
	case storage.IsolationLevelReadCommitted:
		res = pgx.TxOptions{
			IsoLevel:       pgx.ReadCommitted,
			AccessMode:     access.ToPg(),
			DeferrableMode: pgx.NotDeferrable,
		}
	case storage.IsolationLevelRepeatableRead:
		res = pgx.TxOptions{
			IsoLevel:       pgx.RepeatableRead,
			AccessMode:     access.ToPg(),
			DeferrableMode: pgx.NotDeferrable,
		}
	case storage.IsolationLevelSerializable:
		res = pgx.TxOptions{
			IsoLevel:       pgx.Serializable,
			AccessMode:     access.ToPg(),
			DeferrableMode: pgx.NotDeferrable,
		}
	case storage.IsolationLevelInheritedOrDefault:
		fallthrough
	default:
		err = fmt.Errorf("specified tx isolation level '%s' is not supported; %w", iso, errs.ErrNotSupported)
	}

	return
}

func initDBScan() (*dbscan.API, error) {
	opts := []dbscan.APIOption{
		dbscan.WithScannableTypes(
			(*sql.Scanner)(nil),
			(*pgt.BinaryDecoder)(nil),
			(*pgt.TextDecoder)(nil),
		),
	}

	return dbscan.NewAPI(opts...)
}

func (api *impl) ExecTxFn(ctx context.Context, fn func() error) error {
	return apipkg.ExecTxFn(ctx, api, fn)
}

func (api *impl) ExecTxIsoFn(ctx context.Context, iso storage.IsolationLevel, fn func() error) error {
	return apipkg.ExecTxIsoFn(ctx, api, iso, fn)
}

func (api *impl) ExecTxIsoAccessFn(ctx context.Context, iso storage.IsolationLevel, access storage.AccessMode, fn func() error) error {
	return apipkg.ExecTxIsoAccessFn(ctx, api, iso, access, fn)
}

func New(opts ...OptionFn) (_ apipkg.SQLAPI, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{commandTimeout: timeout.Infinite}
	if err = fopt.Apply(impl, opts...); err != nil {
		return nil, err
	}

	switch {
	case impl.conn == nil:
		if impl.connConnector == nil {
			if impl.connConf == nil {
				return nil, errors.New("neither existing connection object nor connection config nor connector func specified")
			}
		} else if impl.connConf != nil {
			return nil, errors.New("connection config should be unspecified when connector func specified")
		}
	case impl.connConf != nil:
		return nil, errors.New("connection config should be unspecified when existing connection object specified")
	case impl.connConnector != nil:
		return nil, errors.New("connector func should be unspecified when existing connection object specified")
	}

	if impl.connConfDBName == "" {
		err = errors.New("database name either not specified with API config options or specified as empty string")
	}

	if impl.scan, err = initDBScan(); err != nil {
		return nil, err
	}

	return impl, nil
}
