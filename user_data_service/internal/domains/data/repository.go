package data

import (
	"context"
	"encoding/csv"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/postgresql"
	"github.com/jackc/pgx/v5"
	"io"
	"mime/multipart"
	"strconv"
)

const (
	VideoTable    = "video_features"
	FeaturesTable = "videos"
)

type Repository struct {
	db postgresql.DB
}

func NewRepository(db postgresql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CopyCSV(ctx context.Context, file multipart.File) error {
	reader := csv.NewReader(file)
	header, _ := reader.Read()

	var rows [][]interface{}

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
	}

	_, err := r.db.Client(ctx).CopyFrom(ctx, pgx.Identifier{VideoTable}, header, pgx.CopyFromRows(rows))
	if err != nil {
		return err
	}
	return nil
}
