package fixtures

import (
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/app_errors"
	"net/url"
)

type RegisterRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
}

type LoginRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	GetFaceModelURL            string `json:"get_face_model_url"`
	UploadFaceModelFeaturesURL string `json:"upload_face_model_features_url"`
}

func NewLoginResponse(baseURL, tokenString string) (*LoginResponse, error) {
	op := "fixtures.NewLoginResponse"
	joinedPathFaceModel, err := url.JoinPath(baseURL, "face_model/get")
	if err != nil {
		return nil, app_errors.ErrInternalServerError.WrapError(op, err.Error())
	}

	joinedPathUploadFaceModelFeatures, err := url.JoinPath(baseURL, "face_model/save_features")
	if err != nil {
		return nil, app_errors.ErrInternalServerError.WrapError(op, err.Error())
	}
	return &LoginResponse{
		GetFaceModelURL:            joinedPathFaceModel + "?access_token=" + tokenString,
		UploadFaceModelFeaturesURL: joinedPathUploadFaceModelFeatures + "?access_token=" + tokenString,
	}, nil
}
