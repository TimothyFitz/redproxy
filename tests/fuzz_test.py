import sys
import time
import threading
import random
import string


from util import proxied_redis, unproxied_redis
import random

rand_str = lambda length: "".join([random.choice(string.lowercase) for x in range(length)])

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

class FindMisplacedReply(object):
    rounds = 1
    threads = 2
    request_count = 1000

    def thread_main(self):
        value = rand_str(32)
        conn = proxied_redis()
        for x in xrange(self.request_count):
            result = conn.echo(value)
            if result != value:
                raise Exception("actual(%s) != expected(%s)" % (result, value))

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