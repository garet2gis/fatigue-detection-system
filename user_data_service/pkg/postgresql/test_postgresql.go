package postgresql

import (
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"strings"
	"sync"
	"testing"
	"time"
)

type TBD struct {
	mu   *sync.Mutex
	pool *pgxpool.Pool
}

func (db TBD) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, db.pool, dest, query, args...)
}

func (db TBD) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, db.pool, dest, query, args...)
}

func (db TBD) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, query, args...)
}

func (db TBD) ExecQueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return db.pool.QueryRow(ctx, query, args...)
}

func (db TBD) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return db.pool.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func (db TBD) Client(ctx context.Context) PGXdb {
	tx := ExtractTx(ctx)

	if tx != nil {
		return tx
	}
	return db
}

func (db TBD) Acquire(ctx context.Context) (Connection, error) {
	tx := ExtractTx(ctx)
	if tx != nil {
		return tx, nil
	}

	conn, err := db.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	return NewPGXPoolConnection(conn), nil
}

func (db TBD) Close() {
	db.pool.Close()
}

func (db TBD) SetUp(t *testing.T) {
	t.Helper()
	db.mu.Lock()
	db.Truncate(context.Background())
}

func (db TBD) TearDown() {
	defer db.mu.Unlock()
	db.Truncate(context.Background())
}

func (db TBD) Truncate(ctx context.Context) {
	var tables []string

	getTablesQuery := `SELECT table_name 
						FROM information_schema.tables 
						WHERE table_schema = 'public' 
						  AND table_type = 'BASE TABLE'
						  AND table_name != 'goose_db_version'`

	err := db.Select(ctx, &tables, getTablesQuery)
	if err != nil {
		log.Fatal("failed to truncate table in tests environment")
	}

	if len(tables) == 0 {
		log.Fatal("not found tables, maybe forget to run migrations")
	}

	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", strings.Join(tables, ","))
	_, err = db.Exec(ctx, query)
	if err != nil {
		log.Fatal(err)
	}
}

type TestDBConfig struct {
	Port         string
	Host         string
	Name         string
	Password     string
	Username     string
	MigrationDir string
}

func NewTestClient(ctx context.Context, sc TestDBConfig) (db *TBD, err error) {
	maxConnectionAttempts := 5

	var dsn string
	if sc.Password == "" {
		dsn = fmt.Sprintf("postgresql://%s@%s:%s/%s", sc.Username, sc.Host, sc.Port, sc.Name)
	} else {
		dsn = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", sc.Username, sc.Password, sc.Host, sc.Port, sc.Name)
	}
	err = DoWithTries(func() error {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			return err
		}

		db = &TBD{
			pool: pool,
			mu:   &sync.Mutex{},
		}

		return nil
	}, maxConnectionAttempts, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed: do with tries connect to postgresql, error: %w", err)
	}

	if err = db.pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed: ping postgresql, error: %w", err)
	}

	dbString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		sc.Host,
		sc.Username,
		sc.Password,
		sc.Name,
		sc.Port)

	err = migrateUp(dbString, sc.MigrationDir)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db TBD) WithinTransaction(ctx context.Context, tFunc func(ctx context.Context) error) (err error) {
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
