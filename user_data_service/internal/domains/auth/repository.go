package auth

import (
	"context"
	"errors"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/app_errors"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/postgresql"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/tools"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	UsersTable = "users"
)

type Repository struct {
	db           postgresql.DB
	queryBuilder sq.StatementBuilderType
}

func NewRepository(db postgresql.DB) *Repository {
	return &Repository{db: db, queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar)}
}

func (r *Repository) CreateUser(ctx context.Context, model User) (string, error) {
	op := "auth.Repository.CreateUser"
	l := logger.EntryWithRequestIDFromContext(ctx)

	userUUID, err := uuid.NewUUID()
	if err != nil {
		return "", app_errors.ErrInternalServerError.WrapError(op, err.Error())
	}
	userID := userUUID.String()

	model.PasswordHash, err = tools.Hash(model.PasswordHash)
	if err != nil {
		return "", app_errors.ErrInternalServerError.WrapError(op, err.Error())
	}

	setMap := sq.Eq{
		"user_id":       userID,
		"password_hash": model.PasswordHash,
		"login":         model.Login,
		"name":          model.Name,
		"surname":       model.Surname,
	}

	q, i, err := r.queryBuilder.
		Insert(UsersTable).
		SetMap(setMap).
		ToSql()
	if err != nil {
		return "", app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	_, err = r.db.Client(ctx).Exec(ctx, q, i...)
	if err != nil {
		return "", app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	l.Info(fmt.Sprintf("%s: create user", op))

	return userID, nil
}

func (r *Repository) GetUserByLogin(ctx context.Context, login string) (*User, error) {
	op := "auth.Repository.GetUserByLogin"

	q, i, err := r.queryBuilder.
		Select(
			"user_id",
			"login",
			"name",
			"surname",
			"password_hash",
		).
		From(UsersTable).
		Where(sq.Eq{"login": login}).
		ToSql()
	if err != nil {
		return nil, app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	var user User
	err = r.db.Client(ctx).Get(ctx, &user, q, i...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, app_errors.ErrNotFound.WrapError(op, err.Error())
		}
		return nil, app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	return &user, nil
}
