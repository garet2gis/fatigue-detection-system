import sys
from PyQt5.QtWidgets import QApplication
from windows.login_page import LoginWindow

# Запуск приложения
if __name__ == '__main__':
    app = QApplication(sys.argv)
    ex = LoginWindow()
    ex.show()
    sys.exit(app.exec_())