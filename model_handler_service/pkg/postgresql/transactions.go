//go:generate mockgen -source ./transactions.go -destination=./mock/transactor.go -package=mock

package postgresql

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
)

type Transactor interface {
	WithinTransaction(context.Context, func(ctx context.Context) error) error
}

type txKey struct{}

type Tx struct {
	tx pgx.Tx
}

func NewTx(tx pgx.Tx) *Tx {
	return &Tx{
		tx: tx,
	}
}

func injectTx(ctx context.Context, tx Connection) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// ExtractTx extracts transaction from context
func ExtractTx(ctx context.Context) Connection {
	if tx, ok := ctx.Value(txKey{}).(Connection); ok {
		return tx
	}
	return nil
}

func (t Tx) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	return t.tx.Exec(ctx, query, args...)
}

func (t Tx) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, t.tx, dest, query, args...)
}

func (t Tx) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, t.tx, dest, query, args...)
}

func (t Tx) ExecQueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return t.tx.QueryRow(ctx, query, args...)
}

func (t Tx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return t.tx.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (t Tx) RegisterType(ctx context.Context, typeName string) error {
	conn := t.tx.Conn()
	dt, err := conn.LoadType(ctx, typeName)
	if err != nil {
		return err
	}

	conn.TypeMap().RegisterType(dt)
	return nil
}

func (t Tx) Release() {
	// do nothing for realization interface Connection
}

func (db *Database) WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) (err error) {
	// чтобы не было вложенных транзакций
	existTx := ExtractTx(ctx)
	if existTx != nil {
		txErr := tFunc(ctx)
		if txErr != nil {
			return txErr
		}
		return nil
	}

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	txStruct := NewTx(tx)

	err = tFunc(injectTx(ctx, txStruct))
	if err != nil {
		txErr := tx.Rollback(ctx)
		if txErr != nil {
			log.Print(txErr.Error())
			return txErr
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Print(err.Error())
		return err
	}
	return nil
}
