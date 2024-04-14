import time

from preprocess.feature_uploader import (eye_feature, mouth_feature,
                                         area_eye_feature, area_mouth_feature,
                                         pupil_feature, mp_face_mesh, mp_drawing, SHOW_MESH, connections_drawing_spec)
import xgboost
import numpy as np
import cv2
import pandas as pd
from .limited_array import LimitedSizeArray
from .circular_queue import CircularQueue
from PyQt5.QtCore import QThread, pyqtSignal
import requests
import joblib


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


class FaceXGBModel(QThread):
    predictionSignal = pyqtSignal(str)

    def __init__(self, model, limited_array_size=16, buf_capacity=1000):
        super().__init__()
        self.limited_array_size = limited_array_size
        self.check_awake = LimitedSizeArray(limited_array_size)

        self.loaded_scaler = joblib.load('./models/face_model/standard_scaler.joblib')

        self.face_model = model
        self.running = True
        self.pause = False

        self.buf_capacity = buf_capacity
        self.last_features = CircularQueue(buf_capacity)
        self.frame_count = 0

    def stop(self):
        self.running = False

    def set_pause(self):
        self.pause = True

    def set_continue(self):
        self.pause = False

    def get_last_features(self):
        return self.last_features.get_raw_array()

    def run(self):
        cap = cv2.VideoCapture(1)

        while self.running:
            if self.pause is True:
                time.sleep(2)
                continue

            with mp_face_mesh.FaceMesh(
                    max_num_faces=1,
                    refine_landmarks=True,
                    min_detection_confidence=0.5,
                    min_tracking_confidence=0.5) as face_mesh:

                success, image = cap.read()
                if not success:
                    break

                # To improve performance, optionally mark the image as not writeable to
                # pass by reference.
                image.flags.writeable = False
                image = cv2.cvtColor(image, cv2.COLOR_BGR2RGB)
                results = face_mesh.process(image)

                # Draw the face mesh annotations on the image.
                image.flags.writeable = True
                image = cv2.cvtColor(image, cv2.COLOR_RGB2BGR)
                if results.multi_face_landmarks:
                    landmarks_positions = []
                    # assume that only face is present in the image
                    for _, data_point in enumerate(results.multi_face_landmarks[0].landmark):
                        landmarks_positions.append(
                            [data_point.x, data_point.y, data_point.z])  # saving normalized landmark positions

                    landmarks_positions = np.array(landmarks_positions)
                    landmarks_positions[:, 0] *= image.shape[1]
                    landmarks_positions[:, 1] *= image.shape[0]

                    eye = eye_feature(landmarks_positions)
                    mouth = mouth_feature(landmarks_positions)
                    area_eye = area_eye_feature(landmarks_positions)
                    area_mouth = area_mouth_feature(landmarks_positions)
                    pupil = pupil_feature(landmarks_positions)

                    self.last_features.enqueue([self.frame_count, eye, mouth, area_eye, area_mouth, pupil])
                    self.frame_count += 1

                    if self.frame_count > self.buf_capacity:
                        self.frame_count = 0

                    features = self.loaded_scaler.transform(pd.DataFrame({
                        'eye': [eye],
                        'mouth': [mouth],
                        'area_eye': [area_eye],
                        'area_mouth': [area_mouth],
                        'pupil': [pupil]
                    }))

                    prediction = self.face_model.predict(xgboost.DMatrix(features))

                    self.check_awake.push(0 if prediction[0] < 0.5 else 1)
                    label = 'Tired'
                    if self.check_awake.count_zeros() >= self.limited_array_size / 2:
                        label = 'Awake'

                    print(label)

                    self.predictionSignal.emit(label)

        cap.release()
