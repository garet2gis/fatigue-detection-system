import time

from preprocess.feature_uploader import (eye_feature, mouth_feature,
                                         perimeter, perimeter_feature,
                                         head_angle, mp_face_mesh, mouth)
import xgboost
import numpy as np
import cv2
import pandas as pd
from .limited_array import LimitedSizeArray
from .circular_queue import CircularQueue
from PyQt5.QtCore import QThread, pyqtSignal
import requests


class FaceModelLoader(QThread):
    loaded = pyqtSignal(object)

    def __init__(self, url):
        super().__init__()
        self.url = url

    def run(self):
        response = requests.get(self.url)
        if response.status_code == 200:
            # Предполагаем, что модель сохранена в бинарном формате XGB
            filename = './models/face_model/model.xgb'
            with open(filename, 'wb') as f:
                f.write(response.content)
            model = xgboost.Booster()
            model.load_model(filename)
            self.loaded.emit(model)
        else:
            self.loaded.emit(None)


# FaceXGBModel - предсказывает состояния усталости и отрисовывает кадры в отдельном потоке
class FaceXGBModel(QThread):
    # Сигнал для результата предсказания
    predictionSignal = pyqtSignal(str)
    # Сигнал для отправления захваченного кадра
    frameSignal = pyqtSignal(object)

    # Конструктор
    def __init__(self, model, limited_array_size=16, buf_capacity=900):
        super().__init__()
        # Количество кадров, которые должны быть классом tired
        self.limited_array_size = limited_array_size
        # Объявяем массив, которая хранит последние {limited_array_size} предсказаний
        self.check_awake = LimitedSizeArray(limited_array_size)

        # Задаем модель
        self.face_model = model
        # Статусы работы класса
        self.running = True
        self.pause = False

        # Количество последних признаков для хранения
        self.buf_capacity = buf_capacity
        # Объявяем циклическую очередь, которая хранит последние {buf_capacity} признаков
        self.last_features = CircularQueue(buf_capacity)
        # Объявяем счетчик кадров
        self.frame_count = 0

    # stop - метод остановки
    def stop(self):
        self.running = False

    # set_pause - метод приостановки
    def set_pause(self):
        self.pause = True

    # set_continue - метод продолжения
    def set_continue(self):
        self.pause = False

    # get_last_features - метод возвращающий последние признаки в формате списка
    def get_last_features(self):
        return self.last_features.get_raw_array()

    # run - метод запуска класса
    def run(self):
        # Захватываем видео веб-камеры
        cap = cv2.VideoCapture(1)

        # Пока не остановили делаем цикл
        while self.running:
            # Если в режиме паузы ждем 2 секунды
            if self.pause is True:
                time.sleep(2)
                continue

            # Инициализируем класс face_mesh
            with mp_face_mesh.FaceMesh(
                    max_num_faces=1,
                    refine_landmarks=True,
                    min_detection_confidence=0.5,
                    min_tracking_confidence=0.5) as face_mesh:

                # Берем текущий кадр с веб-камеры
                success, image = cap.read()
                if not success:
                    break

                # Отправляем кадр на отрисовку
                self.frameSignal.emit(image)

                # Преобразуем кадр в оттенки серого
                image = cv2.cvtColor(image, cv2.COLOR_BGR2RGB)
                # Вычисляем сетку лица с помощью MediaPipe
                results = face_mesh.process(image)
                # Вычисляем значения размера кадра
                img_h, img_w, img_c = image.shape

                # Если лицевые метки были найдены
                if results.multi_face_landmarks:
                    # Вычисляем углы поворота головы
                    x, y = head_angle(results.multi_face_landmarks, img_h, img_w)

                    # Преобразуем точки лица в np.array
                    landmarks_positions = []
                    for _, data_point in enumerate(results.multi_face_landmarks[0].landmark):
                        landmarks_positions.append(
                            [data_point.x, data_point.y, data_point.z])
                    landmarks_positions = np.array(landmarks_positions)
                    landmarks_positions[:, 0] *= image.shape[1]
                    landmarks_positions[:, 1] *= image.shape[0]

                    # Вычисляем признак - EAR
                    ear = eye_feature(landmarks_positions)
                    # Вычисляем признак - MAR
                    mar = mouth_feature(landmarks_positions)
                    # Вычисляем признак - периметр глаз
                    perimeter_eye = perimeter_feature(landmarks_positions)
                    # Вычисляем признак - периметр рта
                    perimeter_mouth = perimeter(landmarks_positions, mouth)

                    # Сохраняем признаки в циклическую очередь
                    self.last_features.enqueue([self.frame_count, ear, mar, perimeter_eye, perimeter_mouth, x, y])

                    # Увеличиваем счетчик кадров
                    self.frame_count += 1

                    # Обнуляем счетчик, когда преодолели значение buf_capacity
                    if self.frame_count > self.buf_capacity:
                        self.frame_count = 0

                    # Преобразуем признаки в data frame
                    features = pd.DataFrame({
                        'eye': [ear],
                        'mouth': [mar],
                        'perimeter_eye': [perimeter_eye],
                        'perimeter_mouth': [perimeter_mouth],
                        'x_angle': [x],
                        'y_angle': [y],
                    })

                    # Предсказываем состояние усталости
                    prediction = self.face_model.predict(xgboost.DMatrix(features))

                    # Сохраняем предсказание
                    self.check_awake.push(0 if prediction[0] < 0.5 else 1)

                    # Если все значения в check_awake = 0, то только тогда устанавливаем значение
                    # label = Текущее состояние: уставшее
                    label = 'Текущее состояние: не уставшее'
                    if self.check_awake.count_zeros() == 0:
                        label = 'Текущее состояние: уставшее'

                    # Отправляем текущее состояние на отрисовку
                    self.predictionSignal.emit(label)

        # высвобождаем захват видео с веб-камеры
        cap.release()
