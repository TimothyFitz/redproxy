import signal
import subprocess
import unittest
import os
import time

import redis.client

def unproxied_redis():
    return redis.client.StrictRedis()

def proxied_redis():
    return redis.client.StrictRedis("localhost", 9999)


class TestProxy(unittest.TestCase):
    def setUp(self):
        self.client = redis.client.StrictRedis("localhost", 9999)

    def test_foo(self):
        self.assertIsNone(self.client.get("not_found"))

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
