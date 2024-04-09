package tools

import (
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/app_errors"
	"golang.org/x/crypto/bcrypt"
)

func Hash(str string) (string, error) {
	op := "tools.Hash"

	hashedStr, err := bcrypt.GenerateFromPassword([]byte(str), bcrypt.DefaultCost)
	if err != nil {
		return "", app_errors.ErrParseError.WrapError(op, err.Error())
	}
	return string(hashedStr), nil
}
