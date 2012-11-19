import signal
import subprocess
import unittest
import os
import time

import redis.client
import redis.connection
import redis.exceptions

def unproxied_redis():
    return redis.client.StrictRedis(port=6379)

def proxied_redis():
    return redis.client.StrictRedis(port=9999)


class TestLowLevelConnectionBehavior(unittest.TestCase):
    def test_quit_command_causes_connection_to_be_closed(self):
        conn = redis.connection.Connection(port=9999, socket_timeout=3)
        conn.connect()
        conn.send_command('QUIT')
        self.assertEqual("OK", conn.read_response())

        with self.assertRaisesRegexp(redis.exceptions.ConnectionError, "Socket closed on remote end"):
            conn.read_response()

    def test_closing_connection_closes_other_side(self):
        ARBITRARY_UNUSED_DB = '9'
        rc = unproxied_redis()

        def is_connected():
            clients = rc.client_list()
            for client in clients:
                if client['db'] == ARBITRARY_UNUSED_DB:
                    return True
            return False

        self.assertFalse(is_connected())

        conn = redis.connection.Connection(port=9999, socket_timeout=3)
        conn.connect()
        conn.send_command("SELECT", ARBITRARY_UNUSED_DB)
        conn.read_response()

        wait_for_true(is_connected, comment="connection")

        conn.disconnect()

        wait_for_true(lambda: not is_connected(), comment="disconnect")



class TestProxy(unittest.TestCase):
    def setUp(self):
        self.client = proxied_redis()

    def test_get_returns_none_if_no_value_set(self):
        self.assertIsNone(self.client.get("not_found"))

    def test_ping(self):
        self.assertTrue(self.client.ping())

    def test_set(self):
        self.assertTrue(self.client.set("foo", "bar"))

    def test_set_and_get(self):
        self.client.set("timothy", "fitz")
        self.assertEqual("fitz", self.client.get("timothy"))

    def test_error_response(self):
        self.client.set("foo", "bar")
        with self.assertRaisesRegexp(redis.exceptions.ResponseError, "Operation against a key holding the wrong kind of value"):
            self.client.rpop("foo")

    def test_integer_response(self):
        self.client.set("counter", 0)
        self.assertEqual(1, self.client.incr("counter"))

def wait_for_true(fun, comment=None):
    start = time.time()
    while time.time() - start < 3:
        if fun():
            return
        time.sleep(0.1)

    message = "Timed out waiting for %s" % comment if comment else "Timed out."
    raise Exception(message)


def wait_for_redis(which, conn_factory):
    c = redis.client.StrictRedis()
    start = time.time()
    while time.time() - start < 3:
        try:
            c.ping()
        except:
            pass
        else:
            return
    raise Exception("Timed out waiting for %s to start")

if __name__ == "__main__":
    process = subprocess.Popen(["redis-server", "test.conf"])
    wait_for_redis("redis", unproxied_redis)
    wait_for_redis("proxy", proxied_redis)
    try:
        unittest.main()
    finally:
        os.kill(process.pid, signal.SIGQUIT)
