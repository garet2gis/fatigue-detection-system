import numpy as np
import pandas as pd
from sklearn.preprocessing import StandardScaler
from sklearn.model_selection import train_test_split, cross_val_score
import xgboost
import logging
import os
import requests


def create_xgb(data, url, user_id, model_type, features_count):
    data = data.drop(columns=['video_id'])
    data = data.drop(columns=['frame_count'])
    data = data.drop(columns=['user_id'])

    df_num = data.select_dtypes(include=[np.number])
    df_cat = data.select_dtypes(include=[object])
    num_cols = df_num.columns.values[:-1]

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

    df = data[num_cols]

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

    file_path = './models/tmp.xgb'
    xgb.save_model(file_path)
    send_model(file_path, url, user_id, model_type, features_count)
    delete_file(file_path)


def delete_file(file_path):
    try:
        if os.path.exists(file_path):
            os.remove(file_path)
            logging.info("Файл успешно удален с диска.")
        else:
            logging.warning("Файл не найден.")
    except Exception as e:
        logging.error(f"Произошла ошибка при удалении файла: {str(e)}")


def send_model(file_path, url, user_id, model_type, features_count):
    try:
        with open(file_path, 'rb') as file:
            data = {
                'user_id': user_id,
                'model_type': model_type,
                'features_count': str(features_count)
            }

            files = {'file': file}
            response = requests.post(url, data=data, files=files)
            if str(response.status_code).startswith('2'):
                logging.info(f"Файл успешно отправлен по HTTP: {file_path}")
            else:
                logging.warning(f"Произошла ошибка при отправке файла: {response.status_code}")
    except Exception as e:
        logging.error(f"Произошла ошибка: {str(e)}")
