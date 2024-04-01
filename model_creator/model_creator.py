import numpy as np
import pandas as pd
from sklearn.preprocessing import StandardScaler
from sklearn.model_selection import train_test_split, cross_val_score
import xgboost
import logging


def create_xgb(data):
    data = data.drop(columns=['video_id'])
    data = data.drop(columns=['frame_count'])
    data = data.drop(columns=['user_id'])

    df_num = data.select_dtypes(include=[np.number])
    df_cat = data.select_dtypes(include=[object])
    num_cols = df_num.columns.values[:-1]
    cat_cols = df_cat.columns.values

    data.dropna(inplace=True, axis=0)

    for col in num_cols:
        Q1, Q3 = data.loc[:, col].quantile([0.25, 0.75]).values
        IQR = Q3 - Q1
        box_max = Q3 + (1.5 * IQR)
        box_min = Q1 - (1.5 * IQR)
        data.loc[data[col] < box_min, col] = np.NaN
        data.loc[data[col] > box_max, col] = np.NaN

    for col in num_cols:
        cur_mean = np.mean(data[col])
        data[col] = data[col].fillna(cur_mean)

    data.dropna(inplace=True, axis=0)

    scaler = StandardScaler()
    data_norm = scaler.fit_transform(data[num_cols])
    df = pd.DataFrame(data=data_norm, columns=num_cols)
    df[cat_cols] = data[cat_cols].values
    df = pd.get_dummies(df, columns=cat_cols)

    X = df
    Y = data["label"]

    X_train, X_test, y_train, y_test = train_test_split(X, Y, test_size=0.2, random_state=42)

    xgb_best_params = {'colsample_bytree': 0.8, 'eta': 0.1, 'max_depth': 9, 'n_estimators': 800, 'subsample': 0.7}
    xgb = xgboost.XGBClassifier(**xgb_best_params)
    xgb.fit(X_train, y_train)

    cv_scores = cross_val_score(xgb, X_test, y_test, cv=5)

    # Вывод результатов
    logging.info("Cross-validation scores:", cv_scores)
    logging.info("Average accuracy:", cv_scores.mean())

    xgb.save_model('./models/xgb.xgb')
