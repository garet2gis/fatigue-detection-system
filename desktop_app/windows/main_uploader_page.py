import os
from PyQt5.QtWidgets import QHBoxLayout, QWidget, QVBoxLayout, QPushButton, QMessageBox, QLabel
from PyQt5.QtGui import QFont, QPalette
from PyQt5.QtCore import QTimer
from PyQt5.QtGui import QColor
import subprocess
import shutil
from video_recorder.video_recorder import VideoRecorder
from preprocess.feature_uploader import FeatureUploader


def create_label(label):
    label = QLabel(label)
    label.setFont(QFont('Arial', 14))
    label.setStyleSheet('QLabel { color : white; }')
    return label


class MainDataUploaderWindow(QWidget):
    def __init__(self, model_cfg):
        self.tired_path = os.path.join('videos', 'tired')
        self.awake_path = os.path.join('videos', 'awake')
        self.model_cfg = model_cfg
        super().__init__()
        self.initUI()

    def initUI(self):
        self.setGeometry(300, 300, 300, 300)
        self.setWindowTitle('Сбор данных для модели (видеоданные)')

        # Цвет фона
        palette = self.palette()
        palette.setColor(QPalette.Window, QColor(62, 94, 104))
        self.setPalette(palette)

        self.open_tired_folder_button = QPushButton('Просмотр', self)
        self.open_tired_folder_button.setFont(QFont('Arial', 14))  # Увеличенный шрифт
        self.open_tired_folder_button.setStyleSheet('QPushButton { background-color: #4C4A74; color: white; }')
        self.open_tired_folder_button.clicked.connect(self.open_tired_folder)

        self.open_awake_folder_button = QPushButton('Просмотр', self)
        self.open_awake_folder_button.setFont(QFont('Arial', 14))  # Увеличенный шрифт
        self.open_awake_folder_button.setStyleSheet('QPushButton { background-color: #4C4A74; color: white; }')
        self.open_awake_folder_button.clicked.connect(self.open_awake_folder)

        self.tired_files_count = create_label(f"Количество видео, на которых вы устали: {count_files(self.tired_path)}")
        self.awake_files_count = create_label(f"Количество видео, на которых вы бодры: {count_files(self.awake_path)}")

        self.init_video_capture_button = QPushButton('Записать видео', self)
        self.init_video_capture_button.clicked.connect(self.ask_to_record_video)

        self.layout = QVBoxLayout()

        top_layout = QHBoxLayout()
        bottom_layout = QHBoxLayout()

        top_layout.addWidget(self.tired_files_count)
        top_layout.addWidget(self.open_tired_folder_button)

        bottom_layout.addWidget(self.awake_files_count)
        bottom_layout.addWidget(self.open_awake_folder_button)

        self.layout.addLayout(top_layout)
        self.layout.addLayout(bottom_layout)
        self.layout.addWidget(self.init_video_capture_button)

        self.setLayout(self.layout)

        self.timer = QTimer(self)
        self.timer.timeout.connect(self.ask_to_record_video)
        self.timer.start(60000000)  # Запуск таймера на каждый час (3600000 миллисекунд = 1 час)
        self.video_recorder = VideoRecorder()
        self.feature_uploader = FeatureUploader()
        self.video_recorder.finished.connect(self.ask_if_tired)

    def open_awake_folder(self):
        folder_path = os.path.abspath(self.awake_path)  # Замените на путь к вашей папке
        subprocess.run(['open', folder_path], check=True)

    def open_tired_folder(self):
        folder_path = os.path.abspath(self.tired_path)  # Замените на путь к вашей папке
        subprocess.run(['open', folder_path], check=True)

    def ask_to_record_video(self):
        reply = QMessageBox.question(self, 'Запись видео', 'Хотите ли вы начать запись видео?',
                                     QMessageBox.Yes | QMessageBox.No, QMessageBox.Yes)
        if reply == QMessageBox.Yes:
            self.start_recording()

    def update_count(self):
        self.tired_files_count.setText(f"Количество видео, на которых вы устали: {count_files(self.tired_path)}")
        self.awake_files_count.setText(f"Количество видео, на которых вы бодры: {count_files(self.awake_path)}")

    def start_recording(self):
        if not self.video_recorder.isRunning():
            self.video_recorder.start()

    def ask_if_tired(self, video_id, filename):
        reply = QMessageBox.question(self, 'Состояние', 'Вы устали?',
                                     QMessageBox.Yes | QMessageBox.No, QMessageBox.No)
        user_id = self.model_cfg['content']['user_id']
        upload_features_url = self.model_cfg['content']['face_model']['upload_features_url']

        if reply == QMessageBox.Yes:
            tired_path = './videos/tired'
            shutil.move(filename, tired_path)
            self.feature_uploader.setup(video_id, os.path.join(tired_path, filename), True, upload_features_url,
                                        user_id)

        else:
            awake_path = './videos/awake'
            shutil.move(filename, awake_path)
            self.feature_uploader.setup(video_id, os.path.join(awake_path, filename), False, upload_features_url,
                                        user_id)

        self.update_count()

        if not self.feature_uploader.isRunning():
            self.feature_uploader.start()


def count_files(path):
    return len([name for name in os.listdir(path)
                if os.path.isfile(os.path.join(path, name))])
