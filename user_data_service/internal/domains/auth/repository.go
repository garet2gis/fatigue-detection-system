package auth

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/app_errors"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/logger"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/postgresql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(model.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", app_errors.ErrInternalServerError.WrapError(op, err.Error())
	}

	setMap := sq.Eq{
		"user_id":  userID,
		"password": hashedPassword,
		"login":    model.Login,
		"name":     model.Name,
		"surname":  model.Surname,
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

func (r *Repository) GetUser(ctx context.Context, userID string) (*User, error) {
	op := "auth.Repository.GetUser"

	q, i, err := r.queryBuilder.
		Select(
			"user_id",
			"login",
			"name",
			"surname",
			"password",
		).
		From(UsersTable).
		Where(sq.Eq{"user_id": userID}).
		ToSql()
	if err != nil {
		return nil, app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	var user User
	err = r.db.Client(ctx).Get(ctx, &user, q, i...)
	if err != nil {
		return nil, app_errors.ErrSQLExec.WrapError(op, err.Error())
	}

	return &user, nil
}
