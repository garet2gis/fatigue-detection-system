package data

import (
	"context"
	"encoding/csv"
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
	FeaturesTable   = "video_features"
	UsedVideosTable = "used_videos"
)

type Repository struct {
	db           postgresql.DB
	queryBuilder sq.StatementBuilderType
}

func NewRepository(db postgresql.DB) *Repository {
	return &Repository{db: db, queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar)}
}

func (r *Repository) CopyCSV(ctx context.Context, csvFile multipart.File) error {
	reader := csv.NewReader(csvFile)
	header, _ := reader.Read()

	batchLen := 250
	rows := make([][]interface{}, 0, batchLen)

	for {
		record, err := reader.Read()
		if err != nil {
			// Проверяем ошибку окончания файла
			if err == io.EOF {
				break // Достигли конца файла, выходим из цикла
			}
			return err
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
				return err
			}

			rows = make([][]interface{}, 0, batchLen)
		}
	}

	if len(rows) > 0 {
		_, err := r.db.Client(ctx).CopyFrom(ctx, pgx.Identifier{FeaturesTable}, header, pgx.CopyFromRows(rows))
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) CreateUsedVideo(ctx context.Context, videoID string) error {
	l := logger.EntryWithRequestIDFromContext(ctx)
	setMap := sq.Eq{
		"video_id": videoID,
	}

	qb := r.queryBuilder.
		Insert(UsedVideosTable)

	qb = qb.SetMap(setMap)
	q, i, _ := qb.ToSql()

	_, err := r.db.Client(ctx).Exec(ctx, q, i...)
	if err != nil {
		l.With(zap.Error(err), zap.String("query", q)).Error("failed to exec query")
		return app_errors.ErrSQLExec.SetMessage(err.Error())
	}

	return nil
}
