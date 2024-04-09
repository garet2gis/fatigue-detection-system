package data

import (
	"context"
	"encoding/csv"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/app_errors"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/postgresql"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"strconv"
)

const (
	FeaturesTable      = "video_features"
	FeaturesCountTable = "features_count"
)

type Repository struct {
	db           postgresql.DB
	queryBuilder sq.StatementBuilderType
}

func NewRepository(db postgresql.DB) *Repository {
	return &Repository{db: db, queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar)}
}

func (r *Repository) SaveFaceVideoFeatures(ctx context.Context, csvFile multipart.File) (uint64, error) {
	op := "data.Repository.SaveFaceVideoFeatures"
	reader := csv.NewReader(csvFile)
	header, _ := reader.Read()

	batchLen := 250
	rows := make([][]interface{}, 0, batchLen)

	var featuresCount uint64

	for {
		record, err := reader.Read()
		if err != nil {
			// Проверяем ошибку окончания файла
			if err == io.EOF {
				break // Достигли конца файла, выходим из цикла
			}
			return 0, app_errors.ErrParseError.WrapError(op, err.Error())
		}

		row := make([]interface{}, len(record))

		row[0] = record[0]

		row[1], _ = strconv.Atoi(record[1])
		row[7], _ = strconv.Atoi(record[7])
		row[8], _ = strconv.Atoi(record[8])

		for i := 2; i < 7; i++ {
			row[i], _ = strconv.ParseFloat(record[i], 64)
		}

		rows = append(rows, row)
		if len(rows) == batchLen {
			_, err = r.db.Client(ctx).CopyFrom(ctx, pgx.Identifier{FeaturesTable}, header, pgx.CopyFromRows(rows))
			if err != nil {
				return 0, app_errors.ErrSQLExec.WrapError(op, err.Error())
			}

			featuresCount += uint64(len(rows))
			rows = make([][]interface{}, 0, batchLen)
		}
	}

	if len(rows) > 0 {
		featuresCount += uint64(len(rows))
		_, err := r.db.Client(ctx).CopyFrom(ctx, pgx.Identifier{FeaturesTable}, header, pgx.CopyFromRows(rows))
		if err != nil {
			return 0, app_errors.ErrSQLExec.WrapError(op, err.Error())
		}
	}

	return featuresCount, nil
}

func (r *Repository) IncrementFeaturesCount(ctx context.Context, userID string, faceFeaturesCount uint64) error {
	op := "data.Repository.IncrementFeaturesCount"
	l := logger.EntryWithRequestIDFromContext(ctx)

	newValueString := fmt.Sprintf("face_video_features + %d", faceFeaturesCount)

	q, i, err := r.queryBuilder.
		Update(FeaturesCountTable).
		Set("face_video_features", sq.Expr(newValueString)).
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	_, err = r.db.Client(ctx).Exec(ctx, q, i...)
	if err != nil {
		return app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	l.With(
		zap.String("user_id", userID),
		zap.Uint64("face_video_features", faceFeaturesCount),
	).Info(fmt.Sprintf("%s: increase feature count", op))

	return nil
}
