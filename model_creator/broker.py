import pika
import json
from model_creator import create_xgb
import logging


class Broker:
    def __init__(self, model_storage_url, queue_name, repository, rabbitmq_host='localhost', rabbitmq_port=5672,
                 rabbitmq_user='user',
                 rabbitmq_pass='password'):
        self.queue_name = queue_name
        self.repository = repository
        self.model_storage_url = model_storage_url
        self.rabbitmq_host = rabbitmq_host
        self.rabbitmq_port = rabbitmq_port
        self.rabbitmq_user = rabbitmq_user
        self.rabbitmq_pass = rabbitmq_pass

    def connect(self):
        credentials = pika.PlainCredentials(self.rabbitmq_user, self.rabbitmq_pass)
        parameters = pika.ConnectionParameters(self.rabbitmq_host, self.rabbitmq_port, '/', credentials)
        self.connection = pika.BlockingConnection(parameters)
        self.channel = self.connection.channel()
        self.channel.queue_declare(queue=self.queue_name)

    def callback(self, ch, method, properties, body):
        try:
            msg = json.loads(body)

            user_id = msg['user_id']
            model_type = msg['model_type']

            df = self.repository.get_features_by_user_id(user_id)

            features_count = len(df)

            logging.info(f"Get {features_count} records with user_id: {user_id}")
            create_xgb(df, self.model_storage_url, user_id, model_type, features_count)
            logging.info(f"Create xgb model for user_id: {user_id}")




        except Exception as e:
            logging.error("Error processing message:", e)

    def start_consuming(self):
        self.channel.basic_consume(queue=self.queue_name, on_message_callback=self.callback, auto_ack=True)
        print(' [*] Waiting for messages. To exit press CTRL+C')
        self.channel.start_consuming()

    def close_connection(self):
        self.connection.close()
