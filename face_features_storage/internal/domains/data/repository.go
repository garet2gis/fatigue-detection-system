package data

import (
	"context"
	"encoding/csv"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/garet2gis/fatigue-detection-system/face_features_storage/internal/app_errors"
	"github.com/garet2gis/fatigue-detection-system/face_features_storage/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/face_features_storage/pkg/postgresql"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"strconv"
)

const (
	FeaturesTable = "video_features"
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
	l := logger.EntryWithRequestIDFromContext(ctx)

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
		row[8], _ = strconv.Atoi(record[8])
		row[9] = record[9]

		for i := 2; i < 8; i++ {
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

	l.With(zap.Uint64("count", featuresCount)).Info(fmt.Sprintf("%s: save face model features", op))

	return featuresCount, nil
}
