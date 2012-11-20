import time
import threading

from util import proxied_redis
import random

THREADS = 100
REQUEST_COUNT = 100


def thread_main():
    conn = proxied_redis()

    actions = [
        lambda: conn.set("foo", "bar"),
        lambda: conn.get("foo"),
        lambda: conn.incr("biz"),
        lambda: conn.decr("biz"),
        lambda: conn.sadd("funk", random.randrange(100))
    ]

    for r in xrange(REQUEST_COUNT):
        random.choice(actions)()

def main():
    threads = [threading.Thread(target=thread_main) for x in range(THREADS)]

    for thread in threads:
        thread.start()

    for thread in threads:
        thread.join()

    print "Success!"

if __name__ == "__main__":
    main()