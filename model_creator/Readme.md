# Сервис создания моделей

Получает из rabbitMQ очереди таски по созданию модели,
внутри таски содержатся id видео(фич выделенных из него), на которых
следует обучить модель XGBoosting.

Данные фичи считываются из БД, после чего обучает модель и отправляет в model storage service. Где данная модель сохраняется.