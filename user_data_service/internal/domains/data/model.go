package data

type MLModel struct {
	UserID           string  `db:"user_id"`
	ModelFeatures    uint64  `db:"features_count"`
	ModelTrainStatus string  `db:"train_status"`
	ModelType        string  `db:"model_type"`
	ModelURL         *string `db:"model_url"`
}
