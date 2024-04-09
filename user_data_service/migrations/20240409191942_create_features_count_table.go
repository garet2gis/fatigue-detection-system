package migrations

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateFeaturesCountTable, downCreateFeaturesCountTable)
}

func upCreateFeaturesCountTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	CREATE TYPE train_status AS ENUM ('not_train','in_train_process','train');
	;`)

	_, err = tx.ExecContext(ctx, `
	CREATE TABLE features_count
	(
	    user_id CHAR(36) PRIMARY KEY,
	    face_model_features INT NOT NULL DEFAULT(0),
	    face_model_train_status train_status NOT NULL DEFAULT('not_train')
	);`)

	if err != nil {
		return err
	}

	return nil
}

func downCreateFeaturesCountTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `DROP TABLE features_count;`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `DROP TYPE IF EXISTS train_status CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
