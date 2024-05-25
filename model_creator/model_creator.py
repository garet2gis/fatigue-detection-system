import numpy as np
import pandas as pd
from sklearn.preprocessing import StandardScaler
from sklearn.model_selection import train_test_split, cross_val_score
import xgboost
import logging
import os
import requests

# create_xgb - функция создания модели градиентного бустинга
def create_xgb(data, url, user_id, model_type, features_count):
    # Удаляем ненужные в обучении колонки
    data = data.drop(columns=['video_id'])
    data = data.drop(columns=['frame_count'])
    data = data.drop(columns=['user_id'])

    # берем числовые колонки
    df_num = data.select_dtypes(include=[np.number])
    num_cols = df_num.columns.values[:-1]
    df = data[num_cols]

    # определяем дата-фрейм признаков и соответствующих им классов
    X = df
    Y = data["label"]

    # разделяем данные на тестовые и тренировочные
    X_train, X_test, y_train, y_test = train_test_split(X, Y, test_size=0.2, random_state=42)

    # задаем гиперпараметры
    xgb_best_params = {'colsample_bytree': 0.8, 'eta': 0.1, 'max_depth': 9, 'n_estimators': 800, 'subsample': 0.7}
    # объявляем и обучаем модель
    xgb = xgboost.XGBClassifier(**xgb_best_params)
    xgb.fit(X_train, y_train)

    cv_scores = cross_val_score(xgb, X_test, y_test, cv=5)

    # Вывод результатов
    logging.info("Cross-validation scores:", cv_scores)
    logging.info("Average accuracy:", cv_scores.mean())

    file_path = './models/tmp.xgb'
    xgb.save_model(file_path)
    # Отправялем модель в сервис работы с моделями
    send_model(file_path, url, user_id, model_type, features_count)
    # Удаляем модель
    delete_file(file_path)

# delete_file - функция удаления файла
def delete_file(file_path):
    try:
        # если файл существует, то удаляем его
        if os.path.exists(file_path):
            os.remove(file_path)
            logging.info("Файл успешно удален с диска.")
        else:
            logging.warning("Файл не найден.")
    except Exception as e:
        logging.error(f"Произошла ошибка при удалении файла: {str(e)}")

# send_model - функция отправления файла модели
def send_model(file_path, url, user_id, model_type, features_count):
    try:
        # открываем файл обученной модели
        with open(file_path, 'rb') as file:
            # задаем строковые поля формы
            data = {
                'user_id': user_id,
                'model_type': model_type,
                'features_count': str(features_count)
            }
            # задаем файл модели в поле file
            files = {'file': file}
            # отправляем http-запросом модель и другие данные
            response = requests.post(url, data=data, files=files)
            #  логируем успешность выполненного запроса
            if str(response.status_code).startswith('2'):
                logging.info(f"Файл успешно отправлен по HTTP: {file_path}")
            else:
                logging.warning(f"Произошла ошибка при отправке файла: {response.status_code}")
    except Exception as e:
        logging.error(f"Произошла ошибка: {str(e)}")
