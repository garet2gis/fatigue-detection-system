from PyQt5.QtWidgets import QWidget, QLabel, QPushButton, QVBoxLayout, QMessageBox
from xgboost_predictor.video_predictor import FaceXGBModel, FaceModelLoader
from preprocess.feature_uploader import FeatureUploaderForFineTune
import cv2
from PyQt5.QtGui import QImage, QPixmap
from PyQt5.QtCore import Qt


class PredictorWindow(QWidget):
    def __init__(self, cfg):
        super().__init__()
        self.setGeometry(300, 300, 200, 100)
        self.setWindowTitle('XGBoost Video Processing')
        self.label = QLabel('Ожидание предсказаний...')
        self.image_label = QLabel(self)
        self.label.resize(180, 60)
        self.label.move(10, 20)

        self.init_upload = QPushButton('Отправить фичи', self)
        self.init_upload.clicked.connect(self.ask_to_record_video)

        self.layout = QVBoxLayout()

        self.layout.addWidget(self.label)
        self.layout.addWidget(self.image_label)
        self.layout.addWidget(self.init_upload)

        self.setLayout(self.layout)

        face_model_url = cfg['model_urls']['face_model']
        self.upload_features_url = cfg['upload_features']['face_model']
        self.user_id = cfg['user_id']
        self.model_loader = FaceModelLoader(face_model_url)
        self.model_loader.loaded.connect(self.on_model_loaded)
        self.model_loader.start()

        self.feature_uploader = FeatureUploaderForFineTune()
        self.feature_uploader.finished.connect(self.continue_prediction)

    def on_model_loaded(self, model):
        if model is not None:
            self.label.setText('Model loaded, processing video...')
            self.video_processor = FaceXGBModel(model)
            self.video_processor.predictionSignal.connect(self.update_prediction)
            self.video_processor.frameSignal.connect(self.update_frame)
            self.video_processor.start()
        else:
            self.label.setText('Failed to load model.')

    def ask_to_record_video(self):
        reply = QMessageBox.question(self, 'Отправить фичи', 'Вы на самом деле устали?',
                                     QMessageBox.Yes | QMessageBox.No, QMessageBox.Yes)

        if reply == QMessageBox.Yes:
            self.upload_features(True)
        else:
            self.upload_features(False)

    def continue_prediction(self):
        self.video_processor.set_continue()

    def upload_features(self, is_tired):
        self.video_processor.set_pause()
        features = self.video_processor.get_last_features()
        self.feature_uploader.setup(features, "vid_id", is_tired, self.upload_features_url, self.user_id)
        self.feature_uploader.start()

    def update_prediction(self, prediction):
        self.label.setText(prediction)  # Обновление текста метки на основе полученного предсказания

    def update_frame(self, frame):
        rgb_image = cv2.cvtColor(frame, cv2.COLOR_BGR2RGB)
        h, w, ch = rgb_image.shape
        bytes_per_line = ch * w
        convert_to_Qt_format = QImage(rgb_image.data, w, h, bytes_per_line, QImage.Format_RGB888)
        p = convert_to_Qt_format.scaled(640, 480, aspectRatioMode=Qt.KeepAspectRatio)
        self.image_label.setPixmap(QPixmap.fromImage(p))
