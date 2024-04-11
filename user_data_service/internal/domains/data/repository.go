package data

import (
	"context"
	"encoding/csv"
	"errors"
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
	FeaturesTable = "video_features"
	ModelsTable   = "models"
)

const (
	StatusNotTrain       = "not_train"
	StatusInTrainProcess = "in_train_process"
	StatusInTuneProcess  = "in_tune_process"
	StatusTrained        = "train"
)

const (
	FaceModel = "face_model"
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

	l.With(zap.Uint64("count", featuresCount)).Info(fmt.Sprintf("%s: save face model features", op))

	return featuresCount, nil
}

func (r *Repository) ChangeFeaturesCount(ctx context.Context, userID, modelType string, faceFeaturesCount int) error {
	op := "data.Repository.IncrementFeaturesCount"
	l := logger.EntryWithRequestIDFromContext(ctx)

	var newValueString string
	if faceFeaturesCount > 0 {
		newValueString = fmt.Sprintf("face_model_features + %d", faceFeaturesCount)
	} else {
		newValueString = fmt.Sprintf("face_model_features - %d", -faceFeaturesCount)
	}

	q, i, err := r.queryBuilder.
		Update(ModelsTable).
		Set("features_count", sq.Expr(newValueString)).
		Where(sq.Eq{"user_id": userID}).
		Where(sq.Eq{"model_type": modelType}).
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
		zap.String("model_type", modelType),
		zap.Int("face_model_features_delta", faceFeaturesCount),
	).Info(fmt.Sprintf("%s: increase feature count", op))

	return nil
}

func (r *Repository) SetModelURL(ctx context.Context, url, modelType, userID string) error {
	op := "data.Repository.SetModelURL"
	l := logger.EntryWithRequestIDFromContext(ctx)

	qb := r.queryBuilder.
		Update(ModelsTable).
		Set("model_url", url).
		Where(sq.Eq{"user_id": userID}, sq.Eq{"model_type": modelType})

	q, i, err := qb.ToSql()
	if err != nil {
		return app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	// выполняем запрос
	_, err = r.db.Client(ctx).Exec(ctx, q, i...)
	if err != nil {
		return app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	l.With(zap.String("model_url", url), zap.String("user_id", userID), zap.String("model_type", modelType)).
		Info(fmt.Sprintf("%s: set new model_url to model", op))

	return nil
}

func (r *Repository) CreateModel(ctx context.Context, userID, modelType string) error {
	op := "data.Repository.CreateModel"
	l := logger.EntryWithRequestIDFromContext(ctx)

	setMap := sq.Eq{
		"user_id":    userID,
		"model_type": modelType,
	}

	q, i, err := r.queryBuilder.
		Insert(ModelsTable).
		SetMap(setMap).
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
		zap.String("model_type", modelType),
	).Info(fmt.Sprintf("%s: create model", op))

	return nil
}

func (r *Repository) GetModelByUserID(ctx context.Context, userID, modelType string) (*MLModel, error) {
	op := "data.Repository.GetModelByUserID"
	l := logger.EntryWithRequestIDFromContext(ctx)

	q, i, err := r.queryBuilder.
		Select(
			"user_id",
			"features_count",
			"train_status",
			"model_url",
			"model_type",
		).
		From(ModelsTable).
		Where(sq.Eq{"user_id": userID}).
		Where(sq.Eq{"model_type": modelType}).
		ToSql()
	if err != nil {
		return nil, app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	var res MLModel
	err = r.db.Client(ctx).Get(ctx, &res, q, i...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app_errors.ErrNotFound.WrapError(op, err.Error())
		}
		return nil, app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	l.With(
		zap.String("user_id", userID),
		zap.String("model_type", modelType),
	).Info(fmt.Sprintf("%s: get model by userID", op))

	return &res, nil
}

func (r *Repository) ViewNotLearnedModels(ctx context.Context, modelType string, trainThreshold uint64) ([]MLModel, error) {
	op := "data.Repository.ViewNotLearnedModels"
	l := logger.EntryWithRequestIDFromContext(ctx)

	q, i, err := r.queryBuilder.
		Select(
			"user_id",
			"features_count",
			"train_status",
			"model_url",
			"model_type",
		).
		From(ModelsTable).
		Where(
			sq.Eq{"train_status": StatusNotTrain},
			sq.Eq{"model_type": modelType},
			sq.GtOrEq{"features_count": trainThreshold},
		).
		ToSql()
	if err != nil {
		return nil, app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	var res []MLModel
	err = r.db.Client(ctx).Select(ctx, &res, q, i...)
	if err != nil {
		return nil, app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	l.With(zap.Int("count", len(res))).Info(fmt.Sprintf("%s: find not learned models", op))

	return res, nil
}

func (r *Repository) ViewNotFineTunedFaceModels(ctx context.Context, modelType string, tuneThreshold uint64) ([]MLModel, error) {
	op := "data.Repository.ViewNotLearnedModels"
	l := logger.EntryWithRequestIDFromContext(ctx)

	q, i, err := r.queryBuilder.
		Select(
			"user_id",
			"features_count",
			"train_status",
			"model_url",
			"model_type",
		).
		From(ModelsTable).
		Where(
			sq.Eq{"train_status": StatusTrained},
			sq.Eq{"model_type": modelType},
			sq.GtOrEq{"features_count": tuneThreshold},
		).
		ToSql()
	if err != nil {
		return nil, app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	var res []MLModel
	err = r.db.Client(ctx).Select(ctx, &res, q, i...)
	if err != nil {
		return nil, app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	l.With(zap.Int("count", len(res))).Info(fmt.Sprintf("%s: find models for fine tuning", op))

	return res, nil
}

func (r *Repository) SetModelStatus(ctx context.Context, status, modelType string, userID string) error {
	op := "data.Repository.SetModelStatus"
	l := logger.EntryWithRequestIDFromContext(ctx)

	qb := r.queryBuilder.
		Update(ModelsTable).
		Set("train_status", status).
		Where(sq.Eq{"user_id": userID}, sq.Eq{"model_type": modelType})

	q, i, err := qb.ToSql()
	if err != nil {
		return app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	// выполняем запрос
	_, err = r.db.Client(ctx).Exec(ctx, q, i...)
	if err != nil {
		return app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	l.With(zap.String("status", status), zap.String("user_id", userID)).
		Info(fmt.Sprintf("%s: set new status to model", op))

	return nil
}
