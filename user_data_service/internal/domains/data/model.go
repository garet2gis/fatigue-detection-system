package data

type FeatureCount struct {
	UserID               string `db:"user_id"`
	FaceModelFeatures    uint64 `db:"face_model_features"`
	FaceModelTrainStatus string `db:"face_model_train_status"`
}
