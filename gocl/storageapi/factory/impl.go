package factory

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/jackc/pgx/v5"
	pgxp "github.com/jackc/pgx/v5/pgxpool"

	"rn.m.vo7.io/vo7io/m/gocl/app/metricstub"
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi/metrics"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi/pg"
)

type (
	impl[TAPI storageapi.SQLAPI] struct {
		fopt.LogOpt
		_                  [0]TAPI
		baseFactory        *baseImpl[TAPI]
		pgConnConf         *pgx.ConnConfig
		pgPool             *pgxp.Pool
		pgPoolConf         *pgxp.Config
		instanceContextKey contextext.Key
		apiConstructor     APIConstructorFn[TAPI]
		baseAPIConstructor func(context.Context) (storageapi.SQLAPI, error)
		closeFn            func(context.Context) error
		apiConf            storageapi.Conf
		connConf           storage.ConnConf
		mu                 sync.Mutex
		closeRequestMark   int32
	}
)

const closeRequestYes = int32(1)

func (s *impl[TAPI]) Base() FactorySQL { return s.baseFactory }

func (s *impl[TAPI]) Context(parent context.Context, api TAPI) context.Context {
	return context.WithValue(parent, s.instanceContextKey, api)
}

// FromContext requires instance of API from the specified context.
//
// If no API provided by specified context, a wrapped error [errs.ErrStorageAPIMissInContext] returned.
func (s *impl[TAPI]) FromContext(ctx context.Context) (res TAPI, err error) {
	existingInstance, ok, err := contextext.Val[TAPI](ctx, s.instanceContextKey)
	if err != nil {
		return
	} else if ok {
		res = existingInstance

		return
	}

	err = fmt.Errorf("%w; context key: %s", errs.ErrStorageAPIMissInContext, s.instanceContextKey)

	return
}

// New returns instance of API accordingly with following:
//
// If ctx contains instance of API, that instance will be returned.
// outCtx will be the same as ctx, closer will be non-nil NOP-function.
//
// If ctx doesn't contain instance of API, new instance will be created and associated with outCtx.
func (s *impl[TAPI]) New(ctx context.Context) (TAPI, func(), context.Context, error) {
	var (
		outCtx = ctx
		zero   TAPI
	)

	existing, ok, err := contextext.Val[TAPI](ctx, s.instanceContextKey)
	if err != nil {
		return zero, nil, outCtx, err
	} else if ok {
		return existing, s.nopCloseFn, outCtx, nil
	}

	newAPI, closer, err := s.createAPI(ctx)
	if err != nil {
		return zero, nil, outCtx, err
	}

	outCtx = context.WithValue(ctx, s.instanceContextKey, newAPI)

	return newAPI, closer, outCtx, nil
}

func (s *impl[TAPI]) StorageType() storage.StorageType { return storage.StorageType(s.connConf.Type) }

func (s *impl[TAPI]) Close(ctx context.Context) (err error) {
	atomic.StoreInt32(&s.closeRequestMark, closeRequestYes)

	s.acquireLock()

	if s.closeFn != nil {
		err = s.closeFn(ctx)
	}

	s.releaseLock()

	return
}

func (s *impl[TAPI]) CloseLog(ctx context.Context, log logext.Logger) {
	var err error

	if log == nil {
		log = s.Log()
	}

	defer func() {
		if err != nil {
			storage.EnrichLogWithConnConfProps(log.Err(ctx, err), s.connConf).
				Msg("error while closing storage API factory")
		}
	}()

	err = s.Close(ctx)
}

func (s *impl[TAPI]) createAPI(ctx context.Context) (_ TAPI, _ func(), err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of storage API '%s'; %w", stringsext.TypeStringFor[TAPI](), err)
		}
	}()

	var zero TAPI

	_, ok, err := contextext.Val[TAPI](ctx, s.instanceContextKey)
	if err != nil {
		return zero, nil, err
	} else if ok {
		// if context contains another instance of API then it means intention to share same instance between works (funcs, components, etc).
		// to avoid erroneous creation of new instance error returned.
		return zero, nil, fmt.Errorf("unable to create a new instance; specified context contains another instance of API under key '%v' already", s.instanceContextKey)
	}

	base, err := s.createBaseAPI(ctx)
	if err != nil {
		return zero, nil, err
	}

	defer func() {
		if err != nil {
			err = errs.Join(err, base.Close(ctx))
		}
	}()

	api, err := s.apiConstructor(ctx, base)
	if err != nil {
		return zero, nil, fmt.Errorf("error while calling API constructor; %w", err)
	}

	closer := func() { api.CloseWithLog(ctx, s.Log()) }

	return api, closer, nil
}

func (s *impl[TAPI]) createBaseAPI(ctx context.Context) (_ storageapi.SQLAPI, err error) {
	if err = s.ensureNoCloseRequest(); err != nil {
		return nil, err
	}

	api, err := s.baseAPIConstructor(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			err = errs.Join(err, api.Close(ctx))
		}
	}()

	if !s.apiConf.MetricsDisabled {
		var metricsInstrumenting storageapi.SQLAPI

		if metricsInstrumenting, err = metrics.NewSQL(api); err != nil {
			return nil, err
		}

		api = metricsInstrumenting
	}

	return api, nil
}

func (s *impl[TAPI]) acquireLock() {
	s.mu.Lock()
}

func (s *impl[TAPI]) acquireLockIfNoCloseRequest() (err error) {
	err = s.ensureNoCloseRequest()
	if err != nil {
		return
	}

	s.acquireLock()

	err = s.ensureNoCloseRequest()
	if err != nil {
		s.releaseLock()
	}

	return
}

func (s *impl[TAPI]) releaseLock() {
	s.mu.Unlock()
}

func (s *impl[TAPI]) ensureNoCloseRequest() error {
	if atomic.LoadInt32(&s.closeRequestMark) == closeRequestYes {
		return fmt.Errorf("close has been requested fot this storage API factory; %w", errs.ErrCloseRequested)
	}

	return nil
}

func (s *impl[TAPI]) createPgAPI(_ context.Context) (storageapi.SQLAPI, error) {
	return pg.New(
		pg.Option().
			Log(s.Log()).
			APIConf(s.apiConf).
			PgxConnConf(s.pgConnConf),
	)
}

func (s *impl[TAPI]) createPgPoolAPI(_ context.Context) (api storageapi.SQLAPI, err error) {
	return pg.
		New(pg.
			Option().
			Log(s.Log()).
			APIConf(s.apiConf).
			Connector(s.pgPoolConnConector, s.pgPoolConf.ConnConfig.Database))
}

func (s *impl[TAPI]) pgPoolConnConector(ctx context.Context) (conn *pgx.Conn, connCloser pg.PgxConnCloserFn, err error) {
	err = s.acquireLockIfNoCloseRequest()
	if err != nil {
		return
	}

	var (
		pool     *pgxp.Pool
		poolConn *pgxp.Conn
	)

	pool = s.pgPool
	if pool == nil {
		pool, err = pgxp.NewWithConfig(ctx, s.pgPoolConf)
		if err != nil {
			s.releaseLock()

			return
		}

		s.pgPool = pool
	}

	s.releaseLock()

	poolConn, err = pool.Acquire(ctx)
	if err != nil {
		return
	}

	metricstub.ResourceOpenOKCount.WithLabelVals(attrname.StorageConn.String(), s.connConf.Type, s.connConf.Database).Inc()

	conn = poolConn.Conn()
	connCloser = func(_ context.Context, _ *pgx.Conn) error {
		poolConn.Release()
		metricstub.ResourceCloseOKCount.WithLabelVals(attrname.StorageConn.String(), s.connConf.Type, s.connConf.Database).Inc()

		return nil
	}

	return
}

// WARNING! must be called under mu.
func (s *impl[TAPI]) pgPoolClose(_ context.Context) (err error) {
	if s.pgPool != nil {
		s.pgPool.Close()
		s.pgPool = nil
	}

	return
}

func (s *impl[TAPI]) logEnrichHook(_ context.Context, _ level.Num, _ error, _ bool, e logext.Event) logext.Event {
	return storage.EnrichLogWithConnConfProps(e, s.connConf)
}

func (*impl[TAPI]) nopCloseFn() {}

func (s *impl[TAPI]) initLog() error {
	logTune := s.Log().Tune()

	if !s.connConf.LogLevel().IsUnspecified() {
		overrideNum := s.connConf.LogLevel().Num()

		if overrideNum < s.Log().LogLevel() {
			logTune = logTune.Level(overrideNum)
		}
	}

	logTune = logTune.NextHook(s.logEnrichHook)

	return s.SetLog(logTune.Tuned())
}

func New[TAPI storageapi.SQLAPI](opts ...OptionFn[TAPI]) (_ Factory[TAPI], err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl[TAPI]](), err)
		}
	}()

	impl := &impl[TAPI]{instanceContextKey: InstanceContextKey}
	if err = fopt.Apply(impl, opts...); err != nil {
		return nil, err
	}

	if impl.instanceContextKey == "" {
		return nil, errors.New("API instance context key is invalid, can't be an empty string")
	}

	if impl.apiConstructor == nil {
		return nil, errors.New("API constructor not specified whereas it should be")
	}

	if err = impl.initLog(); err != nil {
		return nil, err
	}

	switch {
	case impl.connConf.IsPostgreSQL():
		if impl.connConf.PoolEnabled {
			pgPoolConf, confErr := pg.PgxPoolConfFromCommonConnConf(impl.connConf, impl.Log())
			if confErr != nil {
				return nil, confErr
			}

			impl.pgPoolConf = pgPoolConf
			impl.baseAPIConstructor = impl.createPgPoolAPI
			impl.closeFn = impl.pgPoolClose
		} else {
			pgxConnConf, confErr := pg.PgxConnConfFromCommonConnConf(impl.connConf, impl.Log())
			if confErr != nil {
				return nil, confErr
			}

			impl.pgConnConf = pgxConnConf
			impl.baseAPIConstructor = impl.createPgAPI
		}
	default:
		return nil, fmt.Errorf("specified storage type not supported by this factory; storage type: %s; %w", impl.connConf.Type, errs.ErrStorageTypeUnsupported)
	}

	impl.baseFactory = &baseImpl[TAPI]{u: impl}

	return impl, nil
}
