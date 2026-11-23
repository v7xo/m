package pg

import (
	"fmt"

	"github.com/jackc/pgx/v5"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/storage"
	"rn.m.vo7.io/vo7io/m/gocl/storageapi"
)

type OptionFn func(*impl) error

func (o OptionFn) Log(val logext.Logger) OptionFn {
	return fopt.WithLogScope(o, val, LogScope)
}

// ConnConf option applies all DB common config props and parses specified common config to pgx.ConnConf.
func (o OptionFn) ConnConf(val storage.ConnConf) OptionFn {
	return fopt.ChainErr(o, func(t *impl) error { return applyConnConf(t, val) })
}

func (o OptionFn) PgxConnConf(val *pgx.ConnConfig) OptionFn {
	return fopt.Chain(
		o,
		func(t *impl) {
			t.conn = nil
			t.connConf = val
			t.connConfDBName = val.Database
			t.connCloser = nil
			t.connConnector = nil
		},
	)
}

func (o OptionFn) Conn(val *pgx.Conn, owns bool) OptionFn {
	return fopt.Chain(
		o,
		func(t *impl) {
			t.conn = val
			t.connConf = nil
			t.connConfDBName = val.Config().Database
			t.connConnector = nil

			if owns {
				t.connCloser = defaultPgxConnCloser
			} else {
				t.connCloser = nil
			}
		},
	)
}

func (o OptionFn) Connector(val PgxConnectorFn, dbName string) OptionFn {
	return fopt.Chain(
		o,
		func(t *impl) {
			t.conn = nil
			t.connConf = nil
			t.connConfDBName = dbName
			t.connCloser = nil
			t.connConnector = val
		},
	)
}

func (o OptionFn) APIConf(val storageapi.Conf) OptionFn {
	return fopt.Chain(o, func(t *impl) { applyAPIConf(t, val) })
}

func applyAPIConf(api *impl, cfg storageapi.Conf) {
	api.commandTimeout = cfg.CommandTimeout()
	api.name = cfg.Name
	api.queryNamedParamDisabled = cfg.QueryNamedParameterDisabled
}

func applyConnConf(api *impl, conf storage.ConnConf) error {
	if !storage.StorageType(conf.Type).IsPostgreSQL() {
		return fmt.Errorf("storage type denoted by configuration is '%s' whereas '%s' expected", conf.Type, storage.PostgreSQLStorageType)
	}

	pgxConnCfg, err := PgxConnConfFromCommonConnConf(conf, nil)
	if err != nil {
		return err
	}

	api.conn = nil
	api.connConf = pgxConnCfg
	api.connConfDBName = pgxConnCfg.Database
	api.connCloser = nil
	api.connConnector = nil

	return nil
}

func Option() OptionFn { return fopt.Nop[impl] }
