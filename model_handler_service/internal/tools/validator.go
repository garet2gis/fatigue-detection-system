package tools

import (
	"errors"
	"github.com/garet2gis/fatigue-detection-system/model_handler_service/internal/app_errors"
	"github.com/go-playground/validator/v10"
)

func ValidateStruct(validate *validator.Validate, model interface{}) error {
	op := "tools.ValidateStruct"
	err := validate.Struct(model)

	if err != nil {
		var invalid *validator.InvalidValidationError
		if errors.As(err, &invalid) {
			return app_errors.ErrInternalServerError.WrapError(op, err.Error())
		}

		return app_errors.ErrValidationError.WrapError(op, err.Error())
	}
	return nil
}
