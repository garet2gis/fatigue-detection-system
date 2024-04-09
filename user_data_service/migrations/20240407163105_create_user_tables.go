package migrations

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateUserTables, downCreateUserTables)
}

func upCreateUserTables(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	CREATE TABLE users
	(
	    user_id CHAR(36) PRIMARY KEY,
	    name VARCHAR(64) DEFAULT '',
	    surname VARCHAR(64) DEFAULT '',
	    password VARCHAR(64) NOT NULL,
	    login VARCHAR(64) NOT NULL
	);`)

	if err != nil {
		return err
	}

	return nil
}

func downCreateUserTables(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `DROP TABLE users;`)
	if err != nil {
		return err
	}
	return nil
}
