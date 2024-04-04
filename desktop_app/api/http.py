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
