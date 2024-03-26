import sqlalchemy as db
import pandas as pd
from contextlib import contextmanager

class Repository:
    def __init__(self, db_uri):
        self.engine = db.create_engine(db_uri)

    @contextmanager
    def connect(self):
        connection = self.engine.connect()
        try:
            yield connection
        finally:
            connection.close()

    def get_user_features(self, ids):

        # Пример метода, использующего соединение с базой данных
        with self.connect() as connection:
            ids = ['\'{}\''.format(id) for id in ids]
            query = f"SELECT * FROM video_features WHERE video_id IN ({', '.join(ids)})"

            records_df = pd.read_sql_query(query, connection)
            return records_df
