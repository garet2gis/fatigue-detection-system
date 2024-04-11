package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/app_errors"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/domains/auth"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/handlers/fixtures"
	customTools "github.com/garet2gis/fatigue-detection-system/user_data_service/internal/tools"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/api"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/pkg/logger"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

// Register godoc
//
//	@Summary	Принимает данные пользователя
//	@ID			register
//	@Tags		auth
//	@Param		user_data	body	fixtures.RegisterRequest	true	"Данные для регистрации"
//	@Success	204
//	@Failure	400	{object}	app_errors.AppError
//	@Router		/auth/register [post]
func (c *CoreHandler) Register(_ http.ResponseWriter, r *http.Request) error {
	op := "handlers.CoreHandler.Register"

	var req fixtures.RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return app_errors.ErrParseError.WrapError(op, err.Error())
	}

	appErr := customTools.ValidateStruct(c.validator, req)
	if appErr != nil {
		return appErr
	}

	_, err = c.authRepository.CreateUser(r.Context(), auth.User{
		Name:         req.Name,
		Surname:      req.Surname,
		PasswordHash: req.Password,
		Login:        req.Login,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// Login godoc
//
//	@Summary	Принимает данные пользователя для входа в систему
//	@ID			login
//	@Tags		auth
//	@Param		user_credentials	body		fixtures.LoginRequest	true	"Данные для логина"
//	@Success	200					{object}	fixtures.LoginResponse
//	@Failure	400					{object}	app_errors.AppError
//	@Router		/auth/login [post]
func (c *CoreHandler) Login(w http.ResponseWriter, r *http.Request) error {
	op := "handlers.CoreHandler.Login"
	l := logger.EntryWithRequestIDFromContext(r.Context())

	var req fixtures.LoginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return app_errors.ErrParseError.WrapError(op, err.Error())
	}

	appErr := customTools.ValidateStruct(c.validator, req)
	if appErr != nil {
		return appErr
	}

	user, err := c.authRepository.GetUserByLogin(r.Context(), req.Login)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return app_errors.ErrUnauthorized
	}

	tokenString, err := c.tokenGenerator.CreateAccessToken(r.Context(), user.UserID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	res, err := fixtures.NewLoginResponse(c.BaseURL, tokenString)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	api.WriteSuccess(r.Context(), w, res, http.StatusOK, l)

	return nil
}