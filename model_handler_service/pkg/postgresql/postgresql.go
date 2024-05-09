package postgresql

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/georgysavva/scany/v2/pgxscan"
)

// PGXdb интерфейс для БД
type PGXdb interface {
	Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error)
	ExecQueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

type DB interface {
	Client(ctx context.Context) PGXdb
	Acquire(ctx context.Context) (Connection, error)
	Close()
}

func (db Database) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, query, args...)
}

func (db Database) Select(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Select(ctx, db.pool, dest, query, args...)
}

func (db Database) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return pgxscan.Get(ctx, db.pool, dest, query, args...)
}

func (db Database) ExecQueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return db.pool.QueryRow(ctx, query, args...)
}

func (db Database) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return db.pool.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

type Database struct {
	pool *pgxpool.Pool
}

func (db *Database) Client(ctx context.Context) PGXdb {
	tx := ExtractTx(ctx)

	if tx != nil {
		return tx
	}
	return db
}

func (db *Database) Acquire(ctx context.Context) (Connection, error) {
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

func (db *Database) Close() {
	db.pool.Close()
}

type DBConfig struct {
	Port                  string
	Host                  string
	Name                  string
	Password              string
	Username              string
	MaxConnectionAttempts int

	AutoMigrate   bool
	MigrationsDir string
}

func NewClient(ctx context.Context, sc DBConfig) (db *Database, err error) {
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

		db = &Database{
			pool: pool,
		}

		return nil
	}, sc.MaxConnectionAttempts, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed: do with tries connect to postgresql, error: %w", err)
	}

	if err = db.pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed: ping postgresql, error: %w", err)
	}

	if sc.AutoMigrate {
		dbString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			sc.Host,
			sc.Username,
			sc.Password,
			sc.Name,
			sc.Port)

		err = migrateUp(dbString, sc.MigrationsDir)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func migrateUp(dbString, migrationDir string) error {
	db, err := goose.OpenDBWithDriver("pgx", dbString)
	if err != nil {
		return fmt.Errorf("goose: failed to open DB: %w\n", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v\n", err)
		}
	}()

	err = goose.Up(db, migrationDir)
	if err != nil {
		return fmt.Errorf("goose up: %w", err)
	}

	return nil
}

func DoWithTries(fn func() error, attempts int, delay time.Duration) (err error) {
	for attempts > 0 {
		if err = fn(); err != nil {
			time.Sleep(delay)
			attempts--
			continue
		}
		return nil
	}
	return
}
