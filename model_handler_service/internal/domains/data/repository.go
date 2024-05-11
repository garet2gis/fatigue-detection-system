package data

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/app_errors"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/pkg/postgresql"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

const (
	ModelsTable = "models"
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

func (r *Repository) ChangeFeaturesCount(ctx context.Context, userID, modelType string, featuresCount int) error {
	op := "data.Repository.ChangeFeaturesCount"
	l := logger.EntryWithRequestIDFromContext(ctx)

	var newValueString string
	if featuresCount > 0 {
		newValueString = fmt.Sprintf("features_count + %d", featuresCount)
	} else {
		newValueString = fmt.Sprintf("features_count - %d", -featuresCount)
	}

	q, i, err := r.queryBuilder.
		Update(ModelsTable).
		Set("features_count", sq.Expr(newValueString)).
		Where(sq.Eq{"user_id": userID, "model_type": modelType}).
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
		zap.Int("face_model_features_delta", featuresCount),
	).Info(fmt.Sprintf("%s: change feature count", op))

	return nil
}

func (r *Repository) SetFeaturesCountUsed(ctx context.Context, userID, modelType string, featuresCount int) error {
	op := "data.Repository.SetFeaturesCountUsed"
	l := logger.EntryWithRequestIDFromContext(ctx)

	q, i, err := r.queryBuilder.
		Update(ModelsTable).
		Set("features_count_used", featuresCount).
		Where(sq.Eq{"user_id": userID, "model_type": modelType}).
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
		zap.Int("features_count_used", featuresCount),
	).Info(fmt.Sprintf("%s: set features_count_used count", op))

	return nil
}

func (r *Repository) SetModelS3Key(ctx context.Context, s3Key, modelType, userID string) error {
	op := "data.Repository.SetModelS3Key"
	l := logger.EntryWithRequestIDFromContext(ctx)

	qb := r.queryBuilder.
		Update(ModelsTable).
		Set("s3_key", s3Key).
		Where(sq.Eq{"user_id": userID, "model_type": modelType})

	q, i, err := qb.ToSql()
	if err != nil {
		return app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	// выполняем запрос
	_, err = r.db.Client(ctx).Exec(ctx, q, i...)
	if err != nil {
		return app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	l.With(zap.String("s3_key", s3Key), zap.String("user_id", userID), zap.String("model_type", modelType)).
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
			"s3_key",
			"model_type",
		).
		From(ModelsTable).
		Where(sq.Eq{"user_id": userID, "model_type": modelType}).
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

func (r *Repository) GetModelsByUserID(ctx context.Context, userID string) ([]MLModel, error) {
	op := "data.Repository.GetModelsByUserID"
	l := logger.EntryWithRequestIDFromContext(ctx)

	q, i, err := r.queryBuilder.
		Select(
			"user_id",
			"features_count",
			"train_status",
			"s3_key",
			"model_type",
		).
		From(ModelsTable).
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return nil, app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	var res []MLModel
	err = r.db.Client(ctx).Select(ctx, &res, q, i...)
	if err != nil {
		return nil, app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	l.With(zap.Int("count", len(res))).Info(fmt.Sprintf("%s: find models by user_id", op))

	return res, nil
}

func (r *Repository) ViewNotLearnedModels(ctx context.Context, modelType string, trainThreshold uint64) ([]MLModel, error) {
	op := "data.Repository.ViewNotLearnedModels"
	l := logger.EntryWithRequestIDFromContext(ctx)

	q, i, err := r.queryBuilder.
		Select(
			"user_id",
			"features_count",
			"train_status",
			"s3_key",
			"model_type",
		).
		From(ModelsTable).
		Where(
			sq.Eq{"train_status": StatusNotTrain, "model_type": modelType},
		).
		Where(sq.GtOrEq{"features_count": trainThreshold}).
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
			"s3_key",
			"model_type",
		).
		From(ModelsTable).
		Where(
			sq.Eq{"train_status": StatusTrained,
				"model_type": modelType,
			},
		).
		Where(sq.GtOrEq{"features_count - features_count_used": tuneThreshold}).
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
		Where(sq.Eq{"user_id": userID, "model_type": modelType})

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
