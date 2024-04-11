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

    def get_features_by_user_id(self, user_id):
        with self.connect() as connection:
            query = "SELECT * FROM video_features WHERE user_id = %s"
            records_df = pd.read_sql_query(query, connection, params=(user_id,))
            return records_df

    def get_all_features(self):
        with self.connect() as connection:
            query = "SELECT * FROM video_features"
            records_df = pd.read_sql_query(query, connection)
            return records_df
