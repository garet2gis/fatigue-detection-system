package fixtures

type IncreaseFeaturesRequest struct {
	ModelType     string `json:"model_type"  validate:"required"`
	UserID        string `json:"user_id"  validate:"required"`
	FeaturesCount int    `json:"features_count"  validate:"required"`
}

type GetModelsRequest struct {
	UserID string `json:"string"  validate:"required"`
}
