import pika
import json


class Broker:
    def __init__(self, queue_name, repository, rabbitmq_host='localhost', rabbitmq_port=5672, rabbitmq_user='user',
                 rabbitmq_pass='password'):
        self.queue_name = queue_name
        self.repository = repository
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
            ids = json.loads(body)

            df = self.repository.get_user_features(ids)

            print("Records with ids {}: \n{}".format(ids, df))

        except Exception as e:
            print("Error processing message:", e)

    def start_consuming(self):
        self.channel.basic_consume(queue=self.queue_name, on_message_callback=self.callback, auto_ack=True)
        print(' [*] Waiting for messages. To exit press CTRL+C')
        self.channel.start_consuming()

    def close_connection(self):
        self.connection.close()
