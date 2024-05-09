package app_errors

import (
	"net/http"
)

var (
	UnknownError = NewAppError(
		"UnknownError",
		"",
		0,
		http.StatusTeapot,
	)

	ErrInternalServerError = NewAppError(
		"InternalServerError",
		"internal server error",
		1,
		http.StatusBadRequest)

	ErrSQLExec = NewAppError(
		"SQLError",
		"SQL execution error",
		2,
		http.StatusInternalServerError)

	ErrNotFound = NewAppError(
		"NotFound",
		"entity not found",
		3,
		http.StatusNotFound)

	ErrValidationError = NewAppError(
		"ValidationError",
		"validation error",
		4,
		http.StatusBadRequest)

	ErrParseError = NewAppError(
		"ParseError",
		"parse error",
		5,
		http.StatusBadRequest)

	ErrNoAuthorizationHeader = NewAppError(
		"NoAuthorizationHeader",
		"no authorization header",
		6,
		http.StatusUnauthorized)

	ErrWrongToken = NewAppError(
		"WrongToken",
		"wrong token",
		7,
		http.StatusUnauthorized)
)
