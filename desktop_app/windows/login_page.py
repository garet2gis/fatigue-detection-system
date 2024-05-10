from PyQt5.QtWidgets import QWidget, QLineEdit, QPushButton, QVBoxLayout, QMessageBox
from api import http
from config import config
from windows.main_uploader_page import MainDataUploaderWindow
from windows.predictor_page import PredictorWindow


class LoginWindow(QWidget):
    def __init__(self, cfg):
        super().__init__()
        self.setWindowTitle('Приложение Авторизации')
        self.setGeometry(100, 100, 280, 150)  # x, y, width, height
        self.initUI()
        self.login_url = cfg.login_url
        self.register_url = cfg.register_url

    def initUI(self):
        self.central_widget = QWidget()
        # self.setCentralWidget(self.central_widget)

        self.layout = QVBoxLayout()
        self.username = QLineEdit(self, placeholderText='Логин')
        self.password = QLineEdit(self, placeholderText='Пароль')
        self.password.setEchoMode(QLineEdit.Password)  # Скрыть ввод пароля
        self.login_button = QPushButton('Войти', self)
        self.register_button = QPushButton('Регистрация', self)

        self.layout.addWidget(self.username)
        self.layout.addWidget(self.password)
        self.layout.addWidget(self.login_button)
        self.layout.addWidget(self.register_button)

        self.central_widget.setLayout(self.layout)

        self.login_button.clicked.connect(self.login)
        self.register_button.clicked.connect(self.register)

        self.setLayout(self.layout)

    def login(self):
        username = self.username.text()
        password = self.password.text()
        success, message, data = http.perform_login(self.login_url, username, password)
        if success:
            if data['content']['model_urls']['face_model'] == "":
                self.main_window = MainDataUploaderWindow(data['content'])
            else:
                self.main_window = PredictorWindow(data['content'])
            self.main_window.show()
            self.close()
        QMessageBox.information(self, 'Информация', message)

    def register(self):
        username = self.username.text()
        password = self.password.text()
        success, message = http.perform_register(self.register_url, username, password)
        QMessageBox.information(self, 'Информация', message)