import cv2
import mediapipe as mp
import numpy as np
import math
import csv
import os
import logging
from api.http import send_csv_file
from PyQt5.QtCore import QThread, pyqtSignal

mp_drawing = mp.solutions.drawing_utils
mp_drawing_styles = mp.solutions.drawing_styles
mp_face_mesh = mp.solutions.face_mesh

# For webcam input:
drawing_spec = mp_drawing.DrawingSpec(thickness=1, circle_radius=1)

right_eye = [[33, 133], [160, 144], [159, 145], [158, 153]]  # right eye landmark positions
left_eye = [[263, 362], [387, 373], [386, 374], [385, 380]]  # left eye landmark positions
mouth = [[61, 291], [39, 181], [0, 17], [269, 405]]  # mouth landmark coordinates

SHOW_MESH = frozenset([
    (33, 133), (160, 144), (159, 145), (158, 153),
    (263, 362), (387, 373), (386, 374), (385, 380),
    (61, 291), (39, 181), (0, 17), (269, 405)
])

connections_drawing_spec = mp_drawing.DrawingSpec(
    thickness=1,
    circle_radius=3,
    color=(255, 255, 255)
)


def distance(p1, p2):
    return (((p1[:2] - p2[:2]) ** 2).sum()) ** 0.5


def eye_aspect_ratio(landmarks, eye):
    N1 = distance(landmarks[eye[1][0]], landmarks[eye[1][1]])
    N2 = distance(landmarks[eye[2][0]], landmarks[eye[2][1]])
    N3 = distance(landmarks[eye[3][0]], landmarks[eye[3][1]])
    D = distance(landmarks[eye[0][0]], landmarks[eye[0][1]])
    return (N1 + N2 + N3) / (3 * D)


def eye_feature(landmarks):
    return (eye_aspect_ratio(landmarks, left_eye) + eye_aspect_ratio(landmarks, right_eye)) / 2


def mouth_feature(landmarks):
    N1 = distance(landmarks[mouth[1][0]], landmarks[mouth[1][1]])
    N2 = distance(landmarks[mouth[2][0]], landmarks[mouth[2][1]])
    N3 = distance(landmarks[mouth[3][0]], landmarks[mouth[3][1]])
    D = distance(landmarks[mouth[0][0]], landmarks[mouth[0][1]])
    return (N1 + N2 + N3) / (3 * D)


def perimeter(landmarks, eye):
    return distance(landmarks[eye[0][0]], landmarks[eye[1][0]]) + \
        distance(landmarks[eye[1][0]], landmarks[eye[2][0]]) + \
        distance(landmarks[eye[2][0]], landmarks[eye[3][0]]) + \
        distance(landmarks[eye[3][0]], landmarks[eye[0][1]]) + \
        distance(landmarks[eye[0][1]], landmarks[eye[3][1]]) + \
        distance(landmarks[eye[3][1]], landmarks[eye[2][1]]) + \
        distance(landmarks[eye[2][1]], landmarks[eye[1][1]]) + \
        distance(landmarks[eye[1][1]], landmarks[eye[0][0]])


def perimeter_feature(landmarks):
    return (perimeter(landmarks, left_eye) + perimeter(landmarks, right_eye)) / 2


def area_eye(landmarks, eye):
    return math.pi * ((distance(landmarks[eye[1][0]], landmarks[eye[3][1]]) * 0.5) ** 2)


def head_angle(landmarks, img_w, img_h):
    face_3d = []
    face_2d = []

    for face_landmarks in landmarks:
        for idx, lm in enumerate(face_landmarks.landmark):
            if idx == 33 or idx == 263 or idx == 1 or idx == 61 or idx == 291 or idx == 199:
                x, y = int(lm.x * img_w), int(lm.y * img_h)

                # Get the 2D Coordinates
                face_2d.append([x, y])

                # Get the 3D Coordinates
                face_3d.append([x, y, lm.z])

        # Convert it to the NumPy array
        face_2d = np.array(face_2d, dtype=np.float64)

        # Convert it to the NumPy array
        face_3d = np.array(face_3d, dtype=np.float64)

        # The camera matrix
        focal_length = 1 * img_w

        cam_matrix = np.array([[focal_length, 0, img_h / 2],
                               [0, focal_length, img_w / 2],
                               [0, 0, 1]])

        # The distortion parameters
        dist_matrix = np.zeros((4, 1), dtype=np.float64)

        # Solve PnP
        success, rot_vec, trans_vec = cv2.solvePnP(face_3d, face_2d, cam_matrix, dist_matrix)

        # Get rotational matrix
        rmat, jac = cv2.Rodrigues(rot_vec)

        # Get angles
        angles, mtxR, mtxQ, Qx, Qy, Qz = cv2.RQDecomp3x3(rmat)

        x = angles[0] * 360
        y = angles[1] * 360

        return x, y
def get_features(video_path, pass_frame=2):
    cap = cv2.VideoCapture(video_path)
    features = []
    frame_count = 0

    # every 2 frame
    cur_pass = 0

    with mp_face_mesh.FaceMesh(
            max_num_faces=1,
            refine_landmarks=True,
            min_detection_confidence=0.5,
            min_tracking_confidence=0.5) as face_mesh:
        while cap.isOpened():
            if frame_count == 220:
                break

            success, image = cap.read()
            if not success:
                break

            # skip frames
            cur_pass += 1
            if cur_pass < pass_frame:
                continue
            cur_pass = 0

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

                features.append([frame_count, eye, mouth, area_eye, area_mouth, pupil])
                frame_count += 1

            if cv2.waitKey(5) & 0xFF == 27:
                break

        cap.release()

    return features


class FeatureUploader(QThread):
    finished = pyqtSignal()  # Сигнал о завершении записи

    def __init__(self):
        super().__init__()
        self.video_id = ""
        self.filepath = ""
        self.is_tired = False
        self.url = ""
        self.user_id = ""

    def setup(self, video_id, filepath, is_tired, url, user_id):
        super().__init__()
        self.video_id = video_id
        self.filepath = filepath
        self.is_tired = is_tired
        self.url = url
        self.user_id = user_id

    def run(self):
        header = ['video_id', 'frame_count', 'eye', 'mouth', 'area_eye', 'area_mouth', 'pupil', 'label', 'user_id']

        csv_filename = f'{self.video_id}.csv'
        with open(csv_filename, 'w', encoding='UTF8') as f:
            writer = csv.writer(f)

            # write the header
            writer.writerow(header)

            features = get_features(self.filepath)
            for row in features:
                # write the row
                if self.is_tired:
                    writer.writerow([self.video_id, *row, 1, self.user_id])
                else:
                    writer.writerow([self.video_id, *row, 0, self.user_id])

            logging.info("Создание... " + self.filepath)

        send_csv_file(csv_filename, self.url)
        # удаляем csv
        delete_csv_file(csv_filename)

        self.finished.emit()


class FeatureUploaderForFineTune(QThread):
    finished = pyqtSignal()

    def __init__(self):
        super().__init__()
        self.is_tired = False
        self.url = ""
        self.user_id = ""
        self.video_id = ""
        self.features = []

    def setup(self, features, video_id, is_tired, url, user_id):
        super().__init__()
        self.is_tired = is_tired
        self.url = url
        self.user_id = user_id
        self.features = features
        self.video_id = video_id

    def run(self):
        header = ['video_id', 'frame_count', 'eye', 'mouth', 'area_eye', 'area_mouth', 'pupil', 'label', 'user_id']

        csv_filename = f'{self.video_id}.csv'
        with open(csv_filename, 'w', encoding='UTF8') as f:
            writer = csv.writer(f)

            # write the header
            writer.writerow(header)

            for row in self.features:
                # write the row
                if self.is_tired:
                    writer.writerow([self.video_id, *row, 1, self.user_id])
                else:
                    writer.writerow([self.video_id, *row, 0, self.user_id])

        send_csv_file(csv_filename, self.url)
        # удаляем csv
        # delete_csv_file(csv_filename)

        self.finished.emit()


def delete_csv_file(file_path):
    try:
        if os.path.exists(file_path):
            os.remove(file_path)
            logging.info("Файл успешно удален с диска.")
        else:
            logging.warning("Файл не найден.")
    except Exception as e:
        logging.error(f"Произошла ошибка при удалении файла: {str(e)}")
