from wrapper_sql import WrapperSQL
import mysql.connector

class WrapperMySQL(WrapperSQL):
    def __init__(self, username, password, server):
        print("wrapper mysql")
        self.conn = mysql.connector(user = username, password = password, database = "wt")

    def query(self, sql, params = []):
        cursor = self.conn.cursor()

        return cursor.execute(sql, params);
