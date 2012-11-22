import sys
import time
import threading


from util import proxied_redis
import random

class Basic(object):
    rounds = 1
    threads = 100
    request_count = 100

    def thread_main(self):
        conn = proxied_redis()

        actions = [
            lambda: conn.set("foo", "bar"),
            lambda: conn.get("foo"),
            lambda: conn.incr("biz"),
            lambda: conn.decr("biz"),
            lambda: conn.sadd("funk", random.randrange(100))
        ]

        for r in xrange(self.request_count):
            random.choice(actions)()

class C1k(Basic):
    threads = 1024
    request_count = 10

class FDLeak(Basic):
    rounds = 100
    threads = 100
    request_count = 1


def main():
    if len(sys.argv) == 1:
        program = Basic
    else:
        program = globals()[sys.argv[1]]

    for round in range(program.rounds):
        threads = [threading.Thread(target=program().thread_main) for x in range(program.threads)]

        for thread in threads:
            thread.start()

        for thread in threads:
            thread.join()

    print "Success!"

if __name__ == "__main__":
    main()