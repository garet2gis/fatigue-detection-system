from PyQt5.QtWidgets import QWidget, QLabel, QLineEdit, QPushButton, QVBoxLayout, QMessageBox, QHBoxLayout
from PyQt5.QtGui import QFont, QPalette, QColor
from windows.main_page import MainWindow


def create_label(label):
    label = QLabel(label)
    label.setFont(QFont('Arial', 14))
    label.setStyleSheet('QLabel { color : white; }')
    return label


def create_input():
    input = QLineEdit()
    input.setFont(QFont('Arial', 14))
    input.setStyleSheet('QLineEdit { background-color: white; color: black }')
    return input


class LoginWindow(QWidget):
    def __init__(self):
        super().__init__()
        self.initUI()

    def initUI(self):
        self.setGeometry(300, 300, 400, 250)
        self.setWindowTitle('Вход')

        # Установка палитры для окна
        palette = self.palette()
        palette.setColor(QPalette.Window, QColor(62, 94, 104))
        self.setPalette(palette)

        layout = QVBoxLayout()

        self.user_label = create_label('Имя пользователя:')
        self.user_input = create_input()

        self.pass_label = create_label('Пароль:')
        self.pass_input = create_input()

        self.login_button = QPushButton('Войти')
        self.login_button.setFont(QFont('Arial', 14))  # Увеличенный шрифт
        self.login_button.setStyleSheet('QPushButton { background-color: #5EB850; color: white; }')
        self.login_button.clicked.connect(self.check_credentials)

        layout.addWidget(self.user_label)
        layout.addWidget(self.user_input)
        layout.addWidget(self.pass_label)
        layout.addWidget(self.pass_input)

        button_layout = QHBoxLayout()
        button_layout.addStretch()
        button_layout.addWidget(self.login_button)
        button_layout.addStretch()

        layout.addLayout(button_layout)
        self.setLayout(layout)

    def check_credentials(self):
        username = self.user_input.text()
        password = self.pass_input.text()
        if check_login(username, password):
            self.training_window = MainWindow()
            self.training_window.show()

            self.close()
        else:
            QMessageBox.warning(self, 'Ошибка входа', 'Имя пользователя или пароль неккоректны.', QMessageBox.Ok,
                                QMessageBox.Ok)


def check_login(username, password):
    # TODO: запрос на логин
    if username == "" and password == "":
        return True
    else:
        return False
