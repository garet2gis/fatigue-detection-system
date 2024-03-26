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
	CREATE TABLE used_videos
	(
	    video_id CHAR(36) PRIMARY KEY
	);`)

	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
	CREATE TABLE video_features
	(
	    video_id CHAR(36),
	    
	    frame_count INT,
	    eye DOUBLE PRECISION, 
	    mouth DOUBLE PRECISION, 
	    area_eye DOUBLE PRECISION, 
	    area_mouth DOUBLE PRECISION, 
	    pupil DOUBLE PRECISION, 
	    
	    label INT,
	    user_id INT, 
	    PRIMARY KEY (video_id, frame_count)
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

	_, err = tx.ExecContext(ctx, `DROP TABLE videos;`)
	if err != nil {
		return err
	}

	return nil
}
