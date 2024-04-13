package fixtures

import (
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/app_errors"
	"github.com/garet2gis/fatigue-detection-system/user_data_service/internal/domains/data"
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

type ActionModelURLs struct {
	ModelURL          string `json:"model_url"`
	UploadFeaturesURL string `json:"upload_features_url"`
}

type LoginResponse struct {
	FaceModelURL ActionModelURLs `json:"face_model"`
}

type ModelURLs struct {
	FaceModelURL string
}

func NewLoginResponse(baseURL, tokenString string, models []data.MLModel) (*LoginResponse, error) {
	op := "fixtures.NewLoginResponse"
	joinedPathFaceModel, err := url.JoinPath(baseURL, "face_model/get")
	if err != nil {
		return nil, app_errors.ErrInternalServerError.WrapError(op, err.Error())
	}

	var faceModelURL string
	for _, val := range models {
		switch val.ModelType {
		case data.FaceModel:
			if val.ModelURL != nil {
				faceModelURL = *val.ModelURL
			}
		}
	}

	return &LoginResponse{
		FaceModelURL: ActionModelURLs{
			UploadFeaturesURL: joinedPathFaceModel + "?access_token=" + tokenString,
			ModelURL:          faceModelURL,
		},
	}, nil
}
