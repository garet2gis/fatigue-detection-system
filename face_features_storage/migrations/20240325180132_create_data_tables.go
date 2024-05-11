package migrations

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateDataTables, downCreateDataTables)
}

func upCreateDataTables(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	CREATE TABLE video_features
	(
	    video_id CHAR(36),
	    
	    frame_count INT,
	    eye DOUBLE PRECISION, 
	    mouth DOUBLE PRECISION,
	    perimeter_eye DOUBLE PRECISION,
	    perimeter_mouth DOUBLE PRECISION,
	    x_angle DOUBLE PRECISION,
	    y_angle DOUBLE PRECISION,
	    
	    label INT,
	    user_id CHAR(36)
-- 	    PRIMARY KEY (video_id, frame_count)
	);`)

	if err != nil {
		return err
	}

	return nil
}

func downCreateDataTables(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `DROP TABLE video_features;`)
	if err != nil {
		return err
	}

	return nil
}
