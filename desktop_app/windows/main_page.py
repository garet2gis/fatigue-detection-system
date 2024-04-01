import os
from PyQt5.QtWidgets import QHBoxLayout, QWidget, QVBoxLayout, QPushButton, QMessageBox, QLabel
from PyQt5.QtGui import QFont, QPalette
from PyQt5.QtCore import QTimer
from PyQt5.QtGui import QColor
import subprocess
import time
from vidgear.gears import CamGear, WriteGear
import shutil
from datetime import datetime
import uuid

from preprocess.preprocess_video_to_csv import upload_features_from_video


def create_label(label):
    label = QLabel(label)
    label.setFont(QFont('Arial', 14))
    label.setStyleSheet('QLabel { color : white; }')
    return label


class MainWindow(QWidget):
    def __init__(self, upload_csv_url):
        self.tired_path = os.path.join('videos', 'tired')
        self.awake_path = os.path.join('videos', 'awake')
        self.upload_csv_url = upload_csv_url
        super().__init__()
        self.initUI()

    def initUI(self):
        self.setGeometry(300, 300, 300, 300)
        self.setWindowTitle('Обучение модели (видеоданные)')

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
        self.timer.start(60000)  # Запуск таймера на каждый час (3600000 миллисекунд = 1 час)

    def open_awake_folder(self):
        folder_path = os.path.abspath(self.awake_path)  # Замените на путь к вашей папке
        subprocess.run(['open', folder_path], check=True)

    def open_tired_folder(self):
        folder_path = os.path.abspath(self.tired_path)  # Замените на путь к вашей папке
        subprocess.run(['open', folder_path], check=True)

    def ask_to_record_video(self):
        reply = QMessageBox.question(self, 'Запись видео', 'Хотите ли вы начать запись видео?',
                                     QMessageBox.Yes | QMessageBox.No, QMessageBox.No)
        if reply == QMessageBox.Yes:
            self.record_video()
            self.update_count()

    def update_count(self):
        self.tired_files_count.setText(f"Количество видео, на которых вы устали: {count_files(self.tired_path)}")
        self.awake_files_count.setText(f"Количество видео, на которых вы бодры: {count_files(self.awake_path)}")

    def record_video(self):
        options = {
            "CAP_PROP_FRAME_WIDTH": 244,
            "CAP_PROP_FRAME_HEIGHT": 244,
            "CAP_PROP_FPS": 244,
        }

        stream = CamGear(source=1, **options).start()

        video_id = str(uuid.uuid4())
        filename = f"{video_id}.mp4"

        writer = WriteGear(output=filename)

        start = time.perf_counter()
        while True:
            frame = stream.read()

            if frame is None:
                break

            writer.write(frame)

            stop = time.perf_counter()
            # TODO const 15 seconds
            if stop - start > 15:
                break

        stream.stop()
        writer.close()

        self.ask_if_tired(video_id, filename)

    def ask_if_tired(self, video_id, filename):
        reply = QMessageBox.question(self, 'Состояние', 'Вы устали?',
                                     QMessageBox.Yes | QMessageBox.No, QMessageBox.No)
        if reply == QMessageBox.Yes:
            tired_path = './videos/tired'
            shutil.move(filename, tired_path)
            upload_features_from_video(video_id, os.path.join(tired_path, filename), True, self.upload_csv_url)

        else:
            awake_path = './videos/awake'
            shutil.move(filename, awake_path)
            upload_features_from_video(video_id, os.path.join(awake_path, filename), True, self.upload_csv_url)


def count_files(path):
    return len([name for name in os.listdir(path)
                if os.path.isfile(os.path.join(path, name))])
