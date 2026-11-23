package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type PgxConnectorFn func(context.Context) (*pgx.Conn, PgxConnCloserFn, error)
