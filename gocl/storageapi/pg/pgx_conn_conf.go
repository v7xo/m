package pg

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/jackc/pgx/v5"
	pgxp "github.com/jackc/pgx/v5/pgxpool"

	appinfo "rn.m.vo7.io/vo7io/m/gocl/appmeta/info"
	logging "rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
)

func PgxPoolConfFromCommonConnConf(conf storage.ConnConf, log logging.Logger) (pgxpCfg *pgxp.Config, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error while constructing PostgreSQL pooling connection config from common connection config; err: %w", err)
		}
	}()

	if !conf.PoolEnabled {
		err = errors.New("configuration not specifies use of connection pooling; use PgxConnConfigFromCommonConnConfig method instead")

		return
	}

	connUrl, err := pgxConnUrlStrFromCommonConnConfig(conf, false)
	if err != nil {
		return
	}

	pgxpCfg, err = pgxp.ParseConfig(connUrl)
	if err != nil {
		err = fmt.Errorf("error while parsing PostgreSQL DB connection pool config string; config string: %s; err: %w", connUrl, err)

		return
	}

	pgxpCfg.ConnConfig.DefaultQueryExecMode, err = conf.StmtCacheMode().AsPgxQueryExecMode()
	if err != nil {
		return
	}

	pgxpCfg.ConnConfig.Tracer = newTraceLogAdapter(log, conf.LogLevel())

	return
}

func PgxConnConfFromCommonConnConf(conf storage.ConnConf, log logging.Logger) (pgxConf *pgx.ConnConfig, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error while constructing PostgreSQL connection config from common connection config; err: %w", err)
		}
	}()

	if conf.PoolEnabled {
		err = errors.New("configuration specifies use of connection pooling; use PgxPoolConfigFromCommonConnConfig method instead")

		return
	}

	connUrl, err := pgxConnUrlStrFromCommonConnConfig(conf, true)
	if err != nil {
		return
	}

	pgxConf, err = pgx.ParseConfig(connUrl)
	if err != nil {
		err = fmt.Errorf("error while parsing PostgreSQL connection config string; config string: %s; err: %w", connUrl, err)

		return
	}

	pgxConf.DefaultQueryExecMode, err = conf.StmtCacheMode().AsPgxQueryExecMode()
	if err != nil {
		return
	}

	pgxConf.Tracer = newTraceLogAdapter(log, conf.LogLevel())

	return
}

func pgxConnUrlStrFromCommonConnConfig(cfg storage.ConnConf, ignorePoolParams bool) (connUrl string, err error) {
	connTimeoutAsInt := int(cfg.ConnectTimeoutDur.Seconds())
	if float64(connTimeoutAsInt) != cfg.ConnectTimeoutDur.Seconds() {
		err = fmt.Errorf(
			"unable to build PostgreSQL connection config; connection timeout should be aligned to seconds; actual connection timeout: %s",
			cfg.ConnectTimeoutDur,
		)

		return
	}

	connUrl = fmt.Sprintf(
		`postgres://%s:%s@%s/%s?sslmode=disable&connect_timeout=%d&application_name=%s`,
		cfg.User,
		cfg.Password,
		net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		cfg.Database,
		connTimeoutAsInt,
		url.QueryEscape(appinfo.UserAgentShort()),
	)

	if !ignorePoolParams {
		connUrl = fmt.Sprintf(
			`%s&pool_max_conns=%d&pool_min_conns=%d&pool_max_conn_lifetime=%s&pool_max_conn_idle_time=%s&pool_health_check_period=%s`,
			connUrl,
			cfg.PoolMaxConns,
			cfg.PoolMinConns,
			cfg.PoolConnMaxLifetimeDur.String(),
			cfg.PoolConnIdleTimeDur.String(),
			cfg.PoolHealthCheckPeriod.String(),
		)
	}

	return
}
