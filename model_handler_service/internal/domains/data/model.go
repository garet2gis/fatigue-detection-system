package data

type MLModel struct {
	UserID           string  `db:"user_id"`
	ModelFeatures    uint64  `db:"features_count"`
	ModelTrainStatus string  `db:"train_status"`
	ModelType        string  `db:"model_type"`
	S3Key            *string `db:"s3_key"`
}
