package pg

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type PgxConnCloserFn func(context.Context, *pgx.Conn) error

func defaultPgxConnCloser(ctx context.Context, conn *pgx.Conn) error {
	return conn.Close(ctx)
}
