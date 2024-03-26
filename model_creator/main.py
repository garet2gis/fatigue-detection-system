from broker import Broker
from repository import Repository
import os

RABBITMQ_HOST = os.environ.get('RABBITMQ_HOST')
RABBITMQ_QUEUE = os.environ.get('RABBITMQ_QUEUE')

POSTGRESQL_HOST = os.environ.get('DB_HOST')
POSTGRESQL_DBNAME = os.environ.get('DB_NAME')
POSTGRESQL_USER = os.environ.get('DB_USERNAME')
POSTGRESQL_PASSWORD = os.environ.get('DB_PASSWORD')
POSTGRESQL_PORT = os.environ.get('DB_PORT')

if __name__ == '__main__':
    engine_str = f'postgresql://{POSTGRESQL_USER}:{POSTGRESQL_PASSWORD}@{POSTGRESQL_HOST}:{POSTGRESQL_PORT}/{POSTGRESQL_DBNAME}'
    repository = Repository(engine_str)

    consumer = Broker(queue_name=RABBITMQ_QUEUE, repository=repository)

    consumer.connect()
    consumer.start_consuming()
