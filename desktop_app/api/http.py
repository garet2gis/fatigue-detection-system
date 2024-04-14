import requests
import logging


def send_csv_file(file_path, url):
    try:
        with open(file_path, 'rb') as file:
            files = {'file': file}
            response = requests.post(url, files=files)
            if str(response.status_code).startswith('2'):
                logging.info(f"Файл успешно отправлен по HTTP: {file_path}")
            else:
                logging.warning(f"Произошла ошибка при отправке файла: {response.status_code}")
    except Exception as e:
        logging.error(f"Произошла ошибка: {str(e)}")


def perform_login(login_url, login, password):
    try:
        response = requests.post(login_url, json={'login': login, 'password': password})
        if response.status_code == 200:
            data = response.json()
            return True, 'Успешный вход!', data
        else:
            return False, 'Ошибка входа!', None
    except Exception as e:
        return False, str(e), None


def perform_register(register_url, login, password):
    try:
        response = requests.post(register_url, json={'login': login, 'password': password})
        if response.status_code == 204:
            return True, 'Успешная регистрация!'
        else:
            return False, 'Ошибка регистрации!'
    except Exception as e:
        return False, str(e)


def download_model(model_url):
    response = requests.get(model_url)
    if response.status_code == 200:
        # Сохраняем модель в файл
        with open('model.xgb', 'wb') as f:
            f.write(response.content)
        return True, 'Модель загружена!'
    else:
        return False, 'Ошибка загрузки модели!'


def post_features(features_url, features):
    response = requests.post(features_url, json=features)
    if response.status_code == 200:
        return True, 'Данные отправлены успешно!'
    else:
        return False, 'Ошибка отправки данных!'
