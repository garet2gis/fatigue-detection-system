package postgresql

import (
	"context"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Connection interface {
	PGXdb
	RegisterType(ctx context.Context, typeName string) error
	Release()
}

type pgxPoolConnection struct {
	poolConn *pgxpool.Conn
}

func NewPGXPoolConnection(poolConn *pgxpool.Conn) *pgxPoolConnection {
	return &pgxPoolConnection{poolConn: poolConn}
}

func (c pgxPoolConnection) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	return c.poolConn.Exec(ctx, query, args...)
}

func (c pgxPoolConnection) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, c.poolConn, dest, query, args...)
}

func (c pgxPoolConnection) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, c.poolConn, dest, query, args...)
}

func (c pgxPoolConnection) ExecQueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return c.poolConn.QueryRow(ctx, query, args...)
}

func (c pgxPoolConnection) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return c.poolConn.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (c pgxPoolConnection) RegisterType(ctx context.Context, typeName string) error {
	conn := c.poolConn.Conn()
	dt, err := conn.LoadType(ctx, typeName)
	if err != nil {
		return err
	}

	conn.TypeMap().RegisterType(dt)
	return nil
}

func (c pgxPoolConnection) Release() {
	c.poolConn.Release()
}
