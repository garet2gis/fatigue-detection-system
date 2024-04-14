import logging
import sys
from PyQt5.QtWidgets import QApplication
from windows.login_page import LoginWindow
from custom_logger.logger import setup_logger
from config.config import LoginConfig

# Запуск приложения
if __name__ == '__main__':
    setup_logger()
    logging.info("starting desktop app")

    app = QApplication(sys.argv)

    app_config = LoginConfig(login_url="http://0.0.0.0:3390/api/v1/auth/login",
                             register_url="http://0.0.0.0:3390/api/v1/auth/register")

    ex = LoginWindow(app_config)


    ex.show()
    sys.exit(app.exec_())
