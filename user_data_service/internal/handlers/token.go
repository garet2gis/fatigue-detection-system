package handlers

import (
	"context"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/app_errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

type TokenHandler struct {
	jwtSecret string
}

func NewTokenHandler(jwtSecret string) *TokenHandler {
	return &TokenHandler{
		jwtSecret: jwtSecret,
	}
}

func (c *TokenHandler) CreateAccessToken(_ context.Context, userID string) (string, error) {
	op := "handlers.CoreHandler.generateContentURL"

	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userID
	delta := time.Hour * 24
	claims["exp"] = time.Now().UTC().Add(delta).Unix()

	tokenString, err := token.SignedString([]byte(c.jwtSecret))
	if err != nil {
		return "", app_errors.ErrInternalServerError.WrapError(op, err.Error())
	}
	return tokenString, nil
}

func (c *TokenHandler) GetUserIDFromToken(_ context.Context, contentToken string) (string, error) {
	op := "handlers.CoreHandler.ParseJWT"

	token, err := jwt.Parse(contentToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, app_errors.ErrParseError.WrapError(op, "fail to parse token")
		}
		return []byte(c.jwtSecret), nil
	})
	if err != nil {
		return "", app_errors.ErrInternalServerError.WrapError(op, err.Error())
	}
	if token == nil {
		return "", app_errors.ErrInternalServerError.WrapError(op, "nil pointer token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", app_errors.ErrParseError.WrapError(op, "failed to parse jwt claims")
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return "", app_errors.ErrParseError.WrapError(op, "failed to find sub field: userID")
	}

	return userID, nil
}
