package storage

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/misc/timeout"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type ConnConf struct {
	Database string `validate:"gt=0"`
	Host     string
	Password string
	Type     string
	User     string `validate:"gt=0"`

	stmtCacheMode string
	logLevel      string

	PoolMaxConns int `validate:"gte=1"`
	PoolMinConns int `validate:"gte=0"`
	Port         int `validate:"gt=-1,lt=65536"`

	ConnectTimeoutDur time.Duration `validate:"min=-1ns"`

	PoolConnIdleTimeDur    time.Duration `validate:"min=1s"`
	PoolConnMaxLifetimeDur time.Duration `validate:"min=1s"`
	PoolHealthCheckPeriod  time.Duration `validate:"min=1s"`

	PoolEnabled bool
}

func (c *ConnConf) ConnectionTimeout() timeout.Timeout { return timeout.Timeout(c.ConnectTimeoutDur) }

func (c *ConnConf) IsTarantool() bool { return StorageType(c.Type).IsTarantool() }

func (c *ConnConf) IsPostgreSQL() bool { return StorageType(c.Type).IsPostgreSQL() }

func (c *ConnConf) StmtCacheMode() StmtCacheMode { return StmtCacheMode(c.stmtCacheMode) }

func (c *ConnConf) LogLevel() level.Name { return level.Name(c.logLevel) }

func (c *ConnConf) SetLogLevel(val level.Name) { c.logLevel = val.String() }

func (c *ConnConf) BindFlagSetPrefix(prefix string, flset *pflag.FlagSet) error {
	prefix = flenv.NormalizePrefix(flenv.NormalizePrefix(prefix) + ConnConfFlenvSelfFlagPrefix)

	flset.StringVar(
		&c.logLevel,
		prefix+"log-level",
		"",
		"specifies log level for this storage connection"+
			"\n\tthis log level overrides the level of logger that passed to connection constructor",
	)

	flset.StringVar(
		&c.Type,
		prefix+"type",
		PostgreSQLStorageType.String(),
		fmt.Sprintf("storage type (for ex.: %s, %s)", PostgreSQLStorageType, TarantoolStorageType),
	)
	flset.IntVar(
		&c.Port,
		prefix+"port",
		DefaultPostgreSQLConnPort,
		"storage connection address port",
	)

	flset.StringVar(
		&c.Host,
		prefix+"host",
		"",
		"storage connection address host",
	)
	flset.StringVar(
		&c.Database,
		prefix+"database",
		"",
		"storage database name to connect to",
	)
	flset.DurationVar(
		&c.ConnectTimeoutDur,
		prefix+"connect-timeout",
		DefaultConnConnectTimeout,
		"storage connection timeout in form of time.Duration",
	)
	flset.StringVar(
		&c.User,
		prefix+"user",
		"",
		"storage connection user name",
	)
	flset.StringVar(
		&c.Password,
		prefix+"password",
		"",
		"storage connection user password",
	)

	flset.StringVar(
		&c.stmtCacheMode,
		prefix+"statement-cache-mode",
		DefaultStmtCacheMode.String(),
		"storage query statement cache mode (if applicable for DBMS type)",
	)

	flset.BoolVar(
		&c.PoolEnabled,
		prefix+"pool-enabled",
		DefaultPoolEnabled,
		"enable storage connection pooling",
	)
	flset.DurationVar(
		&c.PoolConnIdleTimeDur,
		prefix+"pool-conn-idle-time",
		DefaultPoolConnIdleTime,
		fmt.Sprintf("duration after which an idle storage connection will be automatically closed and evicted from the pool (see also '%s')", prefix+PoolHealthCheckPeriodConfigFlagName),
	)
	flset.DurationVar(
		&c.PoolConnMaxLifetimeDur,
		prefix+"pool-conn-max-lifetime",
		DefaultPoolConnMaxLifetime,
		fmt.Sprintf("duration since storage connection creation after which connection will be automatically closed and evicted from the pool (see also '%s')", prefix+PoolHealthCheckPeriodConfigFlagName),
	)
	flset.DurationVar(
		&c.PoolHealthCheckPeriod,
		prefix+PoolHealthCheckPeriodConfigFlagName,
		DefaultPoolHealthCheckPeriod,
		"duration between storage pool checks of the health of idle (or lifetime expired) storage connections",
	)
	flset.IntVar(
		&c.PoolMaxConns,
		prefix+"pool-max-conns",
		DefaultPoolMaxConns,
		"max size of storage pool",
	)
	flset.IntVar(
		&c.PoolMinConns,
		prefix+"pool-min-conns",
		DefaultPoolMinConns,
		fmt.Sprintf("min size of storage pool; health check will increase the number of storage connections to this amount if it had dropped below (see also '%s')", prefix+PoolHealthCheckPeriodConfigFlagName),
	)

	return nil
}

func (c *ConnConf) BindFlags(fls *pflag.FlagSet) error {
	return c.BindFlagSetPrefix("", fls)
}

func (c *ConnConf) Validate(v validation.Validator) error {
	l := level.Name(c.logLevel)

	if !l.IsKnown() && !l.IsUnspecified() {
		return fmt.Errorf("specified log level is unknown; level: %s", c.logLevel)
	}

	return v.Struct(c)
}

// [FlenvRootConnConf] returns storage connection configuration as it would be the root configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvRootConnConf() (ConnConf, error) {
	return FlenvMemberConnConf("")
}

// [FlenvMemberConnConf] returns storage connection configuration as it would be the member of parent configuration object.
//
// Configuration forms from values of application flags and environment variables (latter used as fallback source of configuration, if no flag value specified).
func FlenvMemberConnConf(memberPrefix string) (c ConnConf, err error) {
	err = flenv.PopulatePrefix(memberPrefix, &c)

	return
}
