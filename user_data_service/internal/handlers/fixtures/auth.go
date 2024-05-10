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

type UploadFeaturesURLs struct {
	FaceModel string `json:"face_model"`
}

type LoginResponse struct {
	UserID string `json:"user_id"`

	UploadFeaturesURLs UploadFeaturesURLs     `json:"upload_features"`
	Models             map[string]interface{} `json:"model_urls"`
}

type ModelURLs struct {
	FaceModelURL string
}

func NewLoginResponse(userID, baseURL, tokenString string, models map[string]interface{}) (*LoginResponse, error) {
	op := "fixtures.NewLoginResponse"
	joinedPathFaceModel, err := url.JoinPath(baseURL, "face_model/save_features")
	if err != nil {
		return nil, app_errors.ErrInternalServerError.WrapError(op, err.Error())
	}

	return &LoginResponse{
		UserID: userID,
		UploadFeaturesURLs: UploadFeaturesURLs{
			FaceModel: joinedPathFaceModel + "?access_token=" + tokenString,
		},
		Models: models,
	}, nil
}
