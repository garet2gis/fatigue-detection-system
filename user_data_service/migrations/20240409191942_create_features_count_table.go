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
	CREATE TYPE train_status AS ENUM ('not_train','in_train_process','train', 'in_tune_process');
	;`)

	_, err = tx.ExecContext(ctx, `
	CREATE TYPE model_type AS ENUM ('face_model');
	;`)

	_, err = tx.ExecContext(ctx, `
	CREATE TABLE models
	(
	    user_id CHAR(36) NOT NULL,
	    model_type model_type NOT NULL,
	    features_count INT NOT NULL DEFAULT(0),
	    train_status train_status NOT NULL DEFAULT('not_train'),
	    model_url VARCHAR(128) DEFAULT NULL,
	    
	    CONSTRAINT pkey_user_id_model_type PRIMARY KEY (user_id, model_type)
	);`)

	if err != nil {
		return err
	}

	return nil
}

func downCreateFeaturesCountTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `DROP TABLE models;`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `DROP TYPE IF EXISTS train_status CASCADE;`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `DROP TYPE IF EXISTS model_type CASCADE;`)
	if err != nil {
		return err
	}
	return nil
}
